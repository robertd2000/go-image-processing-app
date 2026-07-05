package transformation

type OperationType string

const (
	OperationUnknown   OperationType = ""
	OperationResize    OperationType = "resize"
	OperationCrop      OperationType = "crop"
	OperationRotate    OperationType = "rotate"
	OperationFilter    OperationType = "filter"
	OperationWatermark OperationType = "watermark"
	OperationCompress  OperationType = "compress"
	OperationFormat    OperationType = "format"
)
