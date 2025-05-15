package db

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SSLConfig struct {
	Mode   string
	CAPath string
}

type ConnConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSL      *SSLConfig
}

// Connect returns a pgxpool.Pool for the given config, supporting SSL/TLS.
func Connect(ctx context.Context, cfg ConnConfig) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Database, cfg.User, cfg.Password,
		"disable",
	)
	var tlsConfig *tls.Config
	sslmode := "disable"
	if cfg.SSL != nil && cfg.SSL.Mode != "disable" {
		sslmode = cfg.SSL.Mode
		if cfg.SSL.CAPath != "" {
			caCert, err := os.ReadFile(cfg.SSL.CAPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA cert: %w", err)
			}
			roots := x509.NewCertPool()
			if !roots.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("failed to append CA cert")
			}
			tlsConfig = &tls.Config{RootCAs: roots}
		}
	}
	connStr = fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Database, cfg.User, cfg.Password, sslmode,
	)
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}
	if tlsConfig != nil {
		poolConfig.ConnConfig.TLSConfig = tlsConfig
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	return pool, nil
}

// ExecSQL executes a single SQL statement and returns the command tag or error.
func ExecSQL(ctx context.Context, pool *pgxpool.Pool, sql string) (string, error) {
	ct, err := pool.Exec(ctx, sql)
	if err != nil {
		return "", err
	}
	return ct.String(), nil
}
