## Caching middleware for the iris web framework

Implemented:
* In-memory JSON caching

Todo:
* Cache invalidation
* Stores:
    * Redis
    * Flatfile

### JSON Cache:
This middleware automatically caches the computed JSON response (per route) for the given duration.
The route is not executed when the response is cached as the result is returned immediately by the middleware.

```
// as global middleware, caching every JSON response
iris.Use(cache.NewJSONCache(time.Duration(5) * time.Minute))

// or for specific routes
iris.Get("/json", cache.CacheJSON(time.Duration(5) * time.Minute)), func(c *iris.Context) {
    structWithJSONTags := something()
    c.JSON(http.StatusOK, structWithJSONTags) // gets cached automatically
})
```