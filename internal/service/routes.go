package service

import (
	"fmt"
	"net/http"
	"time"
	"encoding/json"
	"github.com/sambhavKhanna/infra/logger"
)

func NewServer(
	logger logging.Logger,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, logger)

	var handler http.Handler = mux

	handler = LoggingMiddleware(handler, logger)
	return handler
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
	) {
	mux.Handle("/get", GetQueryParam(logger))
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