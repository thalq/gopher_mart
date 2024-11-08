package storage

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	logger "github.com/thalq/gopher_mart/internal/middleware"
)

var db *sql.DB

func InitDB(connectionString string) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		logger.Sugar.Fatalf("Error open db: %s", err)
	}
	if err := db.Ping(); err != nil {
		logger.Sugar.Fatalf("Error ping db: %s", err)
	}
	creaateTable := `CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE,
		password VARCHAR(255)
	)`

	if _, err := db.Exec(creaateTable); err != nil {
		logger.Sugar.Fatalf("Error create table: %s", err)
	}
	logger.Sugar.Info("DB connected")
}

func GetDB() *sql.DB {
	return db
}
