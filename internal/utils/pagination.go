package utils

import (
	"fmt"
	"net/url"
	"strconv"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

type Pagination struct {
	Page   int
	Limit  int
	Offset int
	Search string
}

func ParsePagination(query url.Values) (Pagination, error) {
	page, err := positiveInt(query.Get("page"), DefaultPage)
	if err != nil {
		return Pagination{}, fmt.Errorf("page must be a positive integer")
	}
	limit, err := positiveInt(query.Get("limit"), DefaultLimit)
	if err != nil {
		return Pagination{}, fmt.Errorf("limit must be a positive integer")
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	return Pagination{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
		Search: query.Get("search"),
	}, nil
}

// ParseListPagination returns all rows unless the caller explicitly provides limit.
// When limit is present, normal page/limit defaults and validation apply.
func ParseListPagination(query url.Values) (Pagination, error) {
	if query.Get("limit") == "" {
		return Pagination{Page: DefaultPage, Limit: 0}, nil
	}
	return ParsePagination(query)
}

func PaginationMeta(total int64, pagination Pagination) map[string]int64 {
	pages := int64(1)
	limit := int64(pagination.Limit)
	if pagination.Limit == 0 {
		limit = total
	} else {
		pages = (total + int64(pagination.Limit) - 1) / int64(pagination.Limit)
	}
	return map[string]int64{"page": int64(pagination.Page), "limit": limit, "total": total, "total_pages": pages}
}

func positiveInt(value string, fallback int) (int, error) {
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("expected a positive integer")
	}
	return parsed, nil
}
