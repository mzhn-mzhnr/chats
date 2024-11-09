package config

import (
	"fmt"
	"log/slog"
	"os"

	"mzhn/chats/pkg/prettyslog"
	"mzhn/chats/pkg/sl"

	"github.com/ilyakaznacheev/cleanenv"
)

type App struct {
	Name    string `env:"APP_NAME" env-required:"true"`
	Version string `env:"APP_VERSION" env-required:"true"`
}

type Http struct {
	Enabled bool   `env:"HTTP_ENABLED" env-required:"true" env-default:"false"`
	Host    string `env:"HTTP_HOST" env-default:"0.0.0.0"`
	Port    int    `env:"HTTP_PORT"`
	Cors    Cors
}

type Cors struct {
	AllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" env-default:"localhost:3000"`
}

type Pg struct {
	Host string `env:"PG_HOST" env-required:"true"`
	Port int    `env:"PG_PORT" env-required:"true"`
	User string `env:"PG_USER" env-required:"true"`
	Pass string `env:"PG_PASS" env-required:"true"`
	Name string `env:"PG_NAME" env-required:"true"`
}

func (pg *Pg) ConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", pg.User, pg.Pass, pg.Host, pg.Port, pg.Name)
}

type AuthApi struct {
	Host string `env:"AUTH_API_HOST" env-required:"true"`
}

type RagApi struct {
	Host string `env:"RAG_API_HOST" env-required:"true"`
}

type Redis struct {
	Host string `env:"REDIS_HOST" env-default:"localhost"`
	Port int    `env:"REDIS_PORT" env-default:"6379"`
	Pass string `env:"REDIS_PASS" env-default:""`
	DB   int    `env:"REDIS_DB" env-default:"0"`
}

type Config struct {
	Env     string `env:"ENV" env-default:"local"`
	App     App
	Http    Http
	Redis   Redis
	Pg      Pg
	AuthApi AuthApi
	RagApi  RagApi
}

func New() *Config {
	config := new(Config)

	if err := cleanenv.ReadEnv(config); err != nil {
		slog.Error("error when reading env", sl.Err(err))
		header := fmt.Sprintf("%s - %s", os.Getenv("APP_NAME"), os.Getenv("APP_VERSION"))

		usage := cleanenv.FUsage(os.Stdout, config, &header)
		usage()

		os.Exit(-1)
	}

	setupLogger(config)

	slog.Debug("config", slog.Any("c", config))
	return config
}

func setupLogger(cfg *Config) {
	var log *slog.Logger

	switch cfg.Env {
	case "prod":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case "dev":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	default:
		log = slog.New(prettyslog.NewPrettyHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	slog.SetDefault(log)
}
