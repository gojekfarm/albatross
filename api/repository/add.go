package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gojekfarm/albatross/pkg/logger"

	"github.com/gorilla/mux"
)

// AddRequest is the body for PUT request to repository
// swagger:model addRepoRequestBody
type AddRequest struct {
	Name     string `json:"-"`
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	// example: false
	ForceUpdate bool `json:"force_update"`
	// CertFile              string
	// KeyFile               string
	// CaFile                string
	InsecureSkipTLSverify bool `json:"skip_tls_verify"`
}

type addService interface {
	Add(context.Context, AddRequest) error
}

// AddErrorResponse body of non 2xx response
// swagger:model addRepoErrorResponseBody
type AddErrorResponse struct {
	Error string `json:"error"`
}

// AddOkResponse body of 2xx response
// swagger:model addRepoOkResponseBody
type AddOkResponse struct {
	Message string `json:"message"`
}

const NAME string = "repository-name"

// AddHandler handles a repo add/update request
// swagger:operation PUT /repositories/{repository_name} repository addOperation
//
// Add/Update a chart repository to the server.
// The endpoint is idempotent and a repository can be updated by using the force_update parameter to true
// ---
// produces:
// - application/json
// parameters:
// - name: repository_name
//   in: path
//   required: true
//   type: string
//   format: string
// - name: Body
//   in: body
//   required: true
//   schema:
//    "$ref": "#/definitions/addRepoRequestBody"
// schemes:
// - http
// responses:
//   '200':
//    description: "The repository was added successfully"
//    schema:
//     $ref: "#/definitions/addRepoOkResponseBody"
//   '400':
//    description: "Invalid Request"
//    schema:
//     $ref: "#/definitions/addRepoErrorResponseBody"
//   '500':
//    description: "Something went with the server"
//    schema:
//     $ref: "#/definitions/addRepoErrorResponseBody"
func AddHandler(s addService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		vars := mux.Vars(r)
		var req AddRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Errorf("[RepoAdd] error decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := req.isValid(); err != nil {
			logger.Errorf("[RepoAdd] error validating request %v", err)
			respondAddError(w, "error adding repo", err, http.StatusBadRequest)
			return
		}

		req.Name = vars[NAME]

		err := s.Add(r.Context(), req)

		if err != nil {
			logger.Errorf("[RepoAdd] error adding repo: %v", err)
			respondAddError(w, "error adding repo", err, http.StatusInternalServerError)
			return
		}
		resp := AddOkResponse{Message: fmt.Sprintf("Repo %s added successfully with url: %s", req.Name, req.URL)}
		if err := json.NewEncoder(w).Encode(&resp); err != nil {
			respondAddError(w, "error writing response: %v", err, http.StatusInternalServerError)
			return
		}
	})
}

func respondAddError(w http.ResponseWriter, logprefix string, err error, errorCode int) {
	response := AddErrorResponse{Error: err.Error()}
	w.WriteHeader(errorCode)
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		logger.Errorf("[AddRepo] %s %v", logprefix, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (req AddRequest) isValid() error {
	if req.URL == "" {
		return errors.New("url cannot be empty")
	}
	return nil
}
