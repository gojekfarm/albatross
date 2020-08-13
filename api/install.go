package api

import (
	"encoding/json"
	"net/http"

	"github.com/gojekfarm/albatross/api/logger"
	"github.com/gojekfarm/albatross/helmclient"
)

type InstallRequest struct {
	Name   string
	Chart  string
	Values map[string]interface{}
	Flags  map[string]interface{}
}

type InstallResponse struct {
	Error    string `json:"error,omitempty"`
	Status   string `json:"status,omitempty"`
	Manifest string `json:"manifest,omitempty"`
}

// Install return an http handler that handles the install request
// TODO: we could use interface as well if everything's in same package
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

		installer := helmclient.NewInstaller()
		installer.Setup(req.Name, req.Chart, req.Flags)
		result, err := installer.Run(req.Values)
		if err != nil {
			respondInstallError(w, "error while installing chart: %v", err)
			return
		}

		response.Status = result.Info.Status.String()
		response.Manifest = result.Manifest
		if err := json.NewEncoder(w).Encode(&response); err != nil {
			respondInstallError(w, "error writing response: %v", err)
			return
		}
	})
}

func respondInstallError(w http.ResponseWriter, logprefix string, err error) {
	response := InstallResponse{Error: err.Error()}
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		logger.Errorf("[Install] %s %v", logprefix, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
