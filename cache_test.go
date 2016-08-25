package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	redis "gopkg.in/redis.v3"

	"github.com/kataras/iris"
)

const (
	irisSrvWithMemoryStore = "127.0.0.1:1234"
	irisSrvWithRedisStore  = "127.0.0.1:1235"
	redisService           = "127.0.0.1:6379"
	sleepTime              = time.Duration(2) * time.Second
)

var cacheConfig = CacheConfig{
	AutoRemove:        false,
	CacheTimeDuration: time.Duration(5) * time.Minute,
	ContentType:       ContentTypeJSON,
	IrisGzipEnabled:   false,
}

type dummy struct {
	Name string `json:"name"`
}

func setupMemoryStoreIrisSrv() {
	engine := iris.New()
	engine.Config.DisableBanner = true
	engine.Use(NewCache(cacheConfig, NewInMemoryStore()))
	engine.Get("/json", func(c *iris.Context) {
		<-time.After(sleepTime)
		c.JSON(iris.StatusOK, dummy{"test"})
	})

	go engine.Listen(irisSrvWithMemoryStore)
}

func setupRedisStoreIrisSrv() {
	engine := iris.New()
	engine.Config.DisableBanner = true
	redisClient := redis.NewClient(&redis.Options{Addr: redisService})
	if err := redisClient.Ping().Err(); err != nil {
		panic(err)
	}
	redisClient.FlushDb()
	engine.Use(NewCache(cacheConfig, NewRedisStore(redisClient)))
	engine.Get("/json", func(c *iris.Context) {
		<-time.After(sleepTime)
		c.JSON(iris.StatusOK, dummy{"test"})
	})

	go engine.Listen(irisSrvWithRedisStore)
}

func TestMain(m *testing.M) {

	// spawn web servers
	setupMemoryStoreIrisSrv()
	<-time.After(time.Duration(1) * time.Second)
	setupRedisStoreIrisSrv()
	<-time.After(time.Duration(1) * time.Second)

	// run the tests
	os.Exit(m.Run())
}

func TestNonCachedMemoryStoreRequest(t *testing.T) {
	if err := doRequest(irisSrvWithMemoryStore); err != nil {
		t.Fatal(err)
		return
	}
}

func TestCachedMemoryStoreRequest(t *testing.T) {
	s := time.Now()
	if err := doRequest(irisSrvWithMemoryStore); err != nil {
		t.Fatal(err)
		return
	}
	// check if the request was slower than the sleep time
	if time.Now().Sub(s) > sleepTime {
		t.Fatal("cached request was slower than non cached request")
		return
	}
}

func TestNonCachedRedisStoreRequest(t *testing.T) {
	if err := doRequest(irisSrvWithRedisStore); err != nil {
		t.Fatal(err)
		return
	}
}

func TestCachedRedisStoreRequest(t *testing.T) {
	s := time.Now()
	if err := doRequest(irisSrvWithRedisStore); err != nil {
		t.Fatal(err)
		return
	}
	// check if the request was slower than the sleep time
	if time.Now().Sub(s) > sleepTime {
		t.Fatal("cached request was slower than non cached request")
		return
	}
}

func BenchmarkCachedMemoryStoreResponse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		doRequestWithoutParsing(irisSrvWithMemoryStore)
	}
}

func BenchmarkCachedRedisStoreResponse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		doRequestWithoutParsing(irisSrvWithRedisStore)
	}
}

func doRequest(url string) error {
	res, err := http.Get(fmt.Sprintf("http://%s/json", url))
	if err != nil {
		return err
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	obj := &dummy{}
	err = json.Unmarshal(buf, obj)
	if err != nil {
		return err
	}

	if obj.Name != "test" {
		return errors.New("name doesn't match origin name")
	}

	return nil
}

func doRequestWithoutParsing(url string) error {
	_, err := http.Get(fmt.Sprintf("http://%s/json", url))
	return err
}
