package helper

import (
	"encoding/json"
	"log"
	"net/http"
)

// RespondWithJSON writes a JSON response to the http.ResponseWriter with the given status code and payload.
func RespondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		// If marshalling fails, log the error and send a generic server error.
		log.Printf("Failed to marshal JSON response: %v", err)
		http.Error(w, `{"error":"Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	wSiz, err := w.Write(response)
	if err != nil {
		// This is a lower-level error, like the client disconnected. We can only log it.
		log.Printf("Failed to write response to client (wrote %d bytes): %v", wSiz, err)
	}
}

// RespondWithError writes an error message as a JSON response with the given status code.
func RespondWithError(w http.ResponseWriter, statusCode int, message string) {
	RespondWithJSON(w, statusCode, map[string]string{"error": message})
}
