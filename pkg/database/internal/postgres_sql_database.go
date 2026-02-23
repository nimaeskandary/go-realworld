package internal

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	"github.com/nimaeskandary/go-realworld/pkg/database/types"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type postgresDb struct {
	cfg db_types.SqlDbConfig
	db  *sql.DB
}

func NewPostgresSQLDatabase(cfg db_types.SqlDbConfig) (db_types.SQLDatabase, error) {
	return &postgresDb{
		cfg: cfg,
		db:  nil,
	}, nil
}

func (d *postgresDb) Start(ctx context.Context) error {
	connectionString := fmt.Sprintf("postgresql://%v:%v@%v:%v/%v?sslmode=%v",
		url.PathEscape(d.cfg.Username),
		url.PathEscape(string(d.cfg.Password)),
		d.cfg.Host,
		d.cfg.Port,
		url.PathEscape(d.cfg.DBName),
		d.cfg.SslMode)
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return fmt.Errorf("error connecting to postgres database: %w", err)
	}
	err = db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("error pinging postgres database: %w", err)
	}
	d.db = db
	return nil
}

func (d *postgresDb) Stop(_ context.Context) error {
	if d.db != nil {
		_ = d.db.Close()
		d.db = nil
	}
	return nil
}

func (d *postgresDb) GetDB() *sql.DB {
	return d.db
}

func (d *postgresDb) GetDialect() string {
	return "postgres"
}
