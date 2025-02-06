package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/dhaifley/apid/config"
)

const testConfigYAML = `auth:
  token_jwks: '{}'
  token_well_known: /.well-known/openid-configuration
  token_expires_in: 24h0m0s
  token_refresh_expires_in: 24h0m0s
  token_issuer: api
  update_interval: 30s
  identity_domain: localhost
cache:
  type: memcache
  timeout: 10s
  expiration: 10s
  max_bytes: 1024
  pool_size: 20
db:
  user: api-db-user
  password: api
  database: api-db
  migrate_user: postgres
  migrate_password: postgres
  migrate_database: postgres
  host: localhost
  port: "5432"
  max_connections: 20
  type: postgres
  ssl_mode: disable
  monitor: 30s
  default_size: 100
  max_size: 10000
log:
  level: debug
  out: stdout
  format: text
telemetry:
  metric_address: localhost:8888
  metric_interval: 2m0s
  metric_version: v0.1.1
  trace_address: localhost:8888
server:
  address: :8080
  timeout: 10s
  idle_timeout: 5s
  host: example.com
  max_request_size: 20971520
service:
  name: api
  import_interval: 5m0s
  resource_data_retention: 720h0m0s` + "\n"

const testConfigJSON = `{"auth":{"token_jwks":"{}",` +
	`"token_well_known":"/.well-known/openid-configuration",` +
	`"token_expires_in":86400000000000,` +
	`"token_refresh_expires_in":86400000000000,"token_issuer":"api",` +
	`"update_interval":30000000000,"identity_domain":"localhost"},` +
	`"cache":{"type":"memcache","timeout":1000000000,"expiration":1000000000,` +
	`"max_bytes":1048576,"pool_size":20},` +
	`"db":{"user":"api-db-user","password":"api",` +
	`"database":"api-db","migrate_user":"postgres",` +
	`"migrate_password":"postgres","migrate_database":"postgres",` +
	`"host":"localhost","port":"5432","max_connections":20,"type":"postgres",` +
	`"ssl_mode":"disable","monitor":30000000000,"default_size":100,` +
	`"max_size":10000},` +
	`"log":{"level":"debug","out":"stdout","format":"text"},` +
	`"telemetry":{"metric_address":"localhost:8888",` +
	`"metric_interval":60000000000,"metric_version":"v0.1.1",` +
	`"trace_address":"localhost:8888"},` +
	`"server":{"address":":8080",` +
	`"timeout":30000000000,"idle_timeout":5000000000,"host":"example.com",` +
	`"max_request_size":20971520},` +
	`"service":{"name":"api","import_interval":30000000000,` +
	`"resource_data_retention":2592000000000000}}`

func TestConfig(t *testing.T) {
	t.Parallel()

	os.Clearenv()

	cfg := config.New("test")

	cfg.Load([]byte(testConfigJSON))

	if !strings.Contains(cfg.String(), testConfigJSON) {
		t.Errorf("Expected config string: %v, got: %v",
			testConfigJSON, cfg.String())
	}

	cfg.Load([]byte(testConfigYAML))

	if cfg.YAML() != testConfigYAML {
		t.Errorf("Expected config yaml: %v, got: %v",
			testConfigYAML, cfg.YAML())
	}

	cfg = config.NewDefault()

	if cfg.ServiceName() != config.DefaultServiceName {
		t.Errorf("Expected default config service name, got: %v",
			cfg.ServiceName())
	}
}
