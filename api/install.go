package api

import (
	"encoding/json"
	"net/http"

	"github.com/gojekfarm/albatross/pkg/helmclient"
	"github.com/gojekfarm/albatross/pkg/logger"
)

// InstallResponse represents the API response to the install request
type InstallResponse struct {
	Error  string `json:"error,omitempty"`
	Status string `json:"status,omitempty"`
	Data   string `json:"data,omitempty"`
}

// Install return an http handler that handles the install request
func Install() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		operation := helmclient.NewInstallOperation()
		if err := json.NewDecoder(r.Body).Decode(operation); err != nil {
			logger.Errorf("[Install] error decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		installer := helmclient.NewInstaller(operation)
		result, err := installer.Run()
		if err != nil {
			respondInstallError(w, "error while installing chart: %v", err)
			return
		}

		var response InstallResponse
		response.Status = result.Status
		response.Data = result.Data
		if err := json.NewEncoder(w).Encode(&response); err != nil {
			respondInstallError(w, "error writing response: %v", err)
			return
		}
	})
}

// TODO: This does not handle different status codes.
func respondInstallError(w http.ResponseWriter, logprefix string, err error) {
	response := InstallResponse{Error: err.Error()}
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		logger.Errorf("[Install] %s %v", logprefix, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
