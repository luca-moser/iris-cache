package cache

import (
	"crypto/md5"
	"sync"
	"time"

	"github.com/kataras/iris"
)

const contentTypeJson = "application/json; charset=UTF-8"

func CacheJSON(cacheDuration time.Duration) iris.HandlerFunc {
	mu := sync.Mutex{}
	cache := make(map[[16]byte]cachedjson)

	return func(ctx *iris.Context) {
		cacheKey := md5.Sum(ctx.Request.URI().Path())

		mu.Lock()
		cachedJSON, ok := cache[cacheKey]
		if ok && time.Now().Before(cachedJSON.when.Add(cacheDuration)) {
			ctx.SetContentType(contentTypeJson)
			ctx.SetStatusCode(iris.StatusOK)
			ctx.SetBody(cachedJSON.data)
			mu.Unlock()
			return
		}
		mu.Unlock()

		// call other routes
		ctx.Next()

		// check content type
		contentType := ctx.Response.Header.ContentType()
		if string(contentType) != contentTypeJson {
			return
		}

		// get json computed json response
		bytesToCache := ctx.Response.Body()
		fresh := make([]byte, len(bytesToCache))
		copy(fresh, bytesToCache)

		go func() {
			// cache data
			mu.Lock()
			cache[cacheKey] = cachedjson{fresh, time.Now()}
			mu.Unlock()

			// remove cached data after defined duration
			<-time.After(cacheDuration)
			mu.Lock()
			_, ok := cache[cacheKey]
			if ok {
				delete(cache, cacheKey)
			}
			mu.Unlock()
		}()
	}
}

func NewJSONCache(cacheDuration time.Duration) *jsoncache {
	return &jsoncache{cacheDuration: cacheDuration, cache: make(map[[16]byte]cachedjson)}
}

type jsoncache struct {
	mu            sync.Mutex
	cacheDuration time.Duration
	cache         map[[16]byte]cachedjson
}

type cachedjson struct {
	data []byte
	when time.Time
}

func (jc *jsoncache) Serve(ctx *iris.Context) {
	cacheKey := md5.Sum(ctx.Request.URI().Path())

	jc.mu.Lock()
	cachedJSON, ok := jc.cache[cacheKey]
	if ok && time.Now().Before(cachedJSON.when.Add(jc.cacheDuration)) {
		ctx.SetContentType(contentTypeJson)
		ctx.SetStatusCode(iris.StatusOK)
		ctx.SetBody(cachedJSON.data)
		jc.mu.Unlock()
		return
	}
	jc.mu.Unlock()

	// call other routes
	ctx.Next()

	// check content type
	contentType := ctx.Response.Header.ContentType()
	if string(contentType) != contentTypeJson {
		return
	}

	// get json computed json response
	bytesToCache := ctx.Response.Body()
	fresh := make([]byte, len(bytesToCache))
	copy(fresh, bytesToCache)

	go func() {
		// cache data
		jc.mu.Lock()
		jc.cache[cacheKey] = cachedjson{fresh, time.Now()}
		jc.mu.Unlock()

		// remove cached data after defined duration
		<-time.After(jc.cacheDuration)
		jc.mu.Lock()
		_, ok := jc.cache[cacheKey]
		if ok {
			delete(jc.cache, cacheKey)
		}
		jc.mu.Unlock()
	}()
}
