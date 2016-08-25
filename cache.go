package cache

import (
	"time"

	"github.com/kataras/iris"
)

const (
	// Content types used when determining wether to cache the response body
	ContentTypeJSON = "application/json; charset=UTF-8"
	ContentTypeHTML = "text/html; charset=UTF-8"
	// The key used in iris.Context.Set(), which holds the given route's cache key
	CtxIrisCacheKey = "iris_cache_key"
	// Indicator which a previous handler func can use, to explictly skip the cache middleware
	CtxIrisSkipCacheKey = "iris_skip_cache"
)

// Defines a store which can cache data
type CacheStore interface {
	// Stores the data under the given cache key
	Store(string, []byte) error
	// Retrieves the data with the given cache key
	Retrieve(string) (*cacheddata, bool, error)
	// Deletes the data under the given cache key
	Delete(string) error
	// Returns the config for the store
	Config() CacheStoreConfig
	// Replaces the current config with the given config
	SetConfig(config CacheStoreConfig)
}

// Defines a function which computes a cache key (string) out of a iris.Context
type CacheKeySupplier func(ctx *iris.Context) string

// Defines configuration options which are used by the store
type CacheStoreConfig struct {
	// The name of the used store
	StoreName string
}

// Defines configuration options which are used by the middleware
type CacheConfig struct {
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
	CacheKeyFunc CacheKeySupplier
}

// primitive type representing cached data
type cacheddata struct {
	Data      []byte    `json:"data" bson:"data"`
	CreatedOn time.Time `json:"created_on" bson:"created_on"`
}

// Creates a cache handler function with the given cache duration per route.
func NewCacheHF(config CacheConfig, store CacheStore) iris.HandlerFunc {
	return NewCache(config, store).Serve
}

// Creates a new caching middleware with the given cache duration per route
func NewCache(config CacheConfig, store CacheStore) *cache {
	c := &cache{config: config, store: store}
	if c.config.CacheKeyFunc == nil {
		c.config.CacheKeyFunc = RequestPathToMD5
	}
	return c
}

type cache struct {
	store  CacheStore
	config CacheConfig
}

func (c *cache) Serve(ctx *iris.Context) {
	switch t := ctx.Get(CtxIrisSkipCacheKey).(type) {
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
		ctx.SetContentType(c.config.ContentType)
		if c.config.IrisGzipEnabled {
			ctx.Response.Header.Set("Content-Encoding", "gzip")
		}
		ctx.SetStatusCode(iris.StatusOK)
		ctx.SetBody(cachedJSON.Data)
		return
	}

	// call other routes
	ctx.Set(CtxIrisCacheKey, cacheKey)
	ctx.Next()

	// check content type
	contentType := ctx.Response.Header.ContentType()
	if string(contentType) != c.config.ContentType {
		return
	}

	// get computed response
	bytesToCache := ctx.Response.Body()
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

// Invalidates the cached data by the given key
func (c *cache) Invalidate(cacheKey string) {
	c.store.Delete(cacheKey)
}
