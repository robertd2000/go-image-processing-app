package transformation

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusDone       Status = "done"
	StatusFailed     Status = "failed"
)

func (s Status) String() string {
	return string(s)
}

func (s Status) IsValid() bool {
	switch s {
	case StatusPending,
		StatusProcessing,
		StatusDone,
		StatusFailed:
		return true
	default:
		return false
	}
}

func (s Status) IsFinished() bool {
	return s == StatusDone || s == StatusFailed
}
