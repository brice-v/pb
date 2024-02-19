package db

import (
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var DB = sqlx.MustConnect("sqlite", "pb.db?_pragma=journal_mode(WAL)")
