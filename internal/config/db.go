package config

import (
	"net"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	KeyDBConn            = "db/connection"
	KeyDBUser            = "db/user"
	KeyDBPassword        = "db/password"
	KeyDBDatabase        = "db/database"
	KeyDBMigrateUser     = "postgres/user"
	KeyDBMigratePassword = "postgres/password"
	KeyDBMigrateDatabase = "postgres/database"
	KeyDBInstance        = "db/instance"
	KeyDBPrivateIP       = "db/private_ip"
	KeyDBHost            = "postgres/host"
	KeyDBPort            = "postgres/port"
	KeyDBMaxConns        = "db/max_connections"
	KeyDBType            = "db/type"
	KeyDBSSLMode         = "db/ssl_mode"
	KeyDBMonitor         = "db/monitor"
	KeyDBDefaultSize     = "db/default_size"
	KeyDBMaxSize         = "db/max_size"
	KeyDBMigrations      = "db/migrations"

	DefaultDBConn            = ""
	DefaultDBUser            = "api-db-user"
	DefaultDBPassword        = "api"
	DefaultDBDatabase        = "api-db"
	DefaultDBMigrateUser     = "postgres"
	DefaultDBMigratePassword = "postgres"
	DefaultDBMigrateDatabase = "postgres"
	DefaultDBInstance        = ""
	DefaultDBPrivateIP       = ""
	DefaultDBHost            = "postgres"
	DefaultDBPort            = "5432"
	DefaultDBMaxConns        = 20
	DefaultDBType            = "postgres"
	DefaultDBSSLMode         = "disable"
	DefaultDBMonitor         = time.Second * 30
	DefaultDBDefaultSize     = 100
	DefaultDBMaxSize         = 10000
	DefaultDBMigrations      = ""
)

const (
	DBModeNormal = iota
	DBModeMigrate
	DBModeInit
)

// DBConfig values represent database configuration data.
type DBConfig struct {
	Conn            string        `json:"connection,omitempty"       yaml:"connection,omitempty"`
	User            string        `json:"user,omitempty"             yaml:"user,omitempty"`
	Password        string        `json:"password,omitempty"         yaml:"password,omitempty"`
	Database        string        `json:"database,omitempty"         yaml:"database,omitempty"`
	MigrateUser     string        `json:"migrate_user,omitempty"     yaml:"migrate_user,omitempty"`
	MigratePassword string        `json:"migrate_password,omitempty" yaml:"migrate_password,omitempty"`
	MigrateDatabase string        `json:"migrate_database,omitempty" yaml:"migrate_database,omitempty"`
	Instance        string        `json:"instance,omitempty"         yaml:"instance,omitempty"`
	PrivateIP       string        `json:"private_ip,omitempty"       yaml:"private_ip,omitempty"`
	Host            string        `json:"host,omitempty"             yaml:"host,omitempty"`
	Port            string        `json:"port,omitempty"             yaml:"port,omitempty"`
	MaxConns        int64         `json:"max_connections,omitempty"  yaml:"max_connections,omitempty"`
	Type            string        `json:"type,omitempty"             yaml:"type,omitempty"`
	SSLMode         string        `json:"ssl_mode,omitempty"         yaml:"ssl_mode,omitempty"`
	Monitor         time.Duration `json:"monitor,omitempty"          yaml:"monitor,omitempty"`
	DefaultSize     int64         `json:"default_size,omitempty"     yaml:"default_size,omitempty"`
	MaxSize         int64         `json:"max_size,omitempty"         yaml:"max_size,omitempty"`
	Migrations      string        `json:"migrations,omitempty"       yaml:"migrations,omitempty"`
}

// Load reads configuration data from environment variables and applies defaults
// for any missing or invalid configuration data.
func (c *DBConfig) Load() {
	if v := os.Getenv(ReplaceEnv(KeyDBConn)); v != "" {
		c.Conn = v
	}

	if c.Conn == "" {
		c.Conn = DefaultDBConn
	}

	if v := os.Getenv(ReplaceEnv(KeyDBUser)); v != "" {
		c.User = v
	}

	if c.User == "" {
		c.User = DefaultDBUser
	}

	if v := os.Getenv(ReplaceEnv(KeyDBPassword)); v != "" {
		c.Password = v
	}

	if c.Password == "" {
		c.Password = DefaultDBPassword
	}

	if v := os.Getenv(ReplaceEnv(KeyDBDatabase)); v != "" {
		c.Database = v
	}

	if c.Database == "" {
		c.Database = DefaultDBDatabase
	}

	if v := os.Getenv(ReplaceEnv(KeyDBMigrateUser)); v != "" {
		c.MigrateUser = v
	}

	if c.MigrateUser == "" {
		c.MigrateUser = DefaultDBMigrateUser
	}

	if v := os.Getenv(ReplaceEnv(KeyDBMigratePassword)); v != "" {
		c.MigratePassword = v
	}

	if c.MigratePassword == "" {
		c.MigratePassword = DefaultDBMigratePassword
	}

	if v := os.Getenv(ReplaceEnv(KeyDBMigrateDatabase)); v != "" {
		c.MigrateDatabase = v
	}

	if c.MigrateDatabase == "" {
		c.MigrateDatabase = DefaultDBMigrateDatabase
	}

	if v := os.Getenv(ReplaceEnv(KeyDBInstance)); v != "" {
		c.Instance = v
	}

	if c.Instance == "" {
		c.Instance = DefaultDBInstance
	}

	if v := os.Getenv(ReplaceEnv(KeyDBPrivateIP)); v != "" {
		c.PrivateIP = v
	}

	if c.PrivateIP == "" {
		c.PrivateIP = DefaultDBPrivateIP
	}

	if v := os.Getenv(ReplaceEnv(KeyDBHost)); v != "" {
		c.Host = v
	}

	if c.Host == "" {
		c.Host = DefaultDBHost
	}

	if v := os.Getenv(ReplaceEnv(KeyDBPort)); v != "" {
		c.Port = v
	}

	if c.Port == "" {
		c.Port = DefaultDBPort
	}

	if v := os.Getenv(ReplaceEnv(KeyDBMaxConns)); v != "" {
		v, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			v = DefaultDBMaxConns
		}

		c.MaxConns = v
	}

	if c.MaxConns == 0 {
		c.MaxConns = DefaultDBMaxConns
	}

	if v := os.Getenv(ReplaceEnv(KeyDBType)); v != "" {
		c.Type = v
	}

	if c.Type == "" {
		c.Type = DefaultDBType
	}

	if v := os.Getenv(ReplaceEnv(KeyDBSSLMode)); v != "" {
		c.SSLMode = v
	}

	if c.SSLMode == "" {
		c.SSLMode = DefaultDBSSLMode
	}

	if v := os.Getenv(ReplaceEnv(KeyDBMonitor)); v != "" {
		v, err := time.ParseDuration(v)
		if err != nil {
			v = DefaultDBMonitor
		}

		c.Monitor = v
	}

	if c.Monitor == 0 {
		c.Monitor = DefaultDBMonitor
	}

	if v := os.Getenv(ReplaceEnv(KeyDBDefaultSize)); v != "" {
		v, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			v = DefaultDBDefaultSize
		}

		c.DefaultSize = v
	}

	if c.DefaultSize == 0 {
		c.DefaultSize = DefaultDBDefaultSize
	}

	if v := os.Getenv(ReplaceEnv(KeyDBMaxSize)); v != "" {
		v, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			v = DefaultDBMaxSize
		}

		c.MaxSize = v
	}

	if c.MaxSize == 0 {
		c.MaxSize = DefaultDBMaxSize
	}

	if v := os.Getenv(ReplaceEnv(KeyDBMigrations)); v != "" {
		c.Migrations = v
	}

	if c.Migrations == "" {
		c.Migrations = DefaultDBMigrations
	}
}

// DBConn returns the connection string used by the primary database
// connection pool.
func (c *Config) DBConn(mode int) string {
	if c.db == nil {
		return DefaultDBConn
	}

	conn := c.db.Conn

	if conn == "" {
		dbType := url.PathEscape(c.db.Type)

		if dbType == "" {
			dbType = DefaultDBType
		}

		dbUser := url.QueryEscape(c.db.User)

		dbPassword := url.QueryEscape(c.db.Password)

		dbDatabase := url.PathEscape(c.db.Database)

		if mode == DBModeMigrate || mode == DBModeInit {
			dbUser = url.QueryEscape(c.db.MigrateUser)

			dbPassword = url.QueryEscape(c.db.MigratePassword)
		}

		if mode == DBModeInit {
			dbDatabase = url.PathEscape(c.db.MigrateDatabase)
		}

		dbHost := c.db.Host

		if c.db.PrivateIP != "" {
			dbHost = c.db.PrivateIP
		}

		dbPort := c.db.Port

		dbSSLMode := url.QueryEscape(c.db.SSLMode)

		if dbType != "" && dbUser != "" &&
			dbHost != "" && dbDatabase != "" {
			conn = dbType + "://" + dbUser

			if dbPassword != "" {
				conn += ":" + dbPassword
			}

			host := dbHost

			if dbPort != "" {
				host = net.JoinHostPort(host, dbPort)
			}

			conn += "@" + host + "/" + dbDatabase + "?sslmode=" + dbSSLMode
		}
	}

	return conn
}

// DBUser returns the user used by the primary database connection pool.
func (c *Config) DBUser() string {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBUser
	}

	return c.db.User
}

// DBPassword returns the password used by the primary database connection pool.
func (c *Config) DBPassword() string {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBPassword
	}

	return c.db.Password
}

// DBDatabase returns the database used by the primary database connection pool.
func (c *Config) DBDatabase() string {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBDatabase
	}

	return c.db.Database
}

// DBMigrateUser returns the user used by the database connection pool to apply
// database migrations.
func (c *Config) DBMigrateUser() string {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBMigrateUser
	}

	return c.db.MigrateUser
}

// DBMigratePassword returns the password used by the database connection pool
// to apply database migrations.
func (c *Config) DBMigratePassword() string {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBPassword
	}

	return c.db.MigratePassword
}

// DBMigrateDatabase returns the database used when initializing the database.
func (c *Config) DBMigrateDatabase() string {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBMigrateDatabase
	}

	return c.db.MigrateDatabase
}

// DBInstance returns the instance used by the primary database connection pool.
func (c *Config) DBInstance() string {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBInstance
	}

	return c.db.Instance
}

// DBPrivateIP returns the private IP address used by the primary database
// connection pool.
func (c *Config) DBPrivateIP() string {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBPrivateIP
	}

	return c.db.PrivateIP
}

// DBHost returns the host name used by the primary database connection pool.
func (c *Config) DBHost() string {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBHost
	}

	return c.db.Host
}

// DBPort returns the port used by the primary database connection pool.
func (c *Config) DBPort() string {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBPort
	}

	return c.db.Port
}

// DBMaxConns returns the maximum number of open database connections allowed.
func (c *Config) DBMaxConns() int64 {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBMaxConns
	}

	return c.db.MaxConns
}

// DBType returns the type of database used by the service.
func (c *Config) DBType() string {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBType
	}

	return c.db.Type
}

// DBSSLMode returns the SSL connection mode of database used by the service.
func (c *Config) DBSSLMode() string {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBSSLMode
	}

	return c.db.SSLMode
}

// DBMonitor returns the frequency at which database connection pools are
// monitored and checked for health.
func (c *Config) DBMonitor() time.Duration {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBMonitor
	}

	return c.db.Monitor
}

// DBDefaultSize returns the default number of rows any query will return.
func (c *Config) DBDefaultSize() int64 {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBDefaultSize
	}

	return c.db.DefaultSize
}

// DBMaxSize returns the maximum limit of rows any query may return.
func (c *Config) DBMaxSize() int64 {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBMaxSize
	}

	return c.db.MaxSize
}

// DBMigrations returns the GitHub access secret for the database
// migrations files.
func (c *Config) DBMigrations() string {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBMigrations
	}

	return c.db.Migrations
}
