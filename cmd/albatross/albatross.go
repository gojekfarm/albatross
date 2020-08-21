package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/gojekfarm/albatross/api"
	"github.com/gojekfarm/albatross/api/install"
	"github.com/gojekfarm/albatross/api/list"
	"github.com/gojekfarm/albatross/api/upgrade"
	"github.com/gojekfarm/albatross/pkg/logger"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
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

	installHandler := install.Handler(install.Service{})
	upgradeHandler := upgrade.Handler(upgrade.Service{})
	listHandler := list.Handler(list.Service{})

	router.Handle("/ping", ContentTypeMiddle(api.Ping())).Methods(http.MethodGet)
	router.Handle("/list", ContentTypeMiddle(listHandler)).Methods(http.MethodGet)
	router.Handle("/install", ContentTypeMiddle(installHandler)).Methods(http.MethodPut)
	router.Handle("/upgrade", ContentTypeMiddle(upgradeHandler)).Methods(http.MethodPost)

	err := http.ListenAndServe(fmt.Sprintf(":%d", 8080), router)
	if err != nil {
		logger.Errorf("error starting server", err)
	}
}
