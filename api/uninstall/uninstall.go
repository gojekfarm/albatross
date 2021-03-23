package uninstall

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
	"github.com/gojekfarm/albatross/pkg/logger"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

var (
	errInvalidReleaseName    = errors.New("uninstall: invalid release name")
	errUnableToDecodeRequest = errors.New("unable to decode the json payload")
)

// Request Uninstall request body
// swagger:model uninstall-request model
type Request struct {
	// required: true
	ReleaseName string `json:"release_name"`

	// required: false
	DryRun bool `json:"dry_run"`

	// required: false
	KeepHistory bool `json:"keep_history"`

	// required: false
	DisableHooks bool `json:"disable_hooks"`
	flags.GlobalFlags
}

// Release contains metadata about a helm release object
// swagger:model uninstallRelease
type Release struct {
	// example: mysql
	Name string `json:"name"`
	// example: default
	Namespace string `json:"namespace"`
	// example: 1
	Version int `json:"version"`
	// example:
	Updated    time.Time      `json:"updated_at,omitempty"`
	Status     release.Status `json:"status"`
	Chart      string         `json:"chart"`
	AppVersion string         `json:"app_version"`
}

type Response struct {
	// Error error message, field is available only when status code is non 2xx
	Error string `json:"error,omitempty"`
	// Status status of the release, field is available only when status code is 2xx
	Status string `json:"status,omitempty"`
	// Release release meta data, field is available only when status code is 2xx
	Release *Release `json:"release,omitempty"`
}

type service interface {
	Uninstall(context.Context, Request) (Response, error)
}

// Handler handles an uninstall request
// swagger:route DELETE /uninstall uninstallRelease
//
// Uninstall a helm release as specified in the request
//
// consumes:
//	- application/json
// produces:
// 	- application/json
// schemes: http
// responses:
//   200: uninstallResponse
func Handler(s service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Errorf("[Uninstall] error decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			respondWithUninstallError(w, "", errUnableToDecodeRequest)
			return
		}

		if err := req.valid(); err != nil {
			logger.Errorf("[Uninstall] error in request parameters: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			respondWithUninstallError(w, "", err)
			return
		}

		resp, err := s.Uninstall(r.Context(), req)
		if err != nil {
			if errors.Is(err, driver.ErrReleaseNotFound) {
				logger.Errorf("[Uninstall] no release found for %v", req.ReleaseName)
				w.WriteHeader(http.StatusNotFound)
			} else {
				logger.Errorf("[Uninstall] unexpected error occurred: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			resp.Error = err.Error()
			err := json.NewEncoder(w).Encode(&resp)
			if err != nil {
				logger.Errorf("[Uninstall] Error writing response", err)
			}
			return
		}

		if err := json.NewEncoder(w).Encode(&resp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			respondWithUninstallError(w, "error writing response: %v", err)
			return
		}
	})
}

// preemptive checking of params to ensure correct request params, is duplicated in action.Uninstall.Run,
// but cannot fetch the type of error that's being returned since it's privately scoped.
func (req Request) valid() error {
	releaseName := req.ReleaseName
	if releaseName == "" || !action.ValidName.MatchString(releaseName) || len(releaseName) > 53 {
		return errInvalidReleaseName
	}
	return nil
}

func respondWithUninstallError(w io.Writer, logPrefix string, err error) {
	response := Response{Error: err.Error()}
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		logger.Errorf("[Uninstall] %s %v", logPrefix, err)
		return
	}
}
