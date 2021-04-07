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

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

var decoder *schema.Decoder = schema.NewDecoder()

// Request is body of List Route
// swagger:model listRequestBody
type Request struct {
	Flags
}

// Flags contains all the params supported
// swagger:model listRequestFlags
type Flags struct {
	// example: false
	// required: false
	AllNamespaces bool `json:"all-namespaces,omitempty" schema:"all_namespaces"`
	// required: false
	// example: false
	Deployed bool `json:"deployed,omitempty" schema:"deployed"`
	// required: false
	// example: false
	Failed bool `json:"failed,omitempty" schema:"failed"`
	// required: false
	// example: false
	Pending bool `json:"pending,omitempty" schema:"pending"`
	// required: false
	// example: false
	Uninstalled bool `json:"uninstalled,omitempty" schema:"uninstalled"`
	// required: false
	// example: false
	Uninstalling bool `json:"uninstalling,omitempty" schema:"uninstalling"`
	flags.GlobalFlags
}

// Release wraps a helm release
// swagger:model listRelease
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

// Response is the body of /list
// swagger:model listReponseBody
type Response struct {
	// Error field is available only when the response status code is non 2xx
	Error    string    `json:"error,omitempty"`
	Releases []Release `json:"releases,omitempty"`
}

type service interface {
	List(ctx context.Context, req Request) (Response, error)
}

// Handler handles a list request
// swagger:route GET /list listRelease
//
// List helm releases as specified in the request
//
// Deprecated: true
//
// consumes:
//	- application/json
// produces:
// 	- application/json
// schemes: http
// responses:
//   200: listResponse
//   400: listResponse
//   500: listResponse
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

// RestHandler handles an uninstall request
// swagger:operation GET /releases/{kube_context} release listOperation
//
// List helm releases in the kubecontext as specified by query params
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
//   in: query
//   required: false
//   type: string
//   format: string
// - name: all_namespaces
//   in: query
//   type: boolean
//   default: true
// - name: deployed
//   in: query
//   type: boolean
//   default: false
// - name: uninstalled
//   in: query
//   type: boolean
//   default: false
// - name: failed
//   in: query
//   type: boolean
//   default: false
// - name: pending
//   in: query
//   type: boolean
//   default: false
// - name: uninstalling
//   in: query
//   type: boolean
//   default: false
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
		var req Request
		if err := decoder.Decode(&req, r.URL.Query()); err != nil {
			logger.Errorf("[List] error decoding request: %v", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		values := mux.Vars(r)
		req.KubeContext = values["kube_context"]
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
