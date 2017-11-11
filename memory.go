package cache

import (
	"sync"
	"time"
)

const inMemoryStoreName = "In-Memory Store"

// NewInMemoryStore creates a new in-memory store
func NewInMemoryStore() Store {
	return &inmemorystore{
		cache:  make(map[string]cacheddata),
		config: StoreConfig{StoreName: inMemoryStoreName},
	}
}

type inmemorystore struct {
	sync.RWMutex
	cache  map[string]cacheddata
	config StoreConfig
}

func (ims *inmemorystore) Store(cacheKey string, data []byte) error {
	ims.Lock()
	ims.cache[cacheKey] = cacheddata{data, time.Now()}
	ims.Unlock()
	return nil
}

func (ims *inmemorystore) Retrieve(cacheKey string) (*cacheddata, bool, error) {
	ims.RLock()
	cachedData, ok := ims.cache[cacheKey]
	ims.RUnlock()
	return &cachedData, ok, nil
}

func (ims *inmemorystore) Delete(cacheKey string) error {
	ims.Lock()
	delete(ims.cache, cacheKey)
	ims.Unlock()
	return nil
}

func (ims *inmemorystore) Config() StoreConfig {
	return ims.config
}

func (ims *inmemorystore) SetConfig(config StoreConfig) {
	ims.config = config
}
