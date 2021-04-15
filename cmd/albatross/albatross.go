package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/gojekfarm/albatross/api"
	"github.com/gojekfarm/albatross/api/install"
	"github.com/gojekfarm/albatross/api/list"
	"github.com/gojekfarm/albatross/api/repository"
	"github.com/gojekfarm/albatross/api/uninstall"
	"github.com/gojekfarm/albatross/api/upgrade"
	"github.com/gojekfarm/albatross/pkg/helmcli"
	"github.com/gojekfarm/albatross/pkg/logger"
	_ "github.com/gojekfarm/albatross/swagger"

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
	cli := helmcli.New()

	installHandler := install.Handler(install.NewService(cli))
	upgradeHandler := upgrade.Handler(upgrade.NewService(cli))
	listHandler := list.Handler(list.NewService(cli))
	uninstallHandler := uninstall.Handler(uninstall.NewService(cli))

	router.Handle("/ping", ContentTypeMiddle(api.Ping())).Methods(http.MethodGet)
	router.Handle("/clusters/{cluster}/namespaces/{namespace}/releases/{release_name}", ContentTypeMiddle(uninstallHandler)).Methods(http.MethodDelete)
	router.Handle("/clusters/{cluster}/namespaces/{namespace}/releases/{release_name}", ContentTypeMiddle(installHandler)).Methods(http.MethodPut)
	router.Handle("/clusters/{cluster}/namespaces/{namespace}/releases/{release_name}", ContentTypeMiddle(upgradeHandler)).Methods(http.MethodPost)
	router.Handle("/clusters/{cluster}/releases", ContentTypeMiddle(listHandler)).Methods(http.MethodGet)
	router.Handle("/clusters/{cluster}/namespaces/{namespace}/releases", ContentTypeMiddle(listHandler)).Methods(http.MethodGet)

	repositorySubrouter := router.PathPrefix("/repository").Subrouter()
	handleRepositoryRoutes(repositorySubrouter)

	serveDocumentation(router)
	err := http.ListenAndServe(fmt.Sprintf(":%d", 8080), router)
	if err != nil {
		logger.Errorf("error starting server", err)
	}
}

func serveDocumentation(r *mux.Router) {
	docEnv := os.Getenv("DOCUMENTATION")
	serveDoc, err := strconv.ParseBool(docEnv)
	if err == nil && serveDoc {
		fs := http.FileServer(http.Dir("./docs"))
		r.PathPrefix("/docs/").Handler(http.StripPrefix("/docs/", fs))
	}
}

func handleRepositoryRoutes(router *mux.Router) {
	repoClient := helmcli.NewRepoClient()
	repoService := repository.NewService(repoClient)
	router.Handle(fmt.Sprintf("/{%s}", repository.NAME), ContentTypeMiddle(repository.AddHandler(repoService))).Methods(http.MethodPut)
}
