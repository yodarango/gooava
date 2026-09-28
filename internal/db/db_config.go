package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/go-sql-driver/mysql"
)


type DBConfig struct {
	Conn *sql.DB
}


func DBConnection() (*sql.DB, error) {

	// Load environment variables if a .env file is present.
	_ = godotenv.Load()

	host := getenv("DB_HOST", "127.0.0.1")
	port := getenv("DB_PORT", "3306")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := getenv("DB_NAME", "gooava")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true&charset=utf8mb4&collation=utf8mb4_unicode_ci", user, pass, host, port, name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: \n %w", err)
	}

	// Wait for the MySQL container to become healthy before giving up.
	var pingErr error
	for i := 0; i < 30; i++ {
		if pingErr = db.Ping(); pingErr == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if pingErr != nil {
		return nil, fmt.Errorf("unable to ping db: \n %w", pingErr)
	}

	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("migrations failed: \n %w", err)
	}

	return db, nil

}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func NewDBConnection(dbConn * sql.DB) *DBConfig {
	return &DBConfig{
		Conn: dbConn,
	}
}