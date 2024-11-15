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
	createUsersTable := `CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        username VARCHAR(255) UNIQUE,
        password VARCHAR(255)
    )`

	if _, err := db.Exec(createUsersTable); err != nil {
		logger.Sugar.Fatalf("Error create table: %s", err)
	}
	createOrdersTable := `CREATE TABLE IF NOT EXISTS orders (
		user_id INT REFERENCES users(id),
		order_id VARCHAR(255),
		status VARCHAR(10) DEFAULT 'NEW' CHECK (status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED')),
		upload_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		accrual FLOAT DEFAULT 0.0,
		withdrawal FLOAT DEFAULT 0.0,
		current FLOAT DEFAULT 0.0
	)`

	if _, err := db.Exec(createOrdersTable); err != nil {
		logger.Sugar.Fatalf("Error create orders table: %s", err)
	}

	// dropTrigger := `DROP TRIGGER IF EXISTS update_current_trigger ON orders`
	// if _, err := db.Exec(dropTrigger); err != nil {
	// 	logger.Sugar.Fatalf("Error drop existing trigger: %s", err)
	// }

	// createTriggerFunction := `CREATE OR REPLACE FUNCTION update_current()
	// RETURNS TRIGGER AS $$
	// BEGIN
	// 	NEW.current := NEW.accrual - NEW.withdrawal;
	// 	RETURN NEW;
	// END;
	// $$ LANGUAGE plpgsql;`

	// if _, err := db.Exec(createTriggerFunction); err != nil {
	// 	logger.Sugar.Fatalf("Error create trigger function: %s", err)
	// }

	// createTrigger := `CREATE TRIGGER update_current_trigger
	// BEFORE INSERT OR UPDATE ON orders
	// FOR EACH ROW EXECUTE FUNCTION update_current();`

	// if _, err := db.Exec(createTrigger); err != nil {
	// 	logger.Sugar.Fatalf("Error create trigger: %s", err)
	// }

	logger.Sugar.Info("DB connected")
}

func GetDB() *sql.DB {
	if db == nil {
		logger.Sugar.Fatalf("Database connection is not initialized")
	}
	return db
}
