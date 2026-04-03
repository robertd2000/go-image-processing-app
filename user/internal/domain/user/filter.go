package user

import (
	"errors"
	"strings"
)

var (
	ErrInvalidSortBy    = errors.New("invalid sort by")
	ErrInvalidSortOrder = errors.New("invalid sort order")
)

type UserFilter struct {
	limit  int
	offset int

	status *UserStatus
	search *string

	sortBy    string
	sortOrder string
}

func NewUserFilter(
	limit, offset int,
	status *UserStatus,
	search *string,
	sortBy, sortOrder string,
) (UserFilter, error) {

	f := UserFilter{
		limit:     limit,
		offset:    offset,
		status:    status,
		search:    search,
		sortBy:    sortBy,
		sortOrder: sortOrder,
	}

	f.normalize()

	if err := f.validate(); err != nil {
		return UserFilter{}, err
	}

	return f, nil
}

func (f *UserFilter) normalize() {
	// limit
	if f.limit <= 0 {
		f.limit = 20
	}
	if f.limit > 100 {
		f.limit = 100
	}

	// offset
	if f.offset < 0 {
		f.offset = 0
	}

	// sortBy
	if f.sortBy == "" {
		f.sortBy = "created_at"
	}

	// sortOrder
	if f.sortOrder == "" {
		f.sortOrder = "desc"
	}

	// normalize case
	f.sortOrder = strings.ToLower(f.sortOrder)
}

func (f *UserFilter) validate() error {
	switch f.sortBy {
	case "created_at", "username":
	default:
		return ErrInvalidSortBy
	}

	switch f.sortOrder {
	case "asc", "desc":
	default:
		return ErrInvalidSortOrder
	}

	return nil
}

func (f UserFilter) Limit() int {
	return f.limit
}

func (f UserFilter) Offset() int {
	return f.offset
}

func (f UserFilter) Status() *UserStatus {
	return f.status
}

func (f UserFilter) Search() *string {
	return f.search
}

func (f UserFilter) SortBy() string {
	return f.sortBy
}

func (f UserFilter) SortOrder() string {
	return f.sortOrder
}
