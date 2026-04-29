package model

import "github.com/google/uuid"

type ListImagesInput struct {
	UserID uuid.UUID

	Limit  int
	Offset int

	// optional
	IncludeDeleted bool
}
