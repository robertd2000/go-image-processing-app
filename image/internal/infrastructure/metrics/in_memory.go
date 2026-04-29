package metrics

import "sync/atomic"

type InMemoryMetrics struct {
	saveSuccess uint64
	saveError   uint64
}

func (m *InMemoryMetrics) IncImageSaveSuccess() {
	atomic.AddUint64(&m.saveSuccess, 1)
}

func (m *InMemoryMetrics) IncImageSaveError() {
	atomic.AddUint64(&m.saveError, 1)
}
