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
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[image.ID()]; ok {
		return imageDomain.ErrAlreadyExists
	}

	r.data[image.ID()] = cloneImage(image)

	return nil
}

func (r *imageRepo) GetByID(ctx context.Context, id uuid.UUID) (*imageDomain.Image, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, ok := r.data[id]
	if !ok {
		return nil, imageDomain.ErrNotFound
	}

	return cloneImage(data), nil
}

func (r *imageRepo) GetByUser(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]*imageDomain.Image, error) {

	if offset < 0 {
		offset = 0
	}

	if limit == 0 {
		return []*imageDomain.Image{}, nil
	}

	r.mu.RLock()
	var filtered []*imageDomain.Image
	for _, img := range r.data {
		if img.UserID() == userID {
			filtered = append(filtered, cloneImage(img))
		}
	}
	r.mu.RUnlock()

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt().Before(filtered[j].CreatedAt())
	})

	if offset >= len(filtered) {
		return []*imageDomain.Image{}, nil
	}

	filtered = filtered[offset:]

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

func cloneImage(src *imageDomain.Image) *imageDomain.Image {
	if src == nil {
		return nil
	}

	copy := *src
	return &copy
}
