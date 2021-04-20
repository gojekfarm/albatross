package list

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"helm.sh/helm/v3/pkg/release"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
	"github.com/gojekfarm/albatross/pkg/logger"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

var decoder *schema.Decoder = schema.NewDecoder()

type Request struct {
	Flags
}

type Flags struct {
	AllNamespaces bool `schema:"-"`
	Deployed      bool `schema:"deployed"`
	Failed        bool `schema:"failed"`
	Pending       bool `schema:"pending"`
	Uninstalled   bool `schema:"uninstalled"`
	Uninstalling  bool `schema:"uninstalling"`
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
// swagger:operation GET /releases/{cluster} release listOperation
//
//
// ---
// summary: List the helm releases for the cluster
// produces:
// - application/json
// parameters:
// - name: cluster
//   in: path
//   required: true
//   default: minikube
//   type: string
//   format: string
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
//   '204':
//    description: No releases found
//   '400':
//    "$ref": "#/responses/listResponse"
//   '404':
//    "$ref": "#/responses/listResponse"
//   '500':
//    "$ref": "#/responses/listResponse"
func Handler(service service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var req Request
		if err := decoder.Decode(&req, r.URL.Query()); err != nil {
			logger.Errorf("[List] error decoding request: %v", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		values := mux.Vars(r)
		req.KubeContext = values["cluster"]
		populateRequestFlags(&req, values)
		resp, err := service.List(r.Context(), req)
		if err != nil {
			respondListError(w, "error while listing charts: %v", err)
			return
		}

		if resp.Releases == nil || len(resp.Releases) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if err = json.NewEncoder(w).Encode(resp); err != nil {
			respondListError(w, "error writing response: %v", err)
			return
		}
	})
}

// Handler handles a list request
// swagger:operation GET /releases/{cluster}/{namespace} release listOperationWithNamespace
//
//
// ---
// summary: List the helm releases for the cluster and namespace
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
//   type: string
//   format: string
//   default: default
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
//   '204':
//    description: No releases found
//   '400':
//    "$ref": "#/responses/listResponse"
//   '404':
//    "$ref": "#/responses/listResponse"
//   '500':
//    "$ref": "#/responses/listResponse"

func respondListError(w http.ResponseWriter, logprefix string, err error) {
	response := Response{Error: err.Error()}
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		logger.Errorf("[List] %s %v", logprefix, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func populateRequestFlags(req *Request, values map[string]string) {
	if values["namespace"] == "" {
		req.AllNamespaces = true
	} else {
		req.Namespace = values["namespace"]
	}
}
