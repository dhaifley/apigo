package config_test

import (
	"testing"
	"time"

	"github.com/dhaifley/apid/config"
)

func TestCacheConfig(t *testing.T) {
	t.Parallel()

	cfg := config.New("")

	cfg.Load(nil)

	cfg.SetCache(&config.CacheConfig{
		Type:       "memcache",
		Servers:    []string{"test", "test2"},
		Discovery:  true,
		Timeout:    time.Second * 5,
		Expiration: time.Second * 10,
		MaxBytes:   1024,
		PoolSize:   1,
	})

	if cfg.CacheType() != "memcache" {
		t.Errorf("Expected cache type: memcache, got: %v", cfg.CacheType())
	}

	if cfg.CacheServers()[0] != "test" {
		t.Errorf("Expected cache server: test, got: %v", cfg.CacheServers()[0])
	}

	if !cfg.CacheDiscovery() {
		t.Errorf("Expected cache discovery, got: %v", cfg.CacheDiscovery())
	}

	if cfg.CacheTimeout() != (time.Second * 5) {
		t.Errorf("Expected cache timeout: 5s, got: %v", cfg.CacheTimeout())
	}

	if cfg.CacheExpiration() != time.Second*10 {
		t.Errorf("Expected cache expiration: 10s, got: %v",
			cfg.CacheExpiration())
	}

	if cfg.CacheMaxBytes() != 1024 {
		t.Errorf("Expected cache max bytes: 1024, got: %v", cfg.CacheMaxBytes())
	}

	if cfg.CachePoolSize() != 1 {
		t.Errorf("Expected cache pool size: 1, got: %v", cfg.CachePoolSize())
	}
}
