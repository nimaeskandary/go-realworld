package db_types

import (
	"database/sql"

	"github.com/nimaeskandary/go-realworld/pkg/util"
)

type SQLDatabase interface {
	util.FxLifecycle
	GetDB() *sql.DB
	GetDialect() string
}

type PostgresRealWorldAppDb SQLDatabase
