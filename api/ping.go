package api

import (
	"encoding/json"
	"net/http"

	"github.com/gojekfarm/albatross/api/logger"
)

// PingResponse represents the API response for the ping request
type PingResponse struct {
	Error string `json:"error,omitempty"`
	Data  string `json:"data,omitempty"`
}

// Ping returns a http handler that handles the ping api request
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
