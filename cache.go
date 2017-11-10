package cache

import (
	"time"

	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

const (
	// Content types used when determining wether to cache the response body

	// ContentTypeJSON header value for JSON data.
	ContentTypeJSON = context.ContentJSONHeaderValue
	// ContentTypeHTML is the  string of text/html response header's content type value.
	ContentTypeHTML = context.ContentHTMLHeaderValue
	// CtxIrisCacheKey the key used in iris.Context.Values().Set(), which holds the given route's cache key
	CtxIrisCacheKey = "iris_cache_key"
	// CtxIrisSkipCacheKey indicator which a previous handler func can use, to explictly skip the cache middleware
	CtxIrisSkipCacheKey = "iris_skip_cache"
)

// Store defines a store which can cache data
type Store interface {
	// Stores the data under the given cache key
	Store(string, []byte) error
	// Retrieves the data with the given cache key
	Retrieve(string) (*cacheddata, bool, error)
	// Deletes the data under the given cache key
	Delete(string) error
	// Returns the config for the store
	Config() StoreConfig
	// Replaces the current config with the given config
	SetConfig(config StoreConfig)
}

// KeySupplier defines a function which computes a cache key (string) out of a iris.Context
type KeySupplier func(ctx iris.Context) string

// StoreConfig defines configuration options which are used by the store
type StoreConfig struct {
	// The name of the used store
	StoreName string
}

// Config defines configuration options which are used by the middleware
type Config struct {
	// Wether to automatically remove the data after the cache duration
	// or let the middleware only remove it, when a route is called and
	// the cache duration is expired.
	AutoRemove bool
	// Wether iris' gzip compression is used
	IrisGzipEnabled bool
	// The amount of time data is cached in the store
	CacheTimeDuration time.Duration
	// The content type which will be cached
	ContentType string
	// Function which computes the cache key out of an iris.Context
	CacheKeyFunc KeySupplier
}

// primitive type representing cached data
type cacheddata struct {
	Data      []byte    `json:"data" bson:"data"`
	CreatedOn time.Time `json:"created_on" bson:"created_on"`
}

// NewCacheHF creates a cache handler function with the given cache duration per route.
func NewCacheHF(config Config, store Store) iris.Handler {
	return NewCache(config, store).Serve
}

// NewCache creates a new caching middleware with the given cache duration per route
func NewCache(config Config, store Store) *Cache {
	c := &Cache{config: config, store: store}
	if c.config.CacheKeyFunc == nil {
		c.config.CacheKeyFunc = RequestPathToMD5
	}
	return c
}

// Cache contains the store and the configuration, its `Serve` function is the middleware.
type Cache struct {
	store  Store
	config Config
}

// Serve handles the cache action, should be registered before the main handler.
func (c *Cache) Serve(ctx iris.Context) {
	switch t := ctx.Values().Get(CtxIrisSkipCacheKey).(type) {
	case bool:
		if t {
			return
		}
	}
	cacheKey := c.config.CacheKeyFunc(ctx)

	cachedJSON, ok, err := c.store.Retrieve(cacheKey)
	if err != nil {
		panic(err)
	}
	if ok && time.Now().Before(cachedJSON.CreatedOn.Add(c.config.CacheTimeDuration)) {
		ctx.ContentType(c.config.ContentType)
		if c.config.IrisGzipEnabled && ctx.ClientSupportsGzip() {
			ctx.Header("Content-Encoding", "gzip")
		}
		ctx.StatusCode(iris.StatusOK)
		ctx.Write(cachedJSON.Data)
		return
	}

	// call other routes
	ctx.Values().Set(CtxIrisCacheKey, cacheKey)
	rec := ctx.Recorder()
	ctx.Next()

	// check content type
	contentType := ctx.GetContentType()
	if contentType != c.config.ContentType {
		return
	}

	// get computed response
	bytesToCache := rec.Body()
	fresh := make([]byte, len(bytesToCache))
	copy(fresh, bytesToCache)

	c.store.Store(cacheKey, fresh)

	// skip auto remove. remove it when the duration is expired and the route is executed
	if !c.config.AutoRemove {
		return
	}

	// racy
	go func() {
		// remove cached data after defined duration
		<-time.After(c.config.CacheTimeDuration)

		// retrieve cached data as it might be refreshed
		cachedData, ok, err := c.store.Retrieve(cacheKey)
		if err != nil {
			panic(err)
		}
		if !ok || time.Now().Before(cachedData.CreatedOn.Add(c.config.CacheTimeDuration)) {
			return
		}
		c.store.Delete(cacheKey)
	}()
}

// Invalidate invalidates the cached data by the given key
func (c *Cache) Invalidate(cacheKey string) {
	c.store.Delete(cacheKey)
}
