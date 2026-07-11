package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/andyle182810/gframework/httpserver"
	"github.com/andyle182810/gframework/metricserver"
	"github.com/andyle182810/gframework/postgres"
	"github.com/andyle182810/gframework/runner"
	"github.com/andyle182810/gframework/valkey"
	apispec "github.com/go_starter_kit/apispec"
	"github.com/go_starter_kit/internal/config"
	"github.com/go_starter_kit/internal/repo"
	"github.com/go_starter_kit/internal/service"
	"github.com/jackc/pgx/v5/tracelog"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	echoswagger "github.com/swaggo/echo-swagger/v2"
	"resty.dev/v3"
)

const (
	serviceName = "go_starter_kit"
)

// @title			Gaian User API
// @version		    1.0
// @contact.name	VuGiang
// @contact.email	tvugiang@gmail.com
//
// @host			localhost:8081
// @BasePath		/
// @schemes		    http
func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("service", serviceName).Logger()

	if err := run(logger); err != nil {
		logger.Fatal().Err(err).Msg("Application exited with an error")
	}

	logger.Info().Msg("Application shutdown complete")
}

func run(log zerolog.Logger) error {
	cfg, err := config.New(nil)
	if err != nil {
		return err
	}

	postgresClient, err := newPostgresClient(cfg)
	if err != nil {
		return err
	}

	repo := repo.New(postgresClient)

	restyClient := newRestyClient(cfg)
	defer restyClient.Close()

	valkeyClient, err := newValkeyClient(cfg)
	if err != nil {
		return err
	}

	apiService := service.New(
		cfg,
		repo,
		restyClient,
		valkeyClient,
	)

	app := &application{
		cfg:         cfg,
		log:         log,
		service:     apiService,
		restyClient: restyClient,
		repo:        repo,
	}

	appRunner := runner.New(
		runner.WithCoreService(app.newMetricServer()),
		runner.WithCoreService(app.newHTTPServer()),
		runner.WithInfrastructureService(postgresClient),
		runner.WithInfrastructureService(valkeyClient),
	)

	appRunner.Run()

	return nil
}

type application struct {
	cfg         *config.Config
	log         zerolog.Logger
	service     *service.Service
	restyClient *resty.Client
	repo        *repo.PostgresRepo
}

func (app *application) newMetricServer() *metricserver.Server {
	metricCfg := &metricserver.Config{
		Host:         app.cfg.MetricServerHost,
		Port:         app.cfg.MetricServerPort,
		ReadTimeout:  app.cfg.MetricServerReadTimeout,
		WriteTimeout: app.cfg.MetricServerWriteTimeout,
		GracePeriod:  app.cfg.GracefulShutdownPeriod,
	}

	return metricserver.New(metricCfg)
}

func (app *application) registerRoutes(_ *echo.Echo, root *echo.Group) {
	root.GET("/health", httpserver.Wrapper(app.service.CheckHealth))

	if app.cfg.SwaggerEnabled {
		apispec.SwaggerInfo.Host = app.cfg.SwaggerHost
		apispec.SwaggerInfo.Schemes = []string{"http", "https"}

		root.GET("/swagger/*", echoswagger.WrapHandlerV3)
	}
}

func (app *application) newHTTPServer() *httpserver.Server {
	httpCfg := &httpserver.Config{
		Host:         app.cfg.HTTPServerHost,
		Port:         app.cfg.HTTPServerPort,
		EnableCors:   app.cfg.HTTPEnableCORS,
		AllowOrigins: app.cfg.HTTPAllowOrigins,
		BodyLimit:    app.cfg.HTTPBodyLimit,
		ReadTimeout:  app.cfg.HTTPServerReadTimeout,
		WriteTimeout: app.cfg.HTTPServerWriteTimeout,
		GracePeriod:  app.cfg.GracefulShutdownPeriod,
	}

	svr := httpserver.New(httpCfg)
	app.registerRoutes(svr.Echo, svr.Root)

	return svr
}

func newValkeyClient(cfg *config.Config) (*valkey.Valkey, error) {
	redisCfg := &valkey.Config{
		Host:         cfg.ValkeyHost,
		Port:         cfg.ValkeyPort,
		Password:     cfg.ValkeyPassword,
		DB:           cfg.ValkeyDB,
		DialTimeout:  cfg.ValkeyDialTimeout,
		MaxIdleConns: cfg.ValkeyMaxIdleConns,
		MinIdleConns: cfg.ValkeyMinIdleConns,
		PingTimeout:  cfg.ValkeyPingTimeout,
		PoolSize:     cfg.ValkeyPoolSize,
		ReadTimeout:  cfg.ValkeyReadTimeout,
		WriteTimeout: cfg.ValkeyWriteTimeout,
		TLSEnabled:   cfg.ValkeyTLSEnable,
	}

	valkeyClient, err := valkey.New(redisCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize valkey client: %w", err)
	}

	log.Info().Msg("Valkey client initialized successfully")

	return valkeyClient, nil
}

func newPostgresClient(cfg *config.Config) (*postgres.Postgres, error) {
	if cfg.MigrationEnabled {
		log.Info().Str("source", cfg.MigrationSource).Msg("Starting database migration process...")

		if err := postgres.MigrateUp(cfg.PostgresURL, cfg.MigrationSource); err != nil {
			return nil, fmt.Errorf("postgresql migration failed: %w", err)
		}

		log.Info().Msg("Database migration process completed successfully")
	}

	pgCfg := &postgres.Config{
		URL:                   cfg.PostgresURL,
		MaxConnection:         cfg.PostgresMaxConnection,
		MinConnection:         cfg.PostgresMinConnection,
		MaxConnectionIdleTime: cfg.PostgresMaxConnectionIdleTime,
		HealthCheckPeriod:     cfg.PostgresHealthCheckPeriod,
		LogLevel:              tracelog.LogLevel(cfg.PostgresLogLevel),
	}

	postgresDB, err := postgres.New(pgCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres client: %w", err)
	}

	log.Info().Msg("PostgreSQL client initialized successfully")

	return postgresDB, nil
}

func newRestyClient(cfg *config.Config) *resty.Client {
	const defaultRetryCount int = 3

	client := resty.New().
		SetTimeout(cfg.HTTPServerWriteTimeout).
		SetRetryCount(defaultRetryCount).
		SetDebug(strings.EqualFold(cfg.LogLevel, "debug")).
		SetHeader(echo.HeaderAccept, echo.MIMEApplicationJSON).
		SetHeader(echo.HeaderContentType, echo.MIMEApplicationJSON)

	return client
}
