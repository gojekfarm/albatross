package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gojekfarm/albatross/pkg/helmclient"
	"github.com/gojekfarm/albatross/pkg/logger"
)

// UpgradeResponse represents the api response for upgrade request
type UpgradeResponse struct {
	Error  string `json:"error,omitempty"`
	Status string `json:"status,omitempty"`
	Data   string `json:"data,omitempty"`
}

// Upgrade returns a http handler to handle the upgrade api request
func Upgrade() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		operation := helmclient.NewUpgradeOperation()
		if err := json.NewDecoder(r.Body).Decode(operation); err == io.EOF || err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Errorf("[Upgrade] error decoding request: %v", err)
			return
		}
		defer r.Body.Close()

		var response UpgradeResponse

		upgrader := helmclient.NewUpgrader(operation)
		result, err := upgrader.Run()
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
