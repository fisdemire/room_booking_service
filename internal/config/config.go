package config

import (
	"fmt"
	"net"
	"time"

	"github.com/fisdemire/room-booking/internal/repository"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTP      HTTPConfig
	Postgres  repository.PostgresConfig
	JWTSecret string        `env:"JWT_SECRET" env-required:"true"`
	JWTTTL    time.Duration `env:"JWT_TTL" env-default:"24h"`
}

type HTTPConfig struct {
	Port            string        `env:"HTTP_PORT" env-default:"8080"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout     time.Duration `env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" env-default:"10s"`
}

func New() (Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	if cfg.HTTP.Port == "" {
		return Config{}, fmt.Errorf("HTTP_PORT must not be empty")
	}

	if cfg.HTTP.ReadTimeout <= 0 ||
		cfg.HTTP.WriteTimeout <= 0 ||
		cfg.HTTP.IdleTimeout <= 0 ||
		cfg.HTTP.ShutdownTimeout <= 0 {
		return Config{}, fmt.Errorf("HTTP timeouts must be positive")
	}

	if cfg.JWTTTL <= 0 {
		return Config{}, fmt.Errorf("JWT_TTL must be positive")
	}

	return cfg, nil
}

func (c HTTPConfig) GetAddr() string {
	return net.JoinHostPort("", c.Port)
}
