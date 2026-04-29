package model

type ListImagesOutput struct {
	Items []*ImageOutput

	Total  int
	Limit  int
	Offset int
}
