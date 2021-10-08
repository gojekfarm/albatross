package status

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gojekfarm/albatross/pkg/helmcli/flags"
	"github.com/gojekfarm/albatross/pkg/logger"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"helm.sh/helm/v3/pkg/release"
)

const releaseNotFound = "release: not found"

var decoder = schema.NewDecoder()

type Request struct {
	name    string
	Version int `schema:"revision"`
	flags.GlobalFlags
}

// ErrorResponse is the body of /list
// swagger:model statusErrorResponse
type ErrorResponse struct {
	Error string `json:"error"`
}

// Release is the response of a successful status request
//swagger:model statusOkResponse
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

type service interface {
	Status(ctx context.Context, req Request) (*Release, error)
}

// Handler handles a list request
// swagger:operation GET /clusters/{cluster}/namespaces/{namespace}/releases/{release_name} release statusOperation
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
// - name: namespace
//   in: path
//   required: true
//   default: default
//   type: string
//   format: string
// - name: release_name
//   in: path
//   required: true
//   default: mysql
//   type: string
//   format: string
// - name: revision
//   in: query
//   type: number
// schemes:
// - http
// responses:
//   '200':
//    schema:
//     $ref: "#/definitions/statusOkResponse"
//   '400':
//    schema:
//     $ref: "#/definitions/statusErrorResponse"
//   '404':
//    description: Release not found
//   '500':
//    schema:
//     $ref: "#/definitions/statusErrorResponse"
func Handler(s service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var req Request
		if err := decoder.Decode(&req, r.URL.Query()); err != nil {
			logger.Errorf("[Status] error decoding request: %v", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		values := mux.Vars(r)
		req.KubeContext = values["cluster"]
		req.Namespace = values["namespace"]
		req.name = values["release_name"]
		rel, err := s.Status(r.Context(), req)
		if err != nil {
			if err.Error() == releaseNotFound {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			respondStatusError(w, "error while listing charts: %v", err, http.StatusInternalServerError)
			return
		}

		if err = json.NewEncoder(w).Encode(rel); err != nil {
			respondStatusError(w, "error writing response: %v", err, http.StatusInternalServerError)
			return
		}
	})
}

func respondStatusError(w http.ResponseWriter, msg string, err error, statusCode int) {
	response := ErrorResponse{Error: err.Error()}
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		logger.Errorf("[Status] %s %v", msg, err)
		return
	}
}
