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

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

var (
	errInvalidReleaseName    = errors.New("uninstall: invalid release name")
	errUnableToDecodeRequest = errors.New("unable to decode the json payload")
	decoder                  = schema.NewDecoder()
)

type Request struct {
	releaseName  string
	DryRun       bool `json:"dry_run" schema:"dry_run"`
	KeepHistory  bool `json:"keep_history" schema:"keep_history"`
	DisableHooks bool `json:"disable_hooks" schema:"disable_hooks"`
	Timeout      int  `json:"timeout" schema:"timeout"`
	flags.GlobalFlags
}

// Release contains metadata about a helm release object
// swagger:model uninstallRelease
type Release struct {
	// example: mysql-5.7
	Name string `json:"name"`
	// example: default
	Namespace string `json:"namespace"`
	// example: 1
	Version int `json:"version"`
	// example: 2021-03-24T12:24:18.450869+05:30
	Updated time.Time `json:"updated_at,omitempty"`
	// example: deployed
	Status release.Status `json:"status"`
	// example: mysql
	Chart string `json:"chart"`
	// example: 5.7.30
	AppVersion string `json:"app_version"`
}

// Response is the body of uninstall route
// swagger:model uninstallResponseBody
type Response struct {
	// Error error message, field is available only when status code is non 2xx
	Error string `json:"error,omitempty"`
	// Status status of the release, field is available only when status code is 2xx
	// example: uninstalled
	Status string `json:"status,omitempty"`
	// Release release meta data, field is available only when status code is 2xx
	Release *Release `json:"release,omitempty"`
}

type service interface {
	Uninstall(context.Context, Request) (Response, error)
}

// Handler handles an uninstall request
// swagger:operation DELETE /clusters/{cluster}/namespaces/{namespace}/releases/{release_name} release uninstallOperation
//
//
// ---
// summary: Uninstall a helm release
// produces:
// - application/json
// parameters:
// - name: cluster
//   in: path
//   required: true
//   default: minikube
//   type: string
//   format: string
// - name: namespace
//   in: path
//   required: true
//   default: default
//   type: string
//   format: string
// - name: release_name
//   in: path
//   required: true
//   type: string
//   format: string
//   default: mysql-final
// - name: dry_run
//   in: query
//   type: boolean
//   default: false
// - name: keep_history
//   in: query
//   type: boolean
//   default: false
// - name: disable_hooks
//   in: query
//   type: boolean
//   default: false
// - name: timeout
//   in: query
//   type: integer
//   default: 300
// schemes:
// - http
// responses:
//   '200':
//    "$ref": "#/responses/uninstallResponse"
//   '400':
//    schema:
//     $ref: "#/definitions/uninstallErrorResponse"
//   '404':
//    schema:
//     $ref: "#/definitions/uninstallErrorResponse"
//   '500':
//    "$ref": "#/responses/uninstallResponse"
func Handler(s service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req Request
		if err := decoder.Decode(&req, r.URL.Query()); err != nil {
			logger.Errorf("[Uninstall] error decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			respondWithUninstallError(w, "", errUnableToDecodeRequest)
			return
		}
		values := mux.Vars(r)
		req.releaseName = values["release_name"]
		req.GlobalFlags.KubeContext = values["cluster"]
		req.GlobalFlags.Namespace = values["namespace"]
		if err := req.valid(); err != nil {
			logger.Errorf("[Uninstall] error in request parameters: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			respondWithUninstallError(w, "", err)
			return
		}

		resp, err := s.Uninstall(r.Context(), req)
		if err != nil {
			if errors.Is(err, driver.ErrReleaseNotFound) {
				logger.Errorf("[Uninstall] no release found for %v", req.releaseName)
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
	releaseName := req.releaseName
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
