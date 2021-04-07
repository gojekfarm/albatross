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

	"github.com/gorilla/mux"
)

// Request is the body for upgrading a release
// swagger:model upgradeRequestBody
type Request struct {
	// example: mysql
	Name string `json:"name"`
	// example: stable/mysql
	Chart string `json:"chart"`
	// example: {"replicaCount": 1}
	Values map[string]interface{} `json:"values"`
	// Deprecated field
	// example: {"kube_context": "minikube", "namespace":"default"}
	Flags Flags `json:"flags"`
}

// Flags additional flags supported while upgrading a release
// swagger:model upgradeFlags
type Flags struct {
	// example: false
	DryRun bool `json:"dry_run"`
	// example: 1
	Version string `json:"version"`
	// example: true
	Install bool `json:"install"`
	flags.GlobalFlags
}

// Release wrapper for helm release
// swagger:model upgradeReleaseBody
type Release struct {
	// example: mysql-5.7
	Name string `json:"name"`
	// example: default
	Namespace string `json:"namespace"`
	// example: 2
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

// Response represents the api response for upgrade request.
type Response struct {
	// Error field is available only when the response status code is non 2xx
	Error string `json:"error,omitempty"`
	// example: deployed
	Status  string `json:"status,omitempty"`
	Data    string `json:"data,omitempty"`
	Release `json:"-"`
}

type service interface {
	Upgrade(ctx context.Context, req Request) (Response, error)
}

// Handler handles an upgrade request
// swagger:route POST /upgrade upgradeRelease
//
// Upgrades a helm release as specified in the request
//
// Deprecated: true
//
// consumes:
//	- application/json
// produces:
// 	- application/json
// schemes: http
// responses:
//   200: upgradeResponse
//   400: upgradeResponse
//   500: upgradeResponse
func Handler(service service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err == io.EOF || err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Errorf("[Upgrade] error decoding request: %v", err)
			return
		}

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

// RestHandler handles an upgrade request
// swagger:operation POST /releases/{kube_context}/{namespace}/{release_name} release upgradeOperation
//
// Install helm release at the specified location
//
// ---
// produces:
// - application/json
// parameters:
// - name: kube_context
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
// - name: Body
//   in: body
//   required: true
//   schema:
//    "$ref": "#/definitions/upgradeRequestBody"
// schemes:
// - http
// responses:
//   '200':
//    "$ref": "#/responses/listResponse"
//   '400':
//    "$ref": "#/responses/listResponse"
//   '404':
//    "$ref": "#/responses/listResponse"
//   '500':
//    "$ref": "#/responses/listResponse"
func RestHandler(service service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err == io.EOF || err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Errorf("[Upgrade] error decoding request: %v", err)
			return
		}
		values := mux.Vars(r)
		req.Flags.KubeContext = values["kube_context"]
		req.Flags.Namespace = values["namespace"]
		req.Name = values["release_name"]
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
