package helper

import (
	"net/http"
	"strconv"
)

const (
	queryKeyPage     = "page"
	queryKeyPageSize = "page_size"

	defaultPage     = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

// GetPaginationParams extracts page/size from the request and converts them to DB-ready limit and offset.
// It handles defaults and maximums to protect the database.
// Returns (limit, offset)
func GetPaginationParams(r *http.Request) (limit, offset int) {
	pageNumber := getPage(r)
	pageSize := getPageSize(r)
	return calculateLimitAndOffset(pageNumber, pageSize)
}

func getPage(r *http.Request) int {
	page := defaultPage
	if p := r.URL.Query().Get(queryKeyPage); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			page = val
		}
	}
	return page
}

func getPageSize(r *http.Request) int {
	pageSize := defaultPageSize
	if p := r.URL.Query().Get(queryKeyPageSize); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			pageSize = val
		}
	}
	return pageSize
}

// calculateLimitAndOffset calculates database offset and limit
func calculateLimitAndOffset(pageNumber, pageSize int) (limit, offset int) {
	if pageNumber < 1 {
		pageNumber = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	offset = (pageNumber - 1) * pageSize
	limit = pageSize
	return limit, offset
}
