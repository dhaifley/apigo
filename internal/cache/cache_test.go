package cache_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dhaifley/apigo/internal/cache"
	"github.com/dhaifley/apigo/internal/config"
	"github.com/google/gomemcache/memcache"
	"github.com/redis/go-redis/v9"
)

type mockMemcacheClient struct{}

func (m *mockMemcacheClient) Get(key string) (*memcache.Item, error) {
	switch key {
	case "test":
		return &memcache.Item{
			Key:   "test",
			Value: []byte("test"),
		}, nil
	default:
		return nil, memcache.ErrCacheMiss
	}
}

func (m *mockMemcacheClient) GetMulti(keys []string,
) (map[string]*memcache.Item, error) {
	if len(keys) < 1 {
		return nil, memcache.ErrCacheMiss
	}

	switch keys[0] {
	case "test":
		return map[string]*memcache.Item{
			"test": {
				Key:   "test",
				Value: []byte("test"),
			},
		}, nil
	default:
		return nil, memcache.ErrCacheMiss
	}
}

func (m *mockMemcacheClient) Set(item *memcache.Item) error {
	return nil
}

func (m *mockMemcacheClient) Delete(key string) error {
	return nil
}

type mockRedisClient struct{}

func (m *mockRedisClient) Get(ctx context.Context,
	key string,
) *redis.StringCmd {
	switch key {
	case "test":
		return redis.NewStringResult("test", nil)
	default:
		return redis.NewStringResult("", redis.Nil)
	}
}

func (m *mockRedisClient) MGet(ctx context.Context,
	keys ...string,
) *redis.SliceCmd {
	if len(keys) < 1 {
		return redis.NewSliceResult(nil, redis.Nil)
	}

	switch keys[0] {
	case "test":
		return redis.NewSliceResult([]any{"test"}, nil)
	default:
		return redis.NewSliceResult([]any{nil}, nil)
	}
}

func (m *mockRedisClient) Set(ctx context.Context,
	key string, value any,
	expiration time.Duration,
) *redis.StatusCmd {
	return redis.NewStatusResult(fmt.Sprintf("%v", value), nil)
}

func (m *mockRedisClient) Del(ctx context.Context,
	keys ...string,
) *redis.IntCmd {
	return redis.NewIntResult(int64(len(keys)), nil)
}

func TestClient(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}

	cfg.SetCache(&config.CacheConfig{
		Type:       cache.CacheTypeMemcache,
		Servers:    []string{"localhost:11211", "localhost:11212"},
		Discovery:  false,
		Expiration: time.Second,
	})

	mp := cache.NewClient(cfg, nil, nil, nil)
	if mp == nil {
		t.Fatal("Unable to initialize memcache client")
	}

	mp.SetMemcacheClient(&mockMemcacheClient{})

	res, err := mp.Get(context.Background(), "test")
	if err != nil {
		t.Errorf("Unexpected error from get: %v", err.Error())
	}

	if string(res.Value) != "test" {
		t.Errorf("Expected value: test, got: %v", res.Value)
	}

	resM, err := mp.GetMulti(context.Background(), "test")
	if err != nil {
		t.Errorf("Unexpected error from get: %v", err.Error())
	}

	if string(resM["test"].Value) != "test" {
		t.Errorf("Expected multi value: test, got: %v", resM)
	}

	err = mp.Set(context.Background(),
		&cache.Item{Key: "test", Value: []byte("test")})
	if err != nil {
		t.Errorf("Unexpected error from set: %v", err.Error())
	}

	_, err = mp.Get(context.Background(), "invalid")
	if err == nil {
		t.Error("Expected cache miss error, got: nil")
	}

	err = mp.Delete(context.Background(), "test")
	if err != nil {
		t.Errorf("Unexpected error from delete: %v", err.Error())
	}

	cfg.SetCache(&config.CacheConfig{
		Type:       cache.CacheTypeRedis,
		Servers:    []string{"localhost:1234"},
		Expiration: time.Second,
	})

	mp = cache.NewClient(cfg, nil, nil, nil)
	if mp == nil {
		t.Fatal("Unable to initialize redis client")
	}

	mp.SetMemcacheClient(nil)
	mp.SetRedisClient(&mockRedisClient{})

	res, err = mp.Get(context.Background(), "test")
	if err != nil {
		t.Errorf("Unexpected error from get: %v", err.Error())
	}

	if string(res.Value) != "test" {
		t.Errorf("Expected value: test, got: %v", res.Value)
	}

	resM, err = mp.GetMulti(context.Background(), "test")
	if err != nil {
		t.Errorf("Unexpected error from get: %v", err.Error())
	}

	if string(resM["test"].Value) != "test" {
		t.Errorf("Expected multi value: test, got: %v", resM)
	}

	err = mp.Set(context.Background(),
		&cache.Item{Key: "test", Value: []byte("test")})
	if err != nil {
		t.Errorf("Unexpected error from set: %v", err.Error())
	}

	_, err = mp.Get(context.Background(), "invalid")
	if err == nil {
		t.Error("Expected cache miss error, got: nil")
	}

	err = mp.Delete(context.Background(), "test")
	if err != nil {
		t.Errorf("Unexpected error from delete: %v", err.Error())
	}
}
