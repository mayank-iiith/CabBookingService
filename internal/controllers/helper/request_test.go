package helper

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPaginationParams(t *testing.T) {
	t.Parallel() // 1. Mark the main test as parallel

	tests := []struct {
		name           string
		queryParams    string // e.g. "?page=2&page_size=20"
		expectedLimit  int
		expectedOffset int
	}{
		{
			name:           "Default values (empty params)",
			queryParams:    "",
			expectedLimit:  10, // defaultPageSize
			expectedOffset: 0,
		},
		{
			name:           "Valid page and size",
			queryParams:    "?page=2&page_size=20",
			expectedLimit:  20,
			expectedOffset: 20, // (2-1) * 20
		},
		{
			name:           "Page 1",
			queryParams:    "?page=1&page_size=10",
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "Invalid page (string)",
			queryParams:    "?page=abc&page_size=10",
			expectedLimit:  10,
			expectedOffset: 0, // Defaults to page 1
		},
		{
			name:           "Invalid page (zero)",
			queryParams:    "?page=0&page_size=10",
			expectedLimit:  10,
			expectedOffset: 0, // Defaults to page 1
		},
		{
			name:           "Invalid page (negative)",
			queryParams:    "?page=-5&page_size=10",
			expectedLimit:  10,
			expectedOffset: 0, // Defaults to page 1
		},
		{
			name:           "Invalid page_size (string)",
			queryParams:    "?page=1&page_size=xyz",
			expectedLimit:  10, // Defaults to defaultPageSize
			expectedOffset: 0,
		},
		{
			name:           "Page_size exceeds max",
			queryParams:    "?page=1&page_size=1000",
			expectedLimit:  100, // Capped at maxPageSize (100)
			expectedOffset: 0,
		},
		{
			name:           "Complex calculation",
			queryParams:    "?page=3&page_size=15",
			expectedLimit:  15,
			expectedOffset: 30, // (3-1) * 15 = 30
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // 2. Mark the sub-test as parallel

			// Create a dummy request with the query params
			req := httptest.NewRequest(http.MethodGet, "/"+tt.queryParams, nil)

			limit, offset := GetPaginationParams(req)
			require.Equal(t, tt.expectedLimit, limit, "Limit mismatch")
			require.Equal(t, tt.expectedOffset, offset, "Offset mismatch")
		})
	}
}

func TestCalculateLimitAndOffset(t *testing.T) {
	t.Parallel() // Mark as parallel

	// Direct unit test for the math logic
	tests := []struct {
		name           string
		pageNumber     int
		pageSize       int
		expectedLimit  int
		expectedOffset int
	}{
		{"Normal case", 2, 20, 20, 20},
		{"Page 1", 1, 10, 10, 0},
		{"Zero page (defaults to 1)", 0, 10, 10, 0},
		{"Negative page (defaults to 1)", -1, 10, 10, 0},
		{"Zero size (defaults to 10)", 1, 0, 10, 0},
		{"Negative size (defaults to 10)", 1, -5, 10, 0},
		{"Size too big (caps at 100)", 1, 150, 100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Mark sub-test as parallel
			limit, offset := calculateLimitAndOffset(tt.pageNumber, tt.pageSize)
			require.Equal(t, tt.expectedLimit, limit, "Limit mismatch")
			require.Equal(t, tt.expectedOffset, offset, "Offset mismatch")
		})
	}
}
