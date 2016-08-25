package cache

import (
	"sync"
	"time"
)

const inMemoryStoreName = "In-Memory Store"

// Creates a new in-memory store
func NewInMemoryStore() *inmemorystore {
	return &inmemorystore{
		cache:  make(map[string]cacheddata),
		config: CacheStoreConfig{StoreName: inMemoryStoreName}}
}

type inmemorystore struct {
	sync.Mutex
	cache  map[string]cacheddata
	config CacheStoreConfig
}

func (ims *inmemorystore) Store(cacheKey string, data []byte) error {
	ims.Lock()
	ims.cache[cacheKey] = cacheddata{data, time.Now()}
	ims.Unlock()
	return nil
}

func (ims *inmemorystore) Retrieve(cacheKey string) (*cacheddata, bool, error) {
	ims.Lock()
	cachedData, ok := ims.cache[cacheKey]
	ims.Unlock()
	return &cachedData, ok, nil
}

func (ims *inmemorystore) Delete(cacheKey string) error {
	ims.Lock()
	delete(ims.cache, cacheKey)
	ims.Unlock()
	return nil
}

func (ims *inmemorystore) Config() CacheStoreConfig {
	return ims.config
}

func (ims *inmemorystore) SetConfig(config CacheStoreConfig) {
	ims.config = config
}
