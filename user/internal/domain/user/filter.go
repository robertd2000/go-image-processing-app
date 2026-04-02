package user

type UserFilter struct {
	Limit  int
	Offset int

	Status *UserStatus
	Search *string

	SortBy    string
	SortOrder string
}
