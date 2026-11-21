package database

import (
	"context"
	"net"
	"net/url"
	"time"

	pgxuuid "github.com/jackc/pgx-gofrs-uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rizwank123/chatApp/internal/pkg/config"
)

// NewDB returns a new database connection pool
func NewDB(cfg config.ChatApiConfig) (*pgxpool.Pool, error) {
	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.DatabaseUsername, cfg.DatabasePassword),
		Host:   net.JoinHostPort(cfg.DatabaseHost, cfg.DatabasePort),
		Path:   cfg.DatabaseName,
	}
	query := dsn.Query()
	query.Set("sslmode", cfg.DatabaseSSLMode)
	if cfg.DatabaseSSLRootCert != "" {
		query.Set("sslrootcert", cfg.DatabaseSSLRootCert)
	}
	dsn.RawQuery = query.Encode()

	dbConfig, err := pgxpool.ParseConfig(dsn.String())
	if err != nil {
		return nil, err
	}
	dbConfig.MaxConns = 30
	dbConfig.MinConns = 2
	dbConfig.MaxConnLifetime = time.Hour
	dbConfig.MaxConnIdleTime = 15 * time.Minute

	// Register the uuid type
	dbConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxuuid.Register(conn.TypeMap())
		return nil
	}
	// Create the connection pool
	dbPool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := dbPool.Ping(ctx); err != nil {
		dbPool.Close()
		return nil, err
	}
	return dbPool, nil
}
