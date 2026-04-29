package port

type Metrics interface {
	IncImageSaveSuccess()
	IncImageSaveError()
}
