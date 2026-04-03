package model

type UserFilterInput struct {
	Limit     int
	Offset    int
	Search    string
	SortBy    string
	SortOrder string
}
