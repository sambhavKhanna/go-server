package db

import (
	"database/sql"
	"fmt"
	"time"
)

type Db interface {
	Close() error
	Query(query string, args ...interface{}) (interface{}, error)
}

type PostgresDb struct {
	db *sql.DB
}

func New(config func(string) string) (PostgresDb, error) {
	// Initialize the database connection here
	if config("connection_string") == "" {
		return PostgresDb{}, fmt.Errorf("connection_string is not set")
	}

	db, err := sql.Open("postgres", config("connection_string"))
	if err != nil {
		return PostgresDb{}, err
	}
	db.SetMaxOpenConns(20)                
	db.SetMaxIdleConns(5)                 
	db.SetConnMaxIdleTime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return PostgresDb{}, fmt.Errorf("failed to ping database: %v", err)
	}

	return PostgresDb{db}, nil
}

func (pg PostgresDb) Close() error{
	if err := pg.db.Close(); err != nil {
		return err
	}
	return nil
}

func (pg PostgresDb) Query(query string, args ...interface{}) ([]interface{}, error) {
	rows, err := pg.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process the rows and return the result
	// This is just a placeholder; you would need to implement your own logic here
	var result []interface{}
	for rows.Next() {
		// Scan the row into a map or struct
		rows.Scan()
	}

	return result, nil
}

