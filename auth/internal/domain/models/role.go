package models

type Role struct {
	ID   int
	Name string
}

type UserRole struct {
	UserID int
	RoleID int
}
