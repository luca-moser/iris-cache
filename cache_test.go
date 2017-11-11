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
	// you could use that library now to do http testing: "github.com/kataras/iris/httptest"
)

const (
	irisSrvWithMemoryStore = "127.0.0.1:1234"
	irisSrvWithRedisStore  = "127.0.0.1:1235"
	redisService           = "127.0.0.1:6379"
	sleepTime              = time.Duration(2) * time.Second
)

var cacheConfig = Config{
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
	c := NewCacheHF(cacheConfig, NewInMemoryStore())
	engine.Use(c)
	engine.Get("/json", func(ctx iris.Context) {
		<-time.After(sleepTime)
		ctx.JSON(dummy{"test"})
	})

	go engine.Run(iris.Addr(irisSrvWithMemoryStore), iris.WithoutStartupLog, iris.WithoutVersionChecker, iris.WithoutServerError(iris.ErrServerClosed))
	<-time.After(time.Duration(1350 * time.Millisecond))
}

func setupRedisStoreIrisSrv() {
	engine := iris.New()

	redisClient := redis.NewClient(&redis.Options{Addr: redisService})
	if err := redisClient.Ping().Err(); err != nil {
		panic(err)
	}
	redisClient.FlushDb()
	c := NewCacheHF(cacheConfig, NewRedisStore(redisClient))
	engine.Use(c)
	engine.Get("/json", func(ctx iris.Context) {
		<-time.After(sleepTime)
		ctx.JSON(dummy{"test"})
	})

	go engine.Run(iris.Addr(irisSrvWithRedisStore), iris.WithoutStartupLog, iris.WithoutVersionChecker, iris.WithoutServerError(iris.ErrServerClosed))
	<-time.After(time.Duration(1350 * time.Millisecond))
}

func TestMain(m *testing.M) {

	// spawn web servers
	setupMemoryStoreIrisSrv()
	setupRedisStoreIrisSrv()

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
	if time.Now().Sub(s) > sleepTime+200*time.Millisecond {
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
	// check if the request was slower than the sleep time,
	// these numbers are not always percise, especially in services like travis, so give it time.
	if time.Now().Sub(s) > sleepTime+200*time.Millisecond {
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
