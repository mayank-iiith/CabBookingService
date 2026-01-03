package helper

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondWithJSON(t *testing.T) {
	// 1. Setup: Create a "fake" response writer
	rr := httptest.NewRecorder()

	payload := map[string]string{"message": "success"}
	code := http.StatusOK

	// 2. Call the function
	RespondWithJSON(rr, code, payload)

	// 3. Assertions
	if status := rr.Code; status != code {
		t.Errorf("handler returned wrong status code: got %v want %v", status, code)
	}

	// Check the content type header
	expectedHeader := "application/json"
	if rr.Header().Get("Content-Type") != expectedHeader {
		t.Errorf("handler returned wrong content-type: got %v want %v",
			rr.Header().Get("Content-Type"), expectedHeader)
	}

	// Check the response body
	var target map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &target)
	if err != nil {
		t.Fatalf("could not unmarshal response body: %v", err)
	}

	if target["message"] != "success" {
		t.Errorf("handler returned wrong body: got %v want %v",
			target["message"], "success")
	}
}

// TestRespondWithError tests the RespondWithError helper
func TestRespondWithError(t *testing.T) {
	// 1. Setup: Create a "fake" response writer
	rr := httptest.NewRecorder()

	errorMessage := "something went wrong"
	code := http.StatusInternalServerError

	// 2. Call the function
	RespondWithError(rr, code, errorMessage)

	// 3. Assertions
	if status := rr.Code; status != code {
		t.Errorf("handler returned wrong status code: got %v want %v", status, code)
	}

	// Check the content type header
	expectedHeader := "application/json"
	if rr.Header().Get("Content-Type") != expectedHeader {
		t.Errorf("handler returned wrong content-type: got %v want %v",
			rr.Header().Get("Content-Type"), expectedHeader)
	}

	// Check the response body
	var target map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &target)
	if err != nil {
		t.Fatalf("could not unmarshal response body: %v", err)
	}

	if target["error"] != errorMessage {
		t.Errorf("handler returned wrong body: got %v want %v",
			target["error"], errorMessage)
	}
}
