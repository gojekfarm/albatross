package api

import (
	"encoding/json"
	"net/http"

	"github.com/gojekfarm/albatross/api/logger"
	"github.com/gojekfarm/albatross/pkg/helmclient"
)

type InstallRequest struct {
	Name   string
	Chart  string
	Values map[string]interface{}
	Flags  map[string]interface{}
}

type InstallResponse struct {
	Error  string `json:"error,omitempty"`
	Status string `json:"status,omitempty"`
	Data   string `json:"data,omitempty"`
}

// Install return an http handler that handles the install request
func Install() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var req InstallRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Errorf("[Install] error decoding request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		var response InstallResponse

		installer, err := helmclient.NewInstaller(req.Name, req.Chart, req.Flags)
		if err != nil {
			respondInstallError(w, "error while initializing the installer", err)
			return
		}

		result, err := installer.Run(req.Values)
		if err != nil {
			respondInstallError(w, "error while installing chart: %v", err)
			return
		}

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
