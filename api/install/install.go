package install

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"helm.sh/helm/v3/pkg/release"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
	"github.com/gojekfarm/albatross/pkg/logger"

	"github.com/gorilla/mux"
)

const (
	CLUSTER        string = "cluster"
	NAMESPACE      string = "namespace"
	RELEASE        string = "release_name"
	alreadyPresent string = "cannot re-use a name that is still in use"
)

// Request is the body for insatlling a release
// swagger:model installRequestBody
type Request struct {
	Name string `json:"-"`
	// example: stable/mysql
	Chart string `json:"chart"`
	// example: {"replicaCount": 1}
	Values map[string]interface{} `json:"values"`
	Flags  Flags                  `json:"flags"`
}

// Flags additional flags for installing a release
// swagger:model installFlags
type Flags struct {
	// example: false
	DryRun bool `json:"dry_run"`
	// example: 1
	Version string `json:"version"`
	flags.GlobalFlags
}

// Release wrapper for helm release
// swagger:model installRelease
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

// Response body of install response
// swagger:model installResponseBody
type Response struct {
	// Error error message, field is available only when status code is non 2xx
	Error string `json:"error,omitempty"`
	// example: deployed
	Status  string `json:"status,omitempty"`
	Data    string `json:"data,omitempty"`
	Release `json:"-"`
}

type service interface {
	Install(ctx context.Context, req Request) (Response, error)
}

// Handler handles an install request
// swagger:operation PUT /clusters/{cluster}/namespaces/{namespace}/releases/{release_name} release installOperation
//
//
// ---
// summary: Install helm release at the specified cluster and namespace
// consumes:
// - application/json
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
// - name: Body
//   in: body
//   required: true
//   schema:
//    "$ref": "#/definitions/installRequestBody"
// schemes:
// - http
// responses:
//   '200':
//    "$ref": "#/responses/installResponse"
//   '400':
//    description: Invalid request
//   '409':
//    schema:
//     $ref: "#/definitions/installResponseErrorBody"
//   '500':
//    "$ref": "#/responses/installResponse"

func Handler(service service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Errorf("[Install] error decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		values := mux.Vars(r)
		req.Flags.KubeContext = values["cluster"]
		req.Flags.Namespace = values["namespace"]
		req.Name = values["release_name"]
		resp, err := service.Install(r.Context(), req)
		if err != nil {
			code := http.StatusInternalServerError
			if err.Error() == alreadyPresent {
				code = http.StatusConflict
			}
			respondInstallError(w, "error while installing chart: %v", err, code)
			return
		}

		if err := json.NewEncoder(w).Encode(&resp); err != nil {
			respondInstallError(w, "error writing response: %v", err, http.StatusInternalServerError)
			return
		}
	})
}

// TODO: This does not handle different status codes.
func respondInstallError(w http.ResponseWriter, logprefix string, err error, statusCode int) {
	response := Response{Error: err.Error()}
	if statusCode > 0 {
		w.WriteHeader(statusCode)
	}
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		logger.Errorf("[Install] %s %v", logprefix, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
