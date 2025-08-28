package database

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"th-application-technical-assignment/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type PgPool struct {
	pool *pgxpool.Pool
}

func NewPgPoolFromCfg(ctx context.Context, cfg *Config) (*PgPool, error) {
	connStr, err := connString(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "conn string")
	}

	poolCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, errors.Wrap(err, "parse config")
	}

	poolCfg.MaxConns = cfg.PoolMaxConns

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, errors.Wrap(err, "new pool")
	}

    slog.DebugContext(ctx, "created pg pool", "pool", pool)

	return &PgPool{pool: pool}, nil
}

func (p *PgPool) DBTX() sqlc.DBTX {
	return p.pool
}

func (p *PgPool) Close(ctx context.Context) {
	slog.InfoContext(ctx, "closing pg pool")
	p.pool.Close()
}

func connString(ctx context.Context, cfg *Config) (string, error) {
	if cfg == nil {
		return "", errors.New("config is nil")
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Path:   cfg.Name,
	}

	q := u.Query()
	q.Set("connect_timeout", fmt.Sprintf("%d", cfg.ConnectionTimeout))
	q.Set("sslmode", cfg.SSLMode)

	if c := cfg.SSLCertPath; c != "" {
		q.Set("sslcert", c)
	}
	if c := cfg.SSLKeyPath; c != "" {
		q.Set("sslkey", c)
	}
	if c := cfg.SSLRootCertPath; c != "" {
		q.Set("sslrootcert", c)
	}

	u.RawQuery = q.Encode()
	slog.DebugContext(ctx, "connection string", "connStr", u.String())
	return u.String(), nil
}
