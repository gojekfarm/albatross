package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gojekfarm/albatross/api/logger"
	"github.com/gojekfarm/albatross/pkg/helmclient"
)

type UpgradeRequest struct {
	Name   string                 `json:"name"`
	Chart  string                 `json:"chart"`
	Values map[string]interface{} `json:"values,omitempty"`
	Flags  map[string]interface{} `json:"flags,omitempty"`
}

type UpgradeResponse struct {
	Error  string `json:"error,omitempty"`
	Status string `json:"status,omitempty"`
	Data   string `json:"data,omitempty"`
}

func Upgrade() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var req UpgradeRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err == io.EOF || err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Errorf("[Upgrade] error decoding request: %v", err)
			return
		}
		defer r.Body.Close()

		var response UpgradeResponse
		upgrader, err := helmclient.NewUpgrader(req.Name, req.Chart, req.Flags)
		if err != nil {
			respondUpgradeError(w, "error while initializing the upgrader", err)
			return
		}

		result, err := upgrader.Run(req.Values)
		if err != nil {
			respondUpgradeError(w, "error while upgrading release: %v", err)
			return
		}

		response.Status = result.Status
		response.Data = result.Data
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
