package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/gojekfarm/albatross/api"
	"github.com/gojekfarm/albatross/api/logger"
)

func main() {
	startServer()
}

func ContentTypeMiddle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func startServer() {
	router := mux.NewRouter()
	logger.Setup("debug")

	router.Handle("/ping", ContentTypeMiddle(api.Ping())).Methods(http.MethodGet)
	router.Handle("/list", ContentTypeMiddle(api.List())).Methods(http.MethodGet)
	router.Handle("/install", ContentTypeMiddle(api.Install())).Methods(http.MethodPut)
	router.Handle("/upgrade", ContentTypeMiddle(api.Upgrade())).Methods(http.MethodPost)

	err := http.ListenAndServe(fmt.Sprintf(":%d", 8080), router)
	if err != nil {
		logger.Errorf("error starting server", err)
	}
}
