package config

import (
	"log/slog"
	"os"
	"strings"
)

const (
	LogLvlDebug = "debug"
	LogLvlInfo  = "info"
	LogLvlWarn  = "warn"
	LogLvlError = "error"
)

const (
	LogOutStderr = "stderr"
	LogOutStdout = "stdout"
)

const (
	LogFmtJSON = "json"
	LogFmtText = "text"
)

const (
	KeyLogLevel  = "log/level"
	KeyLogOut    = "log/out"
	KeyLogFormat = "log/format"

	DefaultLogLevel  = LogLvlInfo
	DefaultLogOut    = LogOutStderr
	DefaultLogFormat = LogFmtJSON
)

// LogConfig values represent log configuration data.
type LogConfig struct {
	Level  string `json:"level,omitempty"  yaml:"level,omitempty"`
	Out    string `json:"out,omitempty"    yaml:"out,omitempty"`
	Format string `json:"format,omitempty" yaml:"format,omitempty"`
}

// Load reads configuration data from environment variables and applies defaults
// for any missing or invalid configuration data.
func (c *LogConfig) Load() {
	if v := os.Getenv(ReplaceEnv(KeyLogLevel)); v != "" {
		c.Level = v
	}

	switch c.Level {
	case LogLvlDebug, LogLvlError, LogLvlInfo, LogLvlWarn:
	default:
		c.Level = DefaultLogLevel
	}

	if v := os.Getenv(ReplaceEnv(KeyLogOut)); v != "" {
		c.Out = v
	}

	switch c.Out {
	case LogOutStderr, LogOutStdout:
	default:
		c.Out = DefaultLogOut
	}

	if v := os.Getenv(ReplaceEnv(KeyLogFormat)); v != "" {
		c.Format = v
	}

	switch c.Format {
	case LogFmtJSON, LogFmtText:
	default:
		c.Out = DefaultLogFormat
	}
}

// LogLevel is the minimum (most verbose) level of log entries that should be
// written.
func (c *Config) LogLevel() slog.Level {
	c.RLock()
	defer c.RUnlock()

	ll := ""

	if c.log == nil {
		ll = DefaultLogLevel
	} else {
		ll = c.log.Level
	}

	switch strings.ToLower(ll) {
	case LogLvlDebug:
		return slog.LevelDebug
	case LogLvlInfo:
		return slog.LevelInfo
	case LogLvlWarn:
		return slog.LevelWarn
	case LogLvlError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// LogOut is output writer for log entries.
func (c *Config) LogOut() string {
	c.RLock()
	defer c.RUnlock()

	lo := DefaultLogOut

	if c.log != nil && c.log.Out != "" {
		lo = c.log.Out
	}

	return lo
}

// LogFormat is output format to use for log entries.
func (c *Config) LogFormat() string {
	c.RLock()
	defer c.RUnlock()

	lf := DefaultLogFormat

	if c.log != nil && c.log.Format != "" {
		lf = c.log.Format
	}

	return lf
}
