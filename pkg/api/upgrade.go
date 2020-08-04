package api

import (
	"encoding/json"
	"io"
	"net/http"

	"albatross/pkg/api/logger"
)

type UpgradeRequest struct {
	Name      string                 `json:"name"`
	Namespace string                 `json:"namespace"`
	Chart     string                 `json:"chart"`
	Values    map[string]interface{} `json:"values,omitempty"`
	Flags     map[string]interface{} `json:"flags,omitempty"`
}

type UpgradeResponse struct {
	Error  string `json:"error,omitempty"`
	Status string `json:"status,omitempty"`
}

func Upgrade(svc Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var req UpgradeRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err == io.EOF || err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Errorf("[Upgrade] error decoding request: %v", err)
			return
		}
		defer r.Body.Close()

		var install bool
		var version string
		for key, value := range req.Flags {
			if key == "install" {
				install = value.(bool)
			}
			if key == "version" {
				version = value.(string)
			}
		}

		if version == "" {
			logger.Debugf("setting version to >0.0.0-0")
			version = ">0.0.0-0"
		}

		var response UpgradeResponse
		cfg := ReleaseConfig{ChartName: req.Chart, Name: req.Name, Namespace: req.Namespace, Version: version, Install: install}

		res, err := svc.Upgrade(r.Context(), cfg, req.Values)
		if err != nil {
			respondUpgradeError(w, "error while upgrading chart: %v", err)
			return
		}
		response.Status = res.Status
		if err := json.NewEncoder(w).Encode(&response); err != nil {
			respondUpgradeError(w, "error writing response: %v", err)
			return
		}
	})
}

func respondUpgradeError(w http.ResponseWriter, logprefix string, err error) {
	response := UpgradeResponse{Error: err.Error()}
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		logger.Errorf("[Upgrade] %s %v", logprefix, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
