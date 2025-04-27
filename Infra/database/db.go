package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Db interface {
	Close() error
	Query(query string, args ...interface{}) (interface{}, error)
}

type PostgresDb struct {
	db *sql.DB
}

func New(getenv func(string) string) (PostgresDb, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getenv("DB_HOST"),
		getenv("DB_PORT"),
		getenv("DB_USER"),
		getenv("DB_PASSWORD"),
		getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", connStr)
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

func (pg PostgresDb) Close() error {
	if err := pg.db.Close(); err != nil {
		return err
	}
	return nil
}

func (pg PostgresDb) Query(query string, args ...interface{}) (interface{}, error) {
	if isSelectQuery(query) {
		rows, err := pg.db.Query(query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		var results []map[string]interface{}

		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		for rows.Next() {
			if err := rows.Scan(valuePtrs...); err != nil {
				return nil, err
			}

			row := make(map[string]interface{})
			for i, col := range columns {
				val := values[i]
				if b, ok := val.([]byte); ok {
					row[col] = string(b)
				} else {
					row[col] = val
				}
			}
			results = append(results, row)
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}

		return results, nil
	}

	result, err := pg.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func isSelectQuery(query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	return strings.HasPrefix(query, "select")
}
