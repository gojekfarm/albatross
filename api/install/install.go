package install

import (
	"encoding/json"
	"net/http"
	"time"

	"helm.sh/helm/v3/pkg/release"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
	"github.com/gojekfarm/albatross/pkg/logger"
)

type Request struct {
	Name   string
	Chart  string
	Values map[string]interface{}
	Flags  Flags
}

type Flags struct {
	DryRun  bool `json:"dry_run"`
	Version string
	flags.GlobalFlags
}

type Release struct {
	Name       string         `json:"name"`
	Namespace  string         `json:"namespace"`
	Version    int            `json:"version"`
	Updated    time.Time      `json:"updated_at,omitempty"`
	Status     release.Status `json:"status"`
	Chart      string         `json:"chart"`
	AppVersion string         `json:"app_version"`
}

type Response struct {
	Error   string `json:"error,omitempty"`
	Status  string `json:"status,omitempty"`
	Data    string `json:"data,omitempty"`
	Release `json:"-"`
}

func Handler(service service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Errorf("[Install] error decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resp, err := service.Install(r.Context(), req)
		if err != nil {
			respondInstallError(w, "error while installing chart: %v", err)
			return
		}

		if err := json.NewEncoder(w).Encode(&resp); err != nil {
			respondInstallError(w, "error writing response: %v", err)
			return
		}
	})
}

// TODO: This does not handle different status codes.
func respondInstallError(w http.ResponseWriter, logprefix string, err error) {
	response := Response{Error: err.Error()}
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		logger.Errorf("[Install] %s %v", logprefix, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
