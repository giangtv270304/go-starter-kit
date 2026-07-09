package config

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/jackc/pgx/v5/tracelog"
)

type LogLevel tracelog.LogLevel

type Config struct {
	// Common settings
	GracefulShutdownPeriod time.Duration `env:"GRACEFUL_SHUTDOWN_PERIOD" envDefault:"30s"`

	// Global Log Setting
	LogLevel string `env:"LOG_LEVEL" envDefault:"debug"`

	// Metric server settings
	MetricServerHost         string        `env:"METRIC_SERVER_HOST"         envDefault:"0.0.0.0"`
	MetricServerPort         int           `env:"METRIC_SERVER_PORT"         envDefault:"8080"`
	MetricServerReadTimeout  time.Duration `env:"METRIC_SERVER_READ_TIMEOUT" envDefault:"30s"`
	MetricServerWriteTimeout time.Duration `env:"METRIC_SERVER_READ_TIMEOUT" envDefault:"30s"`

	// HTTP server settings
	HTTPServerHost         string        `env:"HTTP_SERVER_HOST"         envDefault:"0.0.0.0"`
	HTTPServerPort         int           `env:"HTTP_SERVER_PORT"         envDefault:"8081"`
	HTTPServerReadTimeout  time.Duration `env:"HTTP_SERVER_READ_TIMEOUT" envDefault:"30s"`
	HTTPServerWriteTimeout time.Duration `env:"HTTP_SERVER_READ_TIMEOUT" envDefault:"30s"`
	HTTPEnableCORS         bool          `env:"HTTP_ENABLE_CORS"          envDefault:"false"`
	HTTPAllowOrigins       []string      `env:"HTTP_ALLOW_ORIGINS"        envSeparator:"," envDefault:"*"`
	HTTPBodyLimit          string        `env:"HTTP_BODY_LIMIT"           envDefault:"100K"`
	HTTPSkipRequestID      bool          `env:"HTTP_SKIP_REQUEST_ID"      envDefault:"false"`

	// Swagger Configuration
	SwaggerHost    string `env:"SWAGGER_HOST"    envDefault:"0.0.0.0:9190"`
	SwaggerEnabled bool   `env:"SWAGGER_ENABLED" envDefault:"false"`

	// Postgres configuration
	PostgresURL                   string        `env:"POSTGRES_URL"                      envDefault:"0.0.0.0:5432"`
	PostgresMaxConnection         int32         `env:"POSTGRES_MAX_CONNECTION"           envDefault:"50"`
	PostgresMinConnection         int32         `env:"POSTGRES_MIN_CONNECTION"           envDefault:"1"`
	PostgresMaxConnectionIdleTime time.Duration `env:"POSTGRES_MAX_CONNECTION_IDLE_TIME" envDefault:"15m"`
	PostgresHealthCheckPeriod     time.Duration `env:"POSTGRES_HEALTH_CHECK_PERIOD"      envDefault:"1m"`
	PostgresLogLevel              LogLevel      `env:"POSTGRES_LOG_LEVEL"                envDefault:"4"`

	// Valkey configuration
	ValkeyHost         string        `env:"VALKEY_HOST"                 envDefault:"0.0.0.0"`
	ValkeyPort         int           `env:"VALKEY_PORT"                 envDefault:"6379"`
	ValkeyPassword     string        `env:"VALKEY_PASSWORD"`
	ValkeyDB           int           `env:"VALKEY_DB"                   envDefault:"0"`
	ValkeyMaxIdleConns int           `env:"VALKEY_MAX_IDLE_CONNS"       envDefault:"1"`
	ValkeyMinIdleConns int           `env:"VALKEY_MIN_IDLE_CONNS"       envDefault:"1"`
	ValkeyPingTimeout  time.Duration `env:"VALKEY_PING_TIMEOUT"         envDefault:"30s"`
	ValkeyDialTimeout  time.Duration `env:"VALKEY_DIAL_TIMEOUT"         envDefault:"30s"`
	ValkeyReadTimeout  time.Duration `env:"VALKEY_SERVER_READ_TIMEOUT"  envDefault:"30s"`
	ValkeyWriteTimeout time.Duration `env:"VALKEY_SERVER_WRITE_TIMEOUT" envDefault:"30s"`
	ValkeyPoolSize     int           `env:"VALKEY_POOL_SIZE"            envDefault:"1"`

	// Migration settings
	MigrationEnabled bool   `env:"MIGRATION_ENABLED" envDefault:"false"`
	MigrationSource  string `env:"MIGRATION_SOURCE"  envDefault:"file://db/migrations"`
}

func New(opts *env.Options) (*Config, error) {
	cfg := new(Config)
	if opts != nil {
		if err := env.ParseWithOptions(cfg, *opts); err != nil {
			return nil, err
		}
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
