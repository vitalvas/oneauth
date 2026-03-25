package config

import (
	"fmt"
	"net/url"
	"time"

	"github.com/vitalvas/gokit/xconfig"
)

type Config struct {
	Server   ServerConfig   `json:"server" yaml:"server"`
	Database DatabaseConfig `json:"database" yaml:"database"`
	Security SecurityConfig `json:"security" yaml:"security"`
	Logging  LoggingConfig  `json:"logging" yaml:"logging"`
}

type ServerConfig struct {
	Address string `json:"address" yaml:"address"`
}

type SecurityConfig struct {
	MasterKey string `json:"master_key" yaml:"master_key"`
}

type LoggingConfig struct {
	Level  string `json:"level" yaml:"level"`
	Format string `json:"format" yaml:"format"`
}

type DatabaseConfig struct {
	Type string `json:"type" yaml:"type"`

	PostgreSQL *PostgreSQLConfig `json:"postgres,omitempty" yaml:"postgres,omitempty"`
	SQLite     *SQLiteConfig     `json:"sqlite,omitempty" yaml:"sqlite,omitempty"`
}

type PostgreSQLConfig struct {
	URL               string        `json:"url" yaml:"url"`
	MaxConnections    int           `json:"max_connections" yaml:"max_connections"`
	ConnectionTimeout time.Duration `json:"connection_timeout" yaml:"connection_timeout"`
}

type SQLiteConfig struct {
	Path        string `json:"path" yaml:"path"`
	JournalMode string `json:"journal_mode" yaml:"journal_mode"`
	Synchronous string `json:"synchronous" yaml:"synchronous"`
}

func (c *ServerConfig) Default() {
	*c = ServerConfig{
		Address: "localhost:8002",
	}
}

func (c *LoggingConfig) Default() {
	*c = LoggingConfig{
		Level:  "info",
		Format: "json",
	}
}

func (c *DatabaseConfig) Default() {
	*c = DatabaseConfig{
		Type: "sqlite",
	}
}

func (c *PostgreSQLConfig) Default() {
	*c = PostgreSQLConfig{
		MaxConnections:    25,
		ConnectionTimeout: 30 * time.Second,
	}
}

func (c *SQLiteConfig) Default() {
	*c = SQLiteConfig{
		Path:        "/var/lib/oneauth/yubikey_ksm.db",
		JournalMode: "WAL",
		Synchronous: "NORMAL",
	}
}

func Load(configPath string) (*Config, error) {
	config := &Config{}

	options := []xconfig.Option{
		xconfig.WithEnv("ONEAUTH_KSM"),
	}

	if configPath != "" {
		options = append(options, xconfig.WithFiles(configPath))
	}

	if err := xconfig.Load(config, options...); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

func (c *Config) validate() error {
	if c.Server.Address == "" {
		return fmt.Errorf("server address cannot be empty")
	}

	if c.Database.Type == "" {
		return fmt.Errorf("database type cannot be empty")
	}

	switch c.Database.Type {
	case "postgres":
		if c.Database.PostgreSQL == nil {
			return fmt.Errorf("PostgreSQL configuration is required")
		}
		if err := c.validatePostgreSQL(); err != nil {
			return err
		}
	case "sqlite":
		if c.Database.SQLite == nil {
			return fmt.Errorf("SQLite configuration is required")
		}
		if err := c.validateSQLite(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported database type: %s", c.Database.Type)
	}

	if c.Security.MasterKey == "" {
		return fmt.Errorf("master key cannot be empty")
	}

	return nil
}

func (c *Config) validatePostgreSQL() error {
	pg := c.Database.PostgreSQL
	if pg.URL == "" {
		return fmt.Errorf("PostgreSQL URL cannot be empty")
	}
	parsed, err := url.Parse(pg.URL)
	if err != nil {
		return fmt.Errorf("PostgreSQL URL is invalid: %w", err)
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return fmt.Errorf("PostgreSQL URL must use postgres:// or postgresql:// scheme")
	}
	if pg.MaxConnections <= 0 {
		return fmt.Errorf("PostgreSQL max connections must be positive")
	}
	if pg.ConnectionTimeout <= 0 {
		return fmt.Errorf("PostgreSQL connection timeout must be positive")
	}
	return nil
}

func (c *Config) validateSQLite() error {
	sqlite := c.Database.SQLite
	if sqlite.Path == "" {
		return fmt.Errorf("SQLite path cannot be empty")
	}
	if sqlite.JournalMode == "" {
		return fmt.Errorf("SQLite journal mode cannot be empty")
	}
	if sqlite.Synchronous == "" {
		return fmt.Errorf("SQLite synchronous cannot be empty")
	}
	return nil
}
