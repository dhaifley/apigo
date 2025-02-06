// Package config provides utility functions for managing configuration.
package config

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	DefaultAccount = "default"
)

// Config values represent full system configuration data.
type Config struct {
	sync.RWMutex
	auth      *AuthConfig
	cache     *CacheConfig
	db        *DBConfig
	log       *LogConfig
	telemetry *TelemetryConfig
	server    *ServerConfig
	service   *ServiceConfig
}

type configFile struct {
	Auth      *AuthConfig      `json:"auth,omitempty"      yaml:"auth,omitempty"`
	Cache     *CacheConfig     `json:"cache,omitempty"     yaml:"cache,omitempty"`
	DB        *DBConfig        `json:"db,omitempty"        yaml:"db,omitempty"`
	Log       *LogConfig       `json:"log,omitempty"       yaml:"log,omitempty"`
	Telemetry *TelemetryConfig `json:"telemetry,omitempty" yaml:"telemetry,omitempty"`
	Server    *ServerConfig    `json:"server,omitempty"    yaml:"server,omitempty"`
	Service   *ServiceConfig   `json:"service,omitempty"   yaml:"service,omitempty"`
}

// New creates a new configuration value.
func New(name string) *Config {
	return &Config{
		service: &ServiceConfig{Name: name},
	}
}

// NewDefault creates a new configuration value with default values and values
// loaded from environment variables. It is intended for fallback use when the
// configuration is not available for a service.
func NewDefault() *Config {
	cfg := &Config{}

	cfg.Load(nil)

	return cfg
}

// SetAuth applies authentication configuration data to the configuration.
func (c *Config) SetAuth(auth *AuthConfig) {
	c.Lock()
	defer c.Unlock()

	c.auth = auth
}

// SetCache applies cache configuration data to the configuration.
func (c *Config) SetCache(cache *CacheConfig) {
	c.Lock()
	defer c.Unlock()

	c.cache = cache
}

// SetDB applies database configuration data to the configuration.
func (c *Config) SetDB(db *DBConfig) {
	c.Lock()
	defer c.Unlock()

	c.db = db
}

// SetLog applies log configuration data to the configuration.
func (c *Config) SetLog(log *LogConfig) {
	c.Lock()
	defer c.Unlock()

	c.log = log
}

// SetTelemetry applies telemetry configuration data to the configuration.
func (c *Config) SetTelemetry(telemetry *TelemetryConfig) {
	c.Lock()
	defer c.Unlock()

	c.telemetry = telemetry
}

// SetServer applies server configuration data to the configuration.
func (c *Config) SetServer(server *ServerConfig) {
	c.Lock()
	defer c.Unlock()

	c.server = server
}

// SetService applies service configuration data to the configuration.
func (c *Config) SetService(service *ServiceConfig) {
	c.Lock()
	defer c.Unlock()

	c.service = service
}

// Load applies provided configuration data and populates missing configuration
// from environment variables and default values.
func (c *Config) Load(b []byte) {
	c.Lock()
	defer c.Unlock()

	if len(b) > 0 {
		if err := yaml.Unmarshal(b, &c); err != nil {
			if err := json.Unmarshal(b, &c); err != nil {
				os.Stderr.WriteString("unable to parse config data: " +
					err.Error())
			}
		}
	}

	if c.auth == nil {
		c.auth = &AuthConfig{}
	}

	c.auth.Load()

	if c.cache == nil {
		c.cache = &CacheConfig{}
	}

	c.cache.Load()

	if c.db == nil {
		c.db = &DBConfig{}
	}

	c.db.Load()

	if c.log == nil {
		c.log = &LogConfig{}
	}

	c.log.Load()

	if c.telemetry == nil {
		c.telemetry = &TelemetryConfig{}
	}

	c.telemetry.Load()

	if c.server == nil {
		c.server = &ServerConfig{}
	}

	c.server.Load()

	if c.service == nil {
		c.service = &ServiceConfig{}
	}

	c.service.Load()
}

// LoadFiles attempts to load any available configuration files.
func (c *Config) LoadFiles() {
	f := "api.yaml"

	b, err := os.ReadFile(f)
	if err != nil {
		os.Stderr.WriteString("unable to read config file: " + f +
			": " + err.Error() + "\n")
	}

	c.Load(b)
}

// UnmarshalJSON decodes a JSON format byte slice into this value.
func (c *Config) UnmarshalJSON(b []byte) error {
	var cf configFile

	buf := bytes.NewBuffer(b)

	if err := json.NewDecoder(buf).Decode(&cf); err != nil {
		return err
	}

	c.auth = cf.Auth
	c.cache = cf.Cache
	c.db = cf.DB
	c.log = cf.Log
	c.telemetry = cf.Telemetry
	c.server = cf.Server
	c.service = cf.Service

	return nil
}

// MarshalJSON encodes this value into a JSON format byte slice.
func (c *Config) MarshalJSON() ([]byte, error) {
	cf := configFile{
		Auth:      c.auth,
		Cache:     c.cache,
		DB:        c.db,
		Log:       c.log,
		Telemetry: c.telemetry,
		Server:    c.server,
		Service:   c.service,
	}

	buf := &bytes.Buffer{}

	if err := json.NewEncoder(buf).Encode(&cf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalYAML decodes a YAML format byte slice into this value.
func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	var cf configFile

	if err := value.Decode(&cf); err != nil {
		return err
	}

	c.auth = cf.Auth
	c.cache = cf.Cache
	c.db = cf.DB
	c.log = cf.Log
	c.telemetry = cf.Telemetry
	c.server = cf.Server
	c.service = cf.Service

	return nil
}

// MarshalYAML encodes a this value into the YAML format.
func (c *Config) MarshalYAML() (any, error) {
	if c == nil {
		return nil, nil
	}

	cf := &configFile{
		Auth:      c.auth,
		Cache:     c.cache,
		DB:        c.db,
		Log:       c.log,
		Telemetry: c.telemetry,
		Server:    c.server,
		Service:   c.service,
	}

	return cf, nil
}

// String returns the configuration object as a string.
func (c *Config) String() string {
	b, err := json.Marshal(c)
	if err != nil {
		return err.Error()
	}

	return string(b)
}

// YAML returns the configuration object as a YAML formatted string.
func (c *Config) YAML() string {
	buf := &bytes.Buffer{}

	enc := yaml.NewEncoder(buf)

	enc.SetIndent(2)

	if err := enc.Encode(c); err != nil {
		return err.Error()
	}

	return buf.String()
}

// ReplaceEnv replaces characters in a string to yield a configuration key
// environment variable name.
func ReplaceEnv(s string) string {
	return strings.ToUpper(strings.NewReplacer("/", "_", "-", "_").Replace(s))
}
