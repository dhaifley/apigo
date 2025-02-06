package config

import (
	"os"
	"strconv"
	"time"
)

const (
	KeyServerAddress        = "server/address"
	KeyServerCert           = "server/certificate"
	KeyServerKey            = "server/key"
	KeyServerTimeout        = "server/timeout"
	KeyServerIdleTimeout    = "server/idle_timeout"
	KeyServerHost           = "server/domain"
	KeyServerMaxRequestSize = "server/max_request_size"

	DefaultServerAddress        = ":8080"
	DefaultServerCert           = ""
	DefaultServerKey            = ""
	DefaultServerTimeout        = time.Second * 30
	DefaultServerIdleTimeout    = time.Second * 5
	DefaultServerHost           = "apid.io"
	DefaultServerMaxRequestSize = int64(20971520) // 20 MB
)

// ServerConfig values represent telemetry configuration data.
type ServerConfig struct {
	Address        string        `json:"address,omitempty"          yaml:"address,omitempty"`
	Cert           string        `json:"cert,omitempty"             yaml:"cert,omitempty"`
	Key            string        `json:"key,omitempty"              yaml:"key,omitempty"`
	Timeout        time.Duration `json:"timeout,omitempty"          yaml:"timeout,omitempty"`
	IdleTimeout    time.Duration `json:"idle_timeout,omitempty"     yaml:"idle_timeout,omitempty"`
	Host           string        `json:"host,omitempty"             yaml:"host,omitempty"`
	MaxRequestSize int64         `json:"max_request_size,omitempty" yaml:"max_request_size,omitempty"`
}

// Load reads configuration data from environment variables and applies defaults
// for any missing or invalid configuration data.
func (c *ServerConfig) Load() {
	if v := os.Getenv(ReplaceEnv(KeyServerAddress)); v != "" {
		c.Address = v
	}

	if c.Address == "" {
		c.Address = DefaultServerAddress
	}

	if v := os.Getenv(ReplaceEnv(KeyServerCert)); v != "" {
		c.Cert = v
	}

	if c.Cert == "" {
		c.Cert = DefaultServerCert
	}

	if v := os.Getenv(ReplaceEnv(KeyServerKey)); v != "" {
		c.Key = v
	}

	if c.Key == "" {
		c.Key = DefaultServerKey
	}

	if v := os.Getenv(ReplaceEnv(KeyServerTimeout)); v != "" {
		v, err := time.ParseDuration(v)
		if err != nil {
			v = DefaultServerTimeout
		}

		c.Timeout = v
	}

	if c.Timeout == 0 {
		c.Timeout = DefaultServerTimeout
	}

	if v := os.Getenv(ReplaceEnv(KeyServerIdleTimeout)); v != "" {
		v, err := time.ParseDuration(v)
		if err != nil {
			v = DefaultServerIdleTimeout
		}

		c.IdleTimeout = v
	}

	if c.IdleTimeout == 0 {
		c.IdleTimeout = DefaultServerIdleTimeout
	}

	if v := os.Getenv(ReplaceEnv(KeyServerHost)); v != "" {
		c.Host = v
	} else if v := os.Getenv("host"); v != "" {
		c.Host = v
	}

	if c.Host == "" {
		c.Host = DefaultServerHost
	}

	if v := os.Getenv(ReplaceEnv(KeyServerMaxRequestSize)); v != "" {
		v, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			v = DefaultServerMaxRequestSize
		}

		c.MaxRequestSize = v
	}

	if c.MaxRequestSize == 0 {
		c.MaxRequestSize = DefaultServerMaxRequestSize
	}
}

// ServerAddress returns the address of the collector where metrics data is
// sent.
func (c *Config) ServerAddress() string {
	c.RLock()
	defer c.RUnlock()

	if c.server == nil {
		return DefaultServerAddress
	}

	return c.server.Address
}

// ServerCert returns the name of a file containing the TLS certificate
// for the server.
func (c *Config) ServerCert() string {
	c.RLock()
	defer c.RUnlock()

	if c.server == nil {
		return DefaultServerCert
	}

	return c.server.Cert
}

// ServerKey returns the name of a file containing the private key for the TLS
// certificate used by the server.
func (c *Config) ServerKey() string {
	c.RLock()
	defer c.RUnlock()

	if c.server == nil {
		return DefaultServerKey
	}

	return c.server.Key
}

// ServerTimeout returns a duration representing the maximum time a server
// request is allowed to run before timing out.
func (c *Config) ServerTimeout() time.Duration {
	c.RLock()
	defer c.RUnlock()

	if c.server == nil {
		return DefaultServerTimeout
	}

	return c.server.Timeout
}

// ServerIdleTimeout returns a duration representing the maximum duration a
// keep-alive server request is allowed to remain idle before timing out.
func (c *Config) ServerIdleTimeout() time.Duration {
	c.RLock()
	defer c.RUnlock()

	if c.server == nil {
		return DefaultServerIdleTimeout
	}

	return c.server.IdleTimeout
}

// ServerHost returns the host name of the server.
func (c *Config) ServerHost() string {
	c.RLock()
	defer c.RUnlock()

	if c.server == nil {
		return DefaultServerHost
	}

	return c.server.Host
}

// ServerMaxRequestSize returns the maximum allowable request size in bytes.
func (c *Config) ServerMaxRequestSize() int64 {
	c.RLock()
	defer c.RUnlock()

	if c.server == nil {
		return DefaultServerMaxRequestSize
	}

	return c.server.MaxRequestSize
}
