package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
)

type DbConfig struct {
	DB *sql.DB
}

const DB_MAX_LIFETIME = time.Minute * 5
const DB_MAX_CONNECTIONS = 10
const DB_MAX_IDLE = 5

// handles connection, pinging and configuration of the connection to the mysql db
func Connect() (*sql.DB, error) {

	// get the .env values
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	// connect to the db
	fmt.Println("🔌 Connecting to the database...")

	// db constants
	dbPassword := os.Getenv("DB_PASSWORD")
	dbUser := os.Getenv("DB_USER")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	dbConn, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, err
	}

	// configure connections
	dbConn.SetConnMaxIdleTime(DB_MAX_LIFETIME)
	dbConn.SetMaxOpenConns(DB_MAX_CONNECTIONS)
	dbConn.SetMaxIdleConns(DB_MAX_IDLE)

	fmt.Println("🔌🧪 Testing DB...")

	// check the db is running
	err = dbConn.Ping()
	if err != nil {
		return nil, err
	}
	fmt.Println("✅ Connection successful")

	return dbConn, nil
}
