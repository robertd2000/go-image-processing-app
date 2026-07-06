package metricsmem

import "sync"

type FakeMetrics struct {
	mu sync.Mutex

	ImageSaveSuccess int
	ImageSaveError   int
}

func NewFakeMetrics() *FakeMetrics {
	return &FakeMetrics{}
}

func (m *FakeMetrics) IncImageSaveSuccess() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ImageSaveSuccess++
}

func (m *FakeMetrics) IncImageSaveError() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ImageSaveError++
}
