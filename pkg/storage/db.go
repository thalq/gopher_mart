package storage

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	logger "github.com/thalq/gopher_mart/internal/middleware"
)

var db *sql.DB

func InitDB(connectionString string) {
	var err error
	db, err = sql.Open("pgx", connectionString)
	if err != nil {
		logger.Sugar.Fatalf("Error open db: %s", err)
	}
	if err := db.Ping(); err != nil {
		logger.Sugar.Fatalf("Error ping db: %s", err)
	}
	createTable := `CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        username VARCHAR(255) UNIQUE,
        password VARCHAR(255)
    )`

	if _, err := db.Exec(createTable); err != nil {
		logger.Sugar.Fatalf("Error create table: %s", err)
	}
	logger.Sugar.Info("DB connected")
}

func GetDB() *sql.DB {
	if db == nil {
		logger.Sugar.Fatalf("Database connection is not initialized")
	}
	return db
}
