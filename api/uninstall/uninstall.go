package uninstall

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
	"github.com/gojekfarm/albatross/pkg/logger"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

var errInvalidReleaseName = errors.New("uninstall: invalid release name")

// Request encapsulates an Http Request.
type Request struct {
	ReleaseName  string `json:"release_name"`
	Dryrun       bool   `json:"dry_run,omitempty"`
	KeepHistory  bool   `json:"keep_history"`
	DisableHooks bool   `json:"disable_hooks"`
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
	Error   string  `json:"error,omitempty"`
	Status  string  `json:"status"`
	Release Release `json:"release"`
}

type service interface {
	Uninstall(context.Context, Request) (Response, error)
}

// Handler creates a handler function to respond to delete requests.
func Handler(s service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Errorf("[Uninstall] error decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := checkMandatoryParams(req); err != nil {
			logger.Errorf("[Uninstall] error in request parameters: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resp, err := s.Uninstall(r.Context(), req)
		if err != nil {
			if errors.Is(err, driver.ErrReleaseNotFound) {
				logger.Errorf("[Uninstall] no release found for %v", req.ReleaseName)
				w.WriteHeader(http.StatusBadRequest)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}

			respondWithUninstallError(w, "error while uninstalling release: %v", err)
			return
		}

		if err := json.NewEncoder(w).Encode(&resp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			respondWithUninstallError(w, "error writing response: %v", err)
			return
		}
	})
}

func respondWithUninstallError(w http.ResponseWriter, logPrefix string, err error) {
	response := Response{Error: err.Error()}
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		logger.Errorf("[Uninstall] %s %v", logPrefix, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// preemptive checking of params to ensure correct request params, is duplicated in action.Uninstall.Run,
// but cannot fetch the type of error that's being returned since it's privately scoped.
func checkMandatoryParams(req Request) error {
	releaseName := req.ReleaseName
	if releaseName == "" || !action.ValidName.MatchString(releaseName) || len(releaseName) > 53 {
		return errInvalidReleaseName
	}
	return nil
}
