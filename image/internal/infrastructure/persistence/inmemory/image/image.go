package imagemem

import (
	"context"
	"sort"
	"sync"

	imageDomain "github.com/robertd2000/go-image-processing-app/image/internal/domain/image"

	"github.com/google/uuid"
)

type imageRepo struct {
	mu   sync.RWMutex
	data map[uuid.UUID]*imageDomain.Image
}

func NewInMemoryImageRepo() *imageRepo {
	return &imageRepo{
		data: make(map[uuid.UUID]*imageDomain.Image),
	}
}

func (r *imageRepo) Save(ctx context.Context, image *imageDomain.Image) error {
	if _, ok := r.data[image.ID()]; ok {
		return imageDomain.ErrAlreadyExists
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[image.ID()] = image

	return nil
}

func (r *imageRepo) GetByID(ctx context.Context, id uuid.UUID) (*imageDomain.Image, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, ok := r.data[id]
	if !ok {
		return nil, imageDomain.ErrNotFound
	}

	return data, nil
}

func (r *imageRepo) GetByUser(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]*imageDomain.Image, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*imageDomain.Image
	for _, img := range r.data {
		if img.UserID() == userID {
			filtered = append(filtered, img)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt().Before(filtered[j].CreatedAt())
	})

	if offset >= len(filtered) {
		return []*imageDomain.Image{}, nil
	}

	filtered = filtered[offset:]

	if limit >= 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	result := make([]*imageDomain.Image, len(filtered))
	copy(result, filtered)

	return result, nil
}
