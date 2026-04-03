package model

type ListUsersRequest struct {
	Limit  int
	Offset int
	Search string
}
