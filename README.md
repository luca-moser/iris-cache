## Caching middleware for the [iris](https://github.com/kataras/iris) web framework ![](https://travis-ci.org/luca-moser/iris-cache.svg)

This middleware automatically caches the computed response (per route) for the given duration.
A store can be defined which handles the storing and retrieving of cached data. 
The route is not executed when the response is cached as the result is returned immediately by the middleware.

Stores:
* implemented:
    * In-memory
    * Redis
* planned:
    * Flatfile

### Usage

```go
// cache the data in-memory
inMemoryStore := cache.NewInMemoryStore()

// middleware configuration
cacheConfig := cache.CacheConfig{
    // automatically remove the cached data after the given cache time duration.
    // if AutoRemove is to false, the cached data will be removed/refreshed up on the first route
    // call after the cache time duration expired.
    AutoRemove:        false,
    // the expire time of the cached data
    CacheTimeDuration: time.Duration(5) * time.Minute,
    // defines the content type which will be cached
    ContentType:       cache.ContentTypeJSON, // cache json responses
    // (!) important: must be set to true if iris' gzip is enabled
    IrisGzipEnabled:   false,
    // you can supply your own cache key generator by implementing `cache.CacheKeySupplier`
    CacheKeyFunc: cache.RequestPathToMD5 // default if non supplied
}

// as global middleware (iris.Handler), caching every response with a JSON content type
iris.Use(cache.NewCache(cacheConfig, inMemoryStore))

// OR

// as iris.HandlerFunc for specific routes
iris.Get("/json", cache.NewCacheHF(cacheConfig, inMemoryStore), func(c *iris.Context) {
    someStructWithJSONTags := something()
    c.JSON(http.StatusOK, someStructWithJSONTags) // gets cached automatically
})
```

```go
// the cache key which will be used can be obtained from the current iris.Context
cacheKey := ctx.Get(cache.CtxIrisCacheKey).(string)

// the cached data can be invalidated by using the obtained cache key
cache.Invalidate(cacheKey)
```

### Licence

```
MIT License

Copyright (c) 2016 Luca Moser

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```