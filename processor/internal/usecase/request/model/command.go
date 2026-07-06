package model

import (
	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/processor/internal/domain/transformation"
)

type Command struct {
	ImageID uuid.UUID
	Source  transformation.SourceImage
	Spec    transformation.TransformSpec
}
