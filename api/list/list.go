package list

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
	Flags
}

type Flags struct {
	AllNamespaces bool `json:"all-namespaces,omitempty"`
	Deployed      bool `json:"deployed,omitempty"`
	Failed        bool `json:"failed,omitempty"`
	Pending       bool `json:"pending,omitempty"`
	Uninstalled   bool `json:"uninstalled,omitempty"`
	Uninstalling  bool `json:"uninstalling,omitempty"`
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
	Error    string    `json:"error,omitempty"`
	Releases []Release `json:"releases,omitempty"`
}

type service interface {
	List(ctx context.Context, req Request) (Response, error)
}

func Handler(service service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err == io.EOF || err != nil {
			logger.Errorf("[List] error decoding request: %v", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resp, err := service.List(r.Context(), req)
		if err != nil {
			respondListError(w, "error while listing charts: %v", err)
			return
		}

		if err = json.NewEncoder(w).Encode(resp); err != nil {
			respondListError(w, "error writing response: %v", err)
			return
		}
	})
}

func respondListError(w http.ResponseWriter, logprefix string, err error) {
	response := Response{Error: err.Error()}
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		logger.Errorf("[List] %s %v", logprefix, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
