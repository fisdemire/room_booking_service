package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fisdemire/room-booking/internal/repository/postgres"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrMigrationsDirNotFound = errors.New("migrations directory not found")

type PostgresConfig struct {
	Host            string        `env:"POSTGRES_HOST" env-default:"localhost"`
	Port            string        `env:"POSTGRES_PORT" env-default:"5432"`
	Username        string        `env:"POSTGRES_USER" env-default:"room_booking"`
	Password        string        `env:"POSTGRES_PASSWORD" env-default:"local_dev_password"`
	Database        string        `env:"POSTGRES_DB" env-default:"room_booking"`
	SSLMode         string        `env:"POSTGRES_SSLMODE" env-default:"disable"`
	MigrationsPath  string        `env:"POSTGRES_MIGRATIONS_PATH" env-default:"./migrations"`
	MaxConns        int32         `env:"POSTGRES_POOL_MAX_CONNS" env-default:"25"`
	MinConns        int32         `env:"POSTGRES_MIN_CONNS" env-default:"5"`
	ConnMaxLifetime time.Duration `env:"POSTGRES_CONN_MAX_LIFETIME" env-default:"1h"`
	ConnMaxIdleTime time.Duration `env:"POSTGRES_CONN_MAX_IDLE_TIME" env-default:"30m"`
	Timeout         time.Duration `env:"POSTGRES_TIMEOUT" env-default:"5s"`
}

func (c PostgresConfig) GetConnString() (string, error) {
	if c.Host == "" || c.Username == "" || c.Database == "" {
		return "", fmt.Errorf(
			"POSTGRES_HOST, POSTGRES_USER and POSTGRES_DBNAME must not be empty",
		)
	}

	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return "", fmt.Errorf("invalid POSTGRES_PORT: %q", c.Port)
	}

	if c.Timeout <= 0 {
		return "", fmt.Errorf("POSTGRES_TIMEOUT must be positive")
	}

	switch c.SSLMode {
	case "disable", "allow", "prefer",
		"require", "verify-ca", "verify-full":
	default:
		return "", fmt.Errorf("unsupported POSTGRES_SSLMODE: %q", c.SSLMode)
	}

	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.Username, c.Password),
		Host:   net.JoinHostPort(c.Host, c.Port),
		Path:   "/" + c.Database,
	}

	query := dsn.Query()
	query.Set("sslmode", c.SSLMode)
	query.Set(
		"connect_timeout",
		strconv.Itoa(int(math.Ceil(c.Timeout.Seconds()))),
	)
	dsn.RawQuery = query.Encode()

	return dsn.String(), nil
}

func NewPool(
	ctx context.Context,
	cfg PostgresConfig,
) (*pgxpool.Pool, error) {
	if cfg.MaxConns < 1 {
		return nil, fmt.Errorf("POSTGRES_MAX_CONNS must be positive")
	}

	if cfg.MinConns < 0 || cfg.MinConns > cfg.MaxConns {
		return nil, fmt.Errorf(
			"POSTGRES_MIN_CONNS must be between 0 and POSTGRES_MAX_CONNS",
		)
	}

	if cfg.ConnMaxLifetime < 0 || cfg.ConnMaxIdleTime < 0 {
		return nil, fmt.Errorf("postgres connection lifetimes cannot be negative")
	}

	connString, err := cfg.GetConnString()
	if err != nil {
		return nil, err
	}

	poolCfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.ConnMaxLifetime
	poolCfg.MaxConnIdleTime = cfg.ConnMaxIdleTime
	poolCfg.ConnConfig.ConnectTimeout = cfg.Timeout

	poolCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(poolCtx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(poolCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

func Migrate(ctx context.Context, cfg PostgresConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	connString, err := cfg.GetConnString()
	if err != nil {
		return fmt.Errorf("build postgres connection string: %w", err)
	}

	migrationsPath, err := filepath.Abs(cfg.MigrationsPath)
	if err != nil {
		return fmt.Errorf("resolve migrations path: %w", err)
	}

	info, err := os.Stat(migrationsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf(
				"%w: %s",
				ErrMigrationsDirNotFound,
				migrationsPath,
			)
		}

		return fmt.Errorf("read migrations directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf(
			"migrations path %q is not a directory",
			migrationsPath,
		)
	}

	sourceURL := (&url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(migrationsPath),
	}).String()

	databaseURL, err := url.Parse(connString)
	if err != nil {
		return fmt.Errorf("parse postgres URL: %w", err)
	}

	query := databaseURL.Query()
	query.Set("x-multi-statement", "true")
	query.Set(
		"x-statement-timeout",
		strconv.FormatInt(cfg.Timeout.Milliseconds(), 10),
	)
	databaseURL.RawQuery = query.Encode()

	m, err := migrate.New(sourceURL, databaseURL.String())
	if err != nil {
		return fmt.Errorf("initialize migrations: %w", err)
	}

	upErr := m.Up()
	if errors.Is(upErr, migrate.ErrNoChange) {
		upErr = nil
	}

	sourceCloseErr, databaseCloseErr := m.Close()

	if upErr != nil {
		return errors.Join(
			fmt.Errorf("apply migrations: %w", upErr),
			sourceCloseErr,
			databaseCloseErr,
		)
	}

	if err := errors.Join(sourceCloseErr, databaseCloseErr); err != nil {
		return fmt.Errorf("close migration resources: %w", err)
	}

	return nil
}

func NewRepositoryWithTxManager(
	ctx context.Context,
	cfg PostgresConfig,
) (*postgres.Repositories, *postgres.TxManager, error) {
	pool, err := NewPool(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"create postgres pool: %w",
			err,
		)
	}

	if err := Migrate(ctx, cfg); err != nil {
		pool.Close()

		return nil, nil, fmt.Errorf(
			"migrate database: %w",
			err,
		)
	}

	txManager := postgres.NewTxManager(pool)
	repositories := postgres.NewWithGetter(pool, txManager.Querier)

	return repositories, txManager, nil
}
