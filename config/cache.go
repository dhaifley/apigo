package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	KeyCacheType       = "cache/type"
	KeyCacheServers    = "cache/servers"
	KeyCacheDiscovery  = "cache/discovery"
	KeyCacheTimeout    = "cache/timeout"
	KeyCacheExpiration = "cache/expiration"
	KeyCacheMaxBytes   = "cache/max_bytes"
	KeyCachePoolSize   = "cache/pool_size"

	DefaultCacheType       = "redis"
	DefaultCacheDiscovery  = false
	DefaultCacheTimeout    = time.Second
	DefaultCacheExpiration = time.Minute * 5
	DefaultCacheMaxBytes   = 1048576
	DefaultCachePoolSize   = 10
)

// CacheConfig values represent cache configuration data.
type CacheConfig struct {
	Type       string        `json:"type,omitempty"       yaml:"type,omitempty"`
	Servers    []string      `json:"servers,omitempty"    yaml:"servers,omitempty"`
	Discovery  bool          `json:"discovery,omitempty"  yaml:"discovery,omitempty"`
	Timeout    time.Duration `json:"timeout,omitempty"    yaml:"timeout,omitempty"`
	Expiration time.Duration `json:"expiration,omitempty" yaml:"expiration,omitempty"`
	MaxBytes   int           `json:"max_bytes,omitempty"  yaml:"max_bytes,omitempty"`
	PoolSize   int           `json:"pool_size,omitempty"  yaml:"pool_size,omitempty"`
}

// Load reads configuration data from environment variables and applies defaults
// for any missing or invalid configuration data.
func (c *CacheConfig) Load() {
	if v := os.Getenv(ReplaceEnv(KeyCacheType)); v != "" {
		c.Type = v
	}

	if c.Type == "" {
		c.Type = DefaultCacheType
	}

	if v := os.Getenv(ReplaceEnv(KeyCacheServers)); v != "" {
		c.Servers = strings.Split(v, " ")
	}

	if c.Servers == nil {
		c.Servers = []string{}
	}

	if v := os.Getenv(ReplaceEnv(KeyCacheDiscovery)); v != "" {
		v, err := strconv.ParseBool(v)
		if err != nil {
			v = DefaultCacheDiscovery
		}

		c.Discovery = v
	}

	if v := os.Getenv(ReplaceEnv(KeyCacheTimeout)); v != "" {
		v, err := time.ParseDuration(v)
		if err != nil {
			v = DefaultCacheTimeout
		}

		c.Timeout = v
	}

	if c.Timeout == 0 {
		c.Timeout = DefaultCacheTimeout
	}

	if v := os.Getenv(ReplaceEnv(KeyCacheExpiration)); v != "" {
		v, err := time.ParseDuration(v)
		if err != nil {
			v = DefaultCacheExpiration
		}

		c.Expiration = v
	}

	if c.Expiration == 0 {
		c.Expiration = DefaultCacheExpiration
	}

	if v := os.Getenv(ReplaceEnv(KeyCacheMaxBytes)); v != "" {
		v, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			v = DefaultCacheMaxBytes
		}

		c.MaxBytes = int(v)
	}

	if c.MaxBytes == 0 {
		c.MaxBytes = DefaultCacheMaxBytes
	}

	if v := os.Getenv(ReplaceEnv(KeyCachePoolSize)); v != "" {
		v, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			v = DefaultCachePoolSize
		}

		c.PoolSize = int(v)
	}

	if c.PoolSize == 0 {
		c.PoolSize = DefaultCachePoolSize
	}
}

// CacheType returns the type of cache service used.
func (c *Config) CacheType() string {
	c.RLock()
	defer c.RUnlock()

	if c.cache == nil {
		return DefaultCacheType
	}

	return c.cache.Type
}

// CacheServers returns a list of available cache servers.
func (c *Config) CacheServers() []string {
	c.RLock()
	defer c.RUnlock()

	if c.cache == nil {
		return nil
	}

	return c.cache.Servers
}

// CacheDiscovery returns whether to connect to an auto-discovery address.
func (c *Config) CacheDiscovery() bool {
	c.RLock()
	defer c.RUnlock()

	if c.cache == nil {
		return DefaultCacheDiscovery
	}

	return c.cache.Discovery
}

// CacheTimeout returns the timeout duration used for cache requests.
func (c *Config) CacheTimeout() time.Duration {
	c.RLock()
	defer c.RUnlock()

	if c.cache == nil {
		return DefaultCacheTimeout
	}

	return c.cache.Timeout
}

// CacheExpiration returns the expiration in seconds used for cache items.
func (c *Config) CacheExpiration() time.Duration {
	c.RLock()
	defer c.RUnlock()

	if c.cache == nil {
		return DefaultCacheExpiration
	}

	return c.cache.Expiration
}

// CacheMaxBytes returns the maximum bytes allowed for cache items.
func (c *Config) CacheMaxBytes() int {
	c.RLock()
	defer c.RUnlock()

	if c.cache == nil {
		return DefaultCacheMaxBytes
	}

	return c.cache.MaxBytes
}

// CachePoolSize returns the maximum pool size for cache connections.
func (c *Config) CachePoolSize() int {
	c.RLock()
	defer c.RUnlock()

	if c.cache == nil {
		return DefaultCachePoolSize
	}

	return c.cache.PoolSize
}
