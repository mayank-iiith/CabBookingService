package util

const (
	// DefaultPageSize is the default number of items per page
	DefaultPageSize = 10
	// MaxPageSize is the maximum number of items per page
	MaxPageSize = 100
)

func GetPageOffsetAndLimit(pageNumber, pageSize int) (int, int) {
	if pageNumber < 1 {
		pageNumber = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	offset := (pageNumber - 1) * pageSize
	return offset, pageSize
}
