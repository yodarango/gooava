package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
)

type SQLConfig struct {
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
	dbConn, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("MYSQL_DATABASE")))

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
