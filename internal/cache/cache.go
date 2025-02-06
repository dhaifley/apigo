// The cache package implements an interface for performing cache operations.
package cache

import (
	"context"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/dhaifley/apid/internal/config"
	"github.com/dhaifley/apid/internal/errors"
	"github.com/dhaifley/apid/internal/logger"
	"github.com/dhaifley/apid/internal/metric"
	"github.com/google/gomemcache/memcache"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Types of cache supported.
const (
	CacheTypeMemcache = "memcache"
	CacheTypeRedis    = "redis"
)

// Item values contain a single cache key value item.
type Item struct {
	Key        string
	Value      []byte
	Expiration time.Duration
}

// Accessor values are used to interact with caches.
type Accessor interface {
	Get(ctx context.Context, key string) (*Item, error)
	GetMulti(ctx context.Context, keys ...string) (map[string]*Item, error)
	Set(ctx context.Context, item *Item) error
	Delete(ctx context.Context, key string) error
}

// memcacheClient values are used to interact with memcached clusters.
type memcacheClient interface {
	Get(key string) (*memcache.Item, error)
	GetMulti(keys []string) (map[string]*memcache.Item, error)
	Set(item *memcache.Item) error
	Delete(key string) error
}

// redisClient values are used to interact with redis clusters.
type redisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	MGet(ctx context.Context, keys ...string) *redis.SliceCmd
	Set(ctx context.Context, key string, value any,
		expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

// Client values are used for interacting with a group of cache servers.
type Client struct {
	sync.RWMutex
	servers   []string
	timeout   time.Duration
	discovery bool
	mc        memcacheClient
	rc        redisClient
	log       logger.Logger
	metric    metric.Recorder
	tracer    trace.Tracer
}

// NewClient initializes a new cache client.
func NewClient(cfg *config.Config,
	log logger.Logger,
	metric metric.Recorder,
	tracer trace.Tracer,
) *Client {
	if log == nil || (reflect.ValueOf(log).Kind() == reflect.Ptr &&
		reflect.ValueOf(log).IsNil()) {
		log = logger.NullLog
	}

	if metric == nil || (reflect.ValueOf(metric).Kind() == reflect.Ptr &&
		reflect.ValueOf(metric).IsNil()) {
		metric = nil
	}

	if tracer == nil || (reflect.ValueOf(tracer).Kind() == reflect.Ptr &&
		reflect.ValueOf(tracer).IsNil()) {
		tracer = nil
	}

	c := &Client{
		servers:   cfg.CacheServers(),
		timeout:   cfg.CacheTimeout(),
		discovery: cfg.CacheDiscovery(),
		log:       log,
		metric:    metric,
		tracer:    tracer,
	}

	switch cfg.CacheType() {
	case CacheTypeRedis:
		if len(c.servers) < 1 {
			return nil
		}

		opts := &redis.Options{
			Addr:                  c.servers[0],
			ContextTimeoutEnabled: true,
			DialTimeout:           c.timeout,
			ReadTimeout:           c.timeout,
			WriteTimeout:          c.timeout,
		}

		if i := cfg.CachePoolSize(); i != 0 {
			opts.PoolSize = i
		}

		rc := redis.NewClient(opts)

		c.rc = rc
		c.mc = nil
	case CacheTypeMemcache:
		if c.discovery && len(c.servers) > 0 {
			cli, err := memcache.NewDiscoveryClient(c.servers[0], c.timeout)
			if err != nil {
				log.Log(context.Background(), logger.LvlError,
					"unable to create memcache client",
					"error", err,
					"discovery", c.discovery,
					"timeout", c.timeout)

				return nil
			}

			c.mc = cli
		} else {
			ss, err := NewServerList(c.servers...)
			if err != nil {
				log.Log(context.Background(), logger.LvlError,
					"unable to create memcache client",
					"error", err,
					"discovery", c.discovery,
					"timeout", c.timeout)

				return nil
			}

			cli := memcache.NewFromSelector(ss)

			cli.Timeout = c.timeout

			c.mc = cli
		}

		c.rc = nil
	}

	return c
}

func (c *Client) SetMemcacheClient(cli memcacheClient) {
	c.Lock()

	c.mc = cli

	c.Unlock()
}

func (c *Client) SetRedisClient(cli redisClient) {
	c.Lock()

	c.rc = cli

	c.Unlock()
}

// Get attempts to retrieve the value of the specified key.
func (c *Client) Get(ctx context.Context, key string) (*Item, error) {
	c.RLock()

	rc, mc, mr := c.rc, c.mc, c.metric

	c.RUnlock()

	if rc == nil && mc == nil {
		return nil, errors.New(errors.ErrCache,
			"no cache connected")
	}

	select {
	case <-ctx.Done():
		return nil, errors.Context(ctx)
	default:
	}

	res := &Item{}

	ctx, finish := c.startCacheSpan(ctx, "get")

	if rc != nil {
		sc := rc.Get(ctx, key)

		val, err := sc.Result()

		finish(err)

		if err != nil {
			if err == redis.Nil {
				if mr != nil {
					mr.Increment(ctx, "cache_misses", "operation:get")
				}

				return nil, errors.New(errors.ErrNotFound,
					"key not found in cache")
			}

			if mr != nil {
				mr.Increment(ctx, "cache_errors", "operation:get")
			}

			return nil, errors.Wrap(err, errors.ErrCache,
				"unable to get cache item")
		}

		if mr != nil {
			mr.Increment(ctx, "cache_hits")

			mr.Add(ctx, "cache_hits_bytes", int64(len(val)))
		}

		res.Key = key
		res.Value = []byte(val)
	} else {
		item, err := mc.Get(key)

		finish(err)

		if err != nil || item == nil {
			if err == memcache.ErrCacheMiss {
				if mr != nil {
					mr.Increment(ctx, "cache_misses", "operation:get")
				}

				return nil, errors.New(errors.ErrNotFound,
					"key not found in cache")
			}

			if mr != nil {
				mr.Increment(ctx, "cache_errors", "operation:get")
			}

			return nil, errors.Wrap(err, errors.ErrCache,
				"unable to get cache item")
		}

		if mr != nil {
			mr.Increment(ctx, "cache_hits")

			mr.Add(ctx, "cache_hits_bytes", int64(len(item.Value)))
		}

		res.Key = item.Key
		res.Value = item.Value
		res.Expiration = time.Duration(item.Expiration) * time.Second
	}

	return res, nil
}

// GetMulti attempts to retrieve a map of the values of the specified keys.
func (c *Client) GetMulti(ctx context.Context,
	keys ...string,
) (map[string]*Item, error) {
	c.RLock()

	rc, mc, mr := c.rc, c.mc, c.metric

	c.RUnlock()

	if rc == nil && mc == nil {
		return nil, errors.New(errors.ErrCache,
			"no cache connected")
	}

	select {
	case <-ctx.Done():
		return nil, errors.Context(ctx)
	default:
	}

	res := map[string]*Item{}

	ctx, finish := c.startCacheSpan(ctx, "get_multi")

	if rc != nil {
		sc := rc.MGet(ctx, keys...)

		vs, err := sc.Result()

		finish(err)

		if err != nil || vs == nil {
			if err == redis.Nil {
				if mr != nil {
					mr.Increment(ctx, "cache_misses", "operation:get_multi")
				}

				return nil, errors.New(errors.ErrNotFound,
					"keys not found in cache")
			}

			if mr != nil {
				mr.Increment(ctx, "cache_errors", "operation:get_multi")
			}

			return nil, errors.Wrap(err, errors.ErrCache,
				"unable to get cache items")
		}

		if len(vs) != len(keys) {
			return nil, errors.New(errors.ErrCache,
				"invalid response received from cache")
		}

		for i, key := range keys {
			vs, ok := vs[i].(string)
			if !ok {
				if mr != nil {
					mr.Increment(ctx, "cache_misses", "operation:get_multi_key")
				}

				continue
			}

			val := []byte(vs)

			if mr != nil {
				mr.Increment(ctx, "cache_hits")

				mr.Add(ctx, "cache_hits_bytes", int64(len(val)))
			}

			res[key] = new(Item)
			res[key].Key = key
			res[key].Value = []byte(val)
		}
	} else {
		items, err := c.mc.GetMulti(keys)

		finish(err)

		if err != nil {
			if err == memcache.ErrCacheMiss {
				if mr != nil {
					mr.Increment(ctx, "cache_misses", "operation:get_multi")
				}

				return nil, errors.New(errors.ErrNotFound,
					"keys not found in cache")
			}

			if mr != nil {
				mr.Increment(ctx, "cache_errors", "operation:get_multi")
			}

			return nil, errors.Wrap(err, errors.ErrCache,
				"unable to get cache items")
		}

		if items == nil {
			return nil, errors.New(errors.ErrCache,
				"null multi response received from cache")
		}

		for _, key := range keys {
			item, ok := items[key]
			if !ok || item == nil {
				if mr != nil {
					mr.Increment(ctx, "cache_misses", "operation:get_multi_key")
				}

				continue
			}

			if mr != nil {
				mr.Increment(ctx, "cache_hits")

				mr.Add(ctx, "cache_hits_bytes", int64(len(item.Value)))
			}

			res[key] = new(Item)
			res[key].Key = key
			res[key].Value = item.Value
			res[key].Expiration = time.Duration(item.Expiration) * time.Second
		}
	}

	return res, nil
}

// Set attempts to set the value of the specified key.
func (c *Client) Set(ctx context.Context, item *Item) error {
	if item == nil {
		return errors.New(errors.ErrCache,
			"unable to cache null item")
	}

	c.RLock()

	rc, mc, mr := c.rc, c.mc, c.metric

	c.RUnlock()

	if rc == nil && mc == nil {
		return errors.New(errors.ErrCache,
			"no cache connected")
	}

	select {
	case <-ctx.Done():
		return errors.Context(ctx)
	default:
	}

	ctx, finish := c.startCacheSpan(ctx, "set")

	var err error

	if rc != nil {
		sc := rc.Set(ctx, item.Key, string(item.Value), item.Expiration)

		err = sc.Err()
	} else {
		req := memcache.Item{
			Key:        item.Key,
			Value:      item.Value,
			Expiration: int32(item.Expiration.Seconds()),
		}

		err = mc.Set(&req)
	}

	finish(err)

	if err != nil {
		if mr != nil {
			mr.Increment(ctx, "cache_errors", "operation:set")
		}

		return errors.Wrap(err, errors.ErrCache,
			"unable to set cache item")
	}

	if mr != nil {
		mr.Increment(ctx, "cache_sets")

		mr.Add(ctx, "cache_sets_bytes", int64(len(item.Value)))
	}

	return nil
}

// Delete attempts to remove the value of the specified key.
func (c *Client) Delete(ctx context.Context, key string) error {
	c.RLock()

	rc := c.rc
	mc := c.mc
	mr := c.metric

	c.RUnlock()

	if rc == nil && mc == nil {
		return errors.New(errors.ErrCache,
			"no cache connected")
	}

	select {
	case <-ctx.Done():
		return errors.Context(ctx)
	default:
	}

	ctx, finish := c.startCacheSpan(ctx, "delete")

	if rc != nil {
		ic := rc.Del(ctx, key)

		_, err := ic.Result()

		finish(err)

		if err != nil {
			if mr != nil {
				mr.Increment(ctx, "cache_errors", "operation:delete")
			}

			return errors.Wrap(err, errors.ErrCache,
				"unable to delete from cache")
		}

		if mr != nil {
			mr.Increment(ctx, "cache_deletes")
		}
	} else {
		err := mc.Delete(key)

		finish(err)

		if err != nil {
			if err == memcache.ErrCacheMiss {
				// Do not record a cache miss for a missed delete.
				return nil
			}

			if mr != nil {
				mr.Increment(ctx, "cache_errors", "operation:delete")
			}

			return errors.Wrap(err, errors.ErrCache,
				"unable to delete from cache")
		}

		if mr != nil {
			mr.Increment(ctx, "cache_deletes")
		}
	}

	return nil
}

// startCacheSpan starts a cache tracing span. It returns an updated context,
// and a closing function.
func (c *Client) startCacheSpan(ctx context.Context, name string,
) (context.Context, func(err error)) {
	c.RLock()

	servers := c.servers
	tracer := c.tracer
	mr := c.metric
	mc := c.mc

	c.RUnlock()

	start := time.Now()

	cType := CacheTypeRedis
	if mc != nil {
		cType = CacheTypeMemcache
	}

	if tracer == nil {
		return ctx, func(err error) {
			if mr != nil {
				mr.RecordDuration(ctx, "cache_latency",
					time.Since(start), "operation:"+name)
			}
		}
	}

	ctx, span := tracer.Start(ctx, name,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("cache.system", cType),
			attribute.String("cache.servers", strings.Join(servers, " ")),
		),
	)

	return ctx, func(err error) {
		if err != nil && reflect.ValueOf(err).Kind() == reflect.Ptr &&
			reflect.ValueOf(err).IsNil() {
			err = nil
		}

		if span != nil {
			if err != nil {
				span.SetStatus(codes.Error, name+" failed")
				span.RecordError(err)
			}

			span.End()
		}

		if mr != nil {
			mr.RecordDuration(ctx, "cache_latency",
				time.Since(start), "operation:"+name)
		}
	}
}

// MockCache values are used to test caching.
type MockCache struct {
	sync.RWMutex
	items  map[string]*Item
	hit    bool
	miss   bool
	set    bool
	delete bool
}

// WasHit returns whether the cache was hit.
func (m *MockCache) WasHit() bool {
	m.RLock()

	defer m.RUnlock()

	return m.hit
}

// WasMissed returns whether the cache was missed.
func (m *MockCache) WasMissed() bool {
	m.RLock()

	defer m.RUnlock()

	return m.miss
}

// WasSet returns whether a cache item was set.
func (m *MockCache) WasSet() bool {
	m.RLock()

	defer m.RUnlock()

	return m.set
}

// WasDeleted returns whether a cache item was deleted.
func (m *MockCache) WasDeleted() bool {
	m.RLock()

	defer m.RUnlock()

	return m.delete
}

// Item returns the cache items.
func (m *MockCache) Items() map[string]*Item {
	m.RLock()

	defer m.RUnlock()

	return m.items
}

// Get simulates a cache get.
func (m *MockCache) Get(ctx context.Context, key string) (*Item, error) {
	m.Lock()

	defer m.Unlock()

	if m.items == nil {
		m.miss = true

		return nil, errors.New(errors.ErrNotFound,
			"key not found in cache")
	}

	for k, i := range m.items {
		if k == key {
			m.hit = true

			return i, nil
		}
	}

	m.miss = true

	return nil, errors.New(errors.ErrNotFound,
		"key not found in cache")
}

// GetMulti simulates a cache get_multi.
func (m *MockCache) GetMulti(ctx context.Context,
	keys ...string,
) (map[string]*Item, error) {
	m.Lock()

	defer m.Unlock()

	if m.items == nil {
		m.miss = true

		return nil, errors.New(errors.ErrNotFound,
			"keys not found in cache")
	}

	for _, key := range keys {
		for k, i := range m.items {
			if k == key {
				m.hit = true

				return map[string]*Item{
					key: i,
				}, nil
			}
		}
	}

	m.miss = true

	return nil, errors.New(errors.ErrNotFound,
		"keys not found in cache")
}

// Set simulates a cache set.
func (m *MockCache) Set(ctx context.Context, item *Item) error {
	m.Lock()

	defer m.Unlock()

	if m.items == nil {
		m.items = map[string]*Item{}
	}

	m.items[item.Key] = item

	m.set = true

	return nil
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	m.Lock()

	defer m.Unlock()

	m.delete = true

	if m.items == nil {
		return nil
	}

	delete(m.items, key)

	return nil
}
