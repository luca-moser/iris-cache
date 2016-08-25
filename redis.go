package cache

import (
	"encoding/json"
	"time"

	"gopkg.in/redis.v3"
)

const redisStoreName = "Redis Store"

// Creates a new redis store using the given client
func NewRedisStore(client *redis.Client) *redisstore {
	return &redisstore{
		client: client,
		config: CacheStoreConfig{StoreName: redisStoreName}}
}

type redisstore struct {
	client *redis.Client
	config CacheStoreConfig
}

func (rs *redisstore) Store(cacheKey string, data []byte) error {
	bytes, err := json.Marshal(cacheddata{data, time.Now()})
	if err != nil {
		return err
	}
	return rs.client.Set(cacheKey, bytes, 0).Err()
}

func (rs *redisstore) Retrieve(cacheKey string) (*cacheddata, bool, error) {
	cachedBytes, err := rs.client.Get(cacheKey).Bytes()
	if err != nil && err != redis.Nil {
		return nil, false, err
	}
	if err == redis.Nil {
		return nil, false, nil
	}
	cachedData := &cacheddata{}
	if err := json.Unmarshal(cachedBytes, cachedData); err != nil {
		return nil, false, err
	}
	return cachedData, true, nil
}

func (rs *redisstore) Delete(cacheKey string) error {
	_, err := rs.client.Del(cacheKey).Result()
	return err
}

func (rs *redisstore) Config() CacheStoreConfig {
	return rs.config
}

func (rs *redisstore) SetConfig(config CacheStoreConfig) {
	rs.config = config
}
