package main

import (
	"fmt"
	"net/http"
	"github.com/sambhavKhanna/Infra/Logger"
)

func GetQueryParam(logger logging.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("Testing server failure")
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