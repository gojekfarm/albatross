package api

import (
	"encoding/json"
	"net/http"

	"helm.sh/helm/v3/pkg/api/logger"
)

type PingResponse struct {
	Error string `json:"error,omitempty"`
	Data  string `json:"data,omitempty"`
}

func Ping() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		response := PingResponse{Error: "", Data: "pong"}
		if err := json.NewEncoder(w).Encode(&response); err != nil {
			respondError(w, "error writing response: %v", err)
			return
		}
	})
}

func respondError(w http.ResponseWriter, logprefix string, err error) {
	response := PingResponse{Error: err.Error()}
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		logger.Errorf("[Install] %s %v", logprefix, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
