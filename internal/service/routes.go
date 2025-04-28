package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sambhavKhanna/infra/database"
	"github.com/sambhavKhanna/infra/logger"
)

func NewServer(
	logger logging.Logger,
	db db.Db,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, logger, db)

	var handler http.Handler = mux

	handler = LoggingMiddleware(handler, logger)
	return handler
}

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func PostUser(logger logging.Logger, db db.Db) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		user, err := decode[User](r)
		if err != nil {
			logger.Error("Failed to decode user: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		query := `INSERT INTO users (name, email) VALUES ($1, $2)`
		_, err = db.Query(query, user.Name, user.Email)
		if err != nil {
			logger.Error("Failed to insert user: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}

func GetUsers(logger logging.Logger, db db.Db) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		query := `SELECT name, email FROM users`
		results, err := db.Query(query)
		if err != nil {
			logger.Error("Failed to query users: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resultsList, ok := results.([]map[string]interface{})
		if !ok {
			logger.Error("Failed to parse query results")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		users := make([]User, 0, len(resultsList))
		for _, result := range resultsList {
			user := User{
				Name:  result["name"].(string),
				Email: result["email"].(string),
			}
			users = append(users, user)
		}

		if err := encode(w, r, http.StatusOK, users); err != nil {
			logger.Error("Failed to encode users: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

func GetQueryParam(logger logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("/get")
		query := r.URL.Query()
		for key, value := range query {
			for _, v := range value {
				fmt.Printf("%s = %s\n", key, v)
			}
		}
	})
}

func addRoutes(
	mux *http.ServeMux,
	logger logging.Logger,
	db db.Db,
) {
	mux.Handle("/get", GetQueryParam(logger))
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			PostUser(logger, db).ServeHTTP(w, r)
		case http.MethodGet:
			GetUsers(logger, db).ServeHTTP(w, r)
		default:
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		}
	})
}

func LoggingMiddleware(next http.Handler, logger logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.Info("Started %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		logger.Info("Completed %s in %v", r.URL.Path, duration)
	})
}

func encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}

	return v, nil
}
