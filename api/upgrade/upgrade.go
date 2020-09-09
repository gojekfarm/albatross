package upgrade

import (
	"context"
	"encoding/json"
	"io"
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
	Install bool
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

// Response represents the api response for upgrade request.
type Response struct {
	Error   string `json:"error,omitempty"`
	Status  string `json:"status,omitempty"`
	Data    string `json:"data,omitempty"`
	Release `json:"-"`
}

type service interface {
	Upgrade(ctx context.Context, req Request) (Response, error)
}

func Handler(service service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err == io.EOF || err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Errorf("[Upgrade] error decoding request: %v", err)
			return
		}
		defer r.Body.Close()

		resp, err := service.Upgrade(r.Context(), req)
		if err != nil {
			respondUpgradeError(w, "error while upgrading release: %v", err)
			return
		}

		if err := json.NewEncoder(w).Encode(&resp); err != nil {
			respondUpgradeError(w, "error writing response: %v", err)
			return
		}
	})
}

func respondUpgradeError(w http.ResponseWriter, logprefix string, err error) {
	response := Response{Error: err.Error()}
	logger.Errorf("[Upgrade] %s %v", logprefix, err)
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
