package transformation

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

func (s Status) String() string {
	return string(s)
}

func (s Status) IsValid() bool {
	switch s {
	case StatusPending,
		StatusProcessing,
		StatusCompleted,
		StatusFailed:
		return true
	default:
		return false
	}
}

func (s Status) IsFinished() bool {
	return s == StatusCompleted || s == StatusFailed
}
