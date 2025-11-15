package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultPath = "config.yaml"

// Duration is a YAML friendly wrapper around time.Duration.
type Duration time.Duration

// Duration returns the underlying time.Duration value.
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// UnmarshalYAML parses duration strings such as "5s".
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var raw string
	if err := value.Decode(&raw); err != nil {
		return err
	}
	if raw == "" {
		*d = Duration(0)
		return nil
	}
	dur, err := time.ParseDuration(raw)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", raw, err)
	}
	*d = Duration(dur)
	return nil
}

// ServerConfig controls the HTTP server and request level timeouts.
type ServerConfig struct {
	Port            int             `yaml:"port"`
	ReadTimeout     Duration        `yaml:"readTimeout"`
	WriteTimeout    Duration        `yaml:"writeTimeout"`
	IdleTimeout     Duration        `yaml:"idleTimeout"`
	ShutdownTimeout Duration        `yaml:"shutdownTimeout"`
	RequestTimeouts RequestTimeouts `yaml:"requestTimeouts"`
}

// MetricsConfig configures the Prometheus endpoint exposure.
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Path    string `yaml:"path"`
}

// RequestTimeouts defines fiber handler level deadlines.
type RequestTimeouts struct {
	Read   Duration `yaml:"read"`
	Write  Duration `yaml:"write"`
	Health Duration `yaml:"health"`
}

// DatabaseConfig selects which backing store to use.
type DatabaseConfig struct {
	Type     string         `yaml:"type"`
	Postgres PostgresConfig `yaml:"postgres"`
	Mongo    MongoConfig    `yaml:"mongo"`
}

// CacheConfig controls the distributed cache behaviour.
type CacheConfig struct {
	Enabled    bool        `yaml:"enabled"`
	DefaultTTL Duration    `yaml:"defaultTTL"`
	StaleTTL   Duration    `yaml:"staleTTL"`
	Redis      RedisConfig `yaml:"redis"`
}

// WriteBehindConfig toggles the write-behind queue.
type WriteBehindConfig struct {
	Enabled       bool     `yaml:"enabled"`
	BatchSize     int      `yaml:"batchSize"`
	FlushInterval Duration `yaml:"flushInterval"`
	StreamName    string   `yaml:"streamName"`
}

// RedisConfig selects the Redis instance to connect to.
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// PostgresConfig configures the SQL backend.
type PostgresConfig struct {
	Host           string   `yaml:"host"`
	Port           int      `yaml:"port"`
	User           string   `yaml:"user"`
	Password       string   `yaml:"password"`
	DB             string   `yaml:"db"`
	ConnectTimeout Duration `yaml:"connectTimeout"`
}

// DSN builds a connection string for pgxpool.
func (p PostgresConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", p.User, p.Password, p.Host, p.Port, p.DB)
}

// MongoConfig configures the MongoDB backend.
type MongoConfig struct {
	URI              string   `yaml:"uri"`
	Database         string   `yaml:"database"`
	Collection       string   `yaml:"collection"`
	ConnectTimeout   Duration `yaml:"connectTimeout"`
	OperationTimeout Duration `yaml:"operationTimeout"`
}

// Config holds strongly typed configuration for the service.
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	Cache       CacheConfig       `yaml:"cache"`
	WriteBehind WriteBehindConfig `yaml:"writeBehind"`
	Metrics     MetricsConfig     `yaml:"metrics"`
}

