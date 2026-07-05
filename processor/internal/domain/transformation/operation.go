package transformation

type Operation struct {
	Resize    *ResizeParameters    `json:"resize,omitempty"`
	Crop      *CropParameters      `json:"crop,omitempty"`
	Rotate    *RotateParameters    `json:"rotate,omitempty"`
	Filter    *FilterParameters    `json:"filter,omitempty"`
	Watermark *WatermarkParameters `json:"watermark,omitempty"`
	Compress  *CompressParameters  `json:"compress,omitempty"`
	Format    *FormatParameters    `json:"format,omitempty"`
}

func (o Operation) Validate() error {
	count := 0

	if o.Resize != nil {
		count++
		if err := o.Resize.Validate(); err != nil {
			return err
		}
	}

	if o.Crop != nil {
		count++
		if err := o.Crop.Validate(); err != nil {
			return err
		}
	}

	if o.Rotate != nil {
		count++
		if err := o.Rotate.Validate(); err != nil {
			return err
		}
	}

	if o.Filter != nil {
		count++
		if err := o.Filter.Validate(); err != nil {
			return err
		}
	}

	if o.Watermark != nil {
		count++
		if err := o.Watermark.Validate(); err != nil {
			return err
		}
	}

	if o.Compress != nil {
		count++
		if err := o.Compress.Validate(); err != nil {
			return err
		}
	}

	if o.Format != nil {
		count++
		if err := o.Format.Validate(); err != nil {
			return err
		}
	}

	if count == 0 {
		return ErrOperationIsEmpty
	}

	if count > 1 {
		return ErrMultipleOperations
	}

	return nil
}

func (o Operation) Type() OperationType {
	switch {
	case o.Resize != nil:
		return OperationResize

	case o.Crop != nil:
		return OperationCrop

	case o.Rotate != nil:
		return OperationRotate

	case o.Filter != nil:
		return OperationFilter

	case o.Watermark != nil:
		return OperationWatermark

	case o.Compress != nil:
		return OperationCompress

	case o.Format != nil:
		return OperationFormat
	}

	return OperationUnknown
}
