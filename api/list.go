package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gojekfarm/albatross/pkg/helmclient"
	"github.com/gojekfarm/albatross/pkg/logger"
)

// ListResponse represents the API response for the list request
type ListResponse struct {
	Error    string                `json:"error,omitempty"`
	Releases []*helmclient.Release `json:"releases,omitempty"`
}

// List return a http handler to handle the list request
func List() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response ListResponse
		operation := helmclient.NewListOperation()
		if err := json.NewDecoder(r.Body).Decode(operation); err == io.EOF || err != nil {
			logger.Errorf("[List] error decoding request: %v", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			response.Error = err.Error()
			json.NewEncoder(w).Encode(response)
			return
		}
		defer r.Body.Close()

		lister := helmclient.NewLister(operation)
		listResult, err := lister.Run()
		if err != nil {
			respondListError(w, "error while listing charts: %v", err)
		}

		response = ListResponse{"", listResult.Releases}
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			respondListError(w, "error writing response: %v", err)
			return
		}
	})
}

func respondListError(w http.ResponseWriter, logprefix string, err error) {
	response := ListResponse{Error: err.Error()}
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		logger.Errorf("[List] %s %v", logprefix, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
