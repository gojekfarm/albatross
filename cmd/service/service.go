package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/api"
	"helm.sh/helm/v3/pkg/api/logger"
	"helm.sh/helm/v3/pkg/servercontext"
)

func main() {
	servercontext.NewApp()
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

	app := servercontext.App()
	logger.Setup("debug")

	actionList := action.NewList(app.ActionConfig)
	actionInstall := action.NewInstall(app.ActionConfig)
	actionUpgrade := action.NewUpgrade(app.ActionConfig)
	actionHistory := action.NewHistory(app.ActionConfig)

	service := api.NewService(app.Config,
		new(action.ChartPathOptions),
		api.NewList(actionList),
		api.NewInstall(actionInstall),
		api.NewUpgrader(actionUpgrade),
		api.NewHistory(actionHistory))

	router.Handle("/ping", ContentTypeMiddle(api.Ping())).Methods(http.MethodGet)
	router.Handle("/list", ContentTypeMiddle(api.List(service))).Methods(http.MethodGet)
	router.Handle("/install", ContentTypeMiddle(api.Install(service))).Methods(http.MethodPut)
	router.Handle("/upgrade", ContentTypeMiddle(api.Upgrade(service))).Methods(http.MethodPost)

	err := http.ListenAndServe(fmt.Sprintf(":%d", 8080), router)
	if err != nil {
		logger.Errorf("error starting server", err)
	}
}
