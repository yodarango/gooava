package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)


type DBConfig struct {
	Conn *sql.DB
}


func DBConnection() (*sql.DB, error) {

	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("could not read env file \n %w", err)
	}
	// db constants
	dbPassword := os.Getenv("DB_PASSWORD")
	dbUser := os.Getenv("DB_USER")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dsn)

	// assicuarti che non avviano errori prima di continuare 
	if err != nil {
		return nil, fmt.Errorf("failed to start db: \n %w", err)
	}

	// Controlla che la DB sia funzionando 
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("unable to ping db: \n %w", err)
	}

	return db, nil

}

func NewDBConnection(dbConn * sql.DB) *DBConfig {
	return &DBConfig{
		Conn: dbConn,
	}
}