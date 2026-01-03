package helper

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

const (
	// HeaderContentType is the HTTP header for Content-Type.
	HeaderContentType = "Content-Type"

	// ContentTypeJSON is the MIME type for JSON content.
	ContentTypeJSON = "application/json"
)

// RespondWithJSON writes a JSON response to the http.ResponseWriter with the given status code and payload.
func RespondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		// If marshalling fails, log the error and send a generic server error.
		log.Error().Err(err).Msg("Failed to marshal JSON response")
		http.Error(w, `{"error":"Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set(HeaderContentType, ContentTypeJSON)

	w.WriteHeader(statusCode)
	wSiz, err := w.Write(response)
	if err != nil {
		// This is a lower-level error, like the client disconnected. We can only log it.
		// Log network errors (broken pipe, client disconnect)
		log.Warn().Err(err).Msg("Failed to write response to client")
	} else if wSiz < len(response) {
		// Partial write, log a warning
		log.Warn().Int("written", wSiz).Int("expected", len(response)).Msg("Partial write of response to client")
	}
}

// RespondWithError writes an error message as a JSON response with the given status code.
func RespondWithError(w http.ResponseWriter, statusCode int, message string) {
	// If it's a server error (500+), log it as an error. Otherwise, it's just info/warn.
	if statusCode >= 500 {
		log.Error().Int("status", statusCode).Str("error_msg", message).Msg("Responding with Server Error")
	} else {
		// Optional: Log client errors (400s) if you want to track bad requests
		log.Debug().Int("status", statusCode).Str("error_msg", message).Msg("Responding with Client Error")
	}

	RespondWithJSON(w, statusCode, map[string]string{"error": message})
}