// Load reads the YAML file and returns a validated Config.
func Load(path string) (*Config, error) {
	if path == "" {
		path = os.Getenv("GO_CONFIG_PATH")
		if path == "" {
			path = defaultPath
		}
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := defaultConfig()
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c Config) validate() error {
	if c.Server.Port <= 0 {
		return errors.New("server.port must be greater than zero")
	}
	if c.Server.ReadTimeout.Duration() <= 0 || c.Server.WriteTimeout.Duration() <= 0 || c.Server.IdleTimeout.Duration() <= 0 {
		return errors.New("server timeouts must be greater than zero")
	}
	if c.Server.RequestTimeouts.Read.Duration() <= 0 || c.Server.RequestTimeouts.Write.Duration() <= 0 || c.Server.RequestTimeouts.Health.Duration() <= 0 {
		return errors.New("request timeouts must be greater than zero")
	}
	if c.Server.ShutdownTimeout.Duration() <= 0 {
		return errors.New("server.shutdownTimeout must be greater than zero")
	}
	if c.Database.Type != "sql" && c.Database.Type != "mongo" {
		return fmt.Errorf("database.type must be either sql or mongo, got %q", c.Database.Type)
	}
	if c.Database.Postgres.ConnectTimeout.Duration() <= 0 {
		return errors.New("database.postgres.connectTimeout must be greater than zero")
	}
	if c.Database.Mongo.ConnectTimeout.Duration() <= 0 || c.Database.Mongo.OperationTimeout.Duration() <= 0 {
		return errors.New("database.mongo timeouts must be greater than zero")
	}
	if c.Cache.Enabled {
		if c.Cache.DefaultTTL.Duration() <= 0 {
			return errors.New("cache.defaultTTL must be greater than zero when cache is enabled")
		}
		if c.Cache.StaleTTL.Duration() <= 0 {
			return errors.New("cache.staleTTL must be greater than zero when cache is enabled")
		}
		if c.Cache.Redis.Host == "" {
			return errors.New("cache.redis.host must be set when cache is enabled")
		}
		if c.Cache.Redis.Port <= 0 {
			return errors.New("cache.redis.port must be greater than zero when cache is enabled")
		}
	}
	if c.WriteBehind.Enabled {
		if c.WriteBehind.BatchSize <= 0 {
			return errors.New("writeBehind.batchSize must be greater than zero when write-behind is enabled")
		}
		if c.WriteBehind.FlushInterval.Duration() <= 0 {
			return errors.New("writeBehind.flushInterval must be greater than zero when write-behind is enabled")
		}
		if c.WriteBehind.StreamName == "" {
			return errors.New("writeBehind.streamName must be set when write-behind is enabled")
		}
	}
	if c.Metrics.Enabled {
		if c.Metrics.Port <= 0 {
			return errors.New("metrics.port must be greater than zero when metrics are enabled")
		}
		if c.Metrics.Path == "" {
			return errors.New("metrics.path must be provided when metrics are enabled")
		}
	}
	return nil
}

func defaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Port:            8081,
			ReadTimeout:     Duration(5 * time.Second),
			WriteTimeout:    Duration(5 * time.Second),
			IdleTimeout:     Duration(120 * time.Second),
			ShutdownTimeout: Duration(5 * time.Second),
			RequestTimeouts: RequestTimeouts{
				Read:   Duration(2 * time.Second),
				Write:  Duration(5 * time.Second),
				Health: Duration(2 * time.Second),
			},
		},
		Database: DatabaseConfig{
			Type: "sql",
			Postgres: PostgresConfig{
				Host:           "postgres",
				Port:           5432,
				User:           "postgres",
				Password:       "postgres",
				DB:             "productsdb",
				ConnectTimeout: Duration(5 * time.Second),
			},
			Mongo: MongoConfig{
				URI:              "mongodb://mongo:27017",
				Database:         "productsdb",
				Collection:       "products",
				ConnectTimeout:   Duration(10 * time.Second),
				OperationTimeout: Duration(5 * time.Second),
			},
		},
		Cache: CacheConfig{
			Enabled:    true,
			DefaultTTL: Duration(30 * time.Second),
			StaleTTL:   Duration(60 * time.Second),
			Redis: RedisConfig{
				Host: "redis",
				Port: 6379,
				DB:   0,
			},
		},
		WriteBehind: WriteBehindConfig{
			Enabled:       true,
			BatchSize:     50,
			FlushInterval: Duration(time.Second),
			StreamName:    "products_write_queue",
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Port:    8081,
			Path:    "/metrics",
		},
	}
}
