package role

import "context"

type Repository interface {
	ByID(ctx context.Context, id int64) (*Role, error)
	ByName(ctx context.Context, name Name) (*Role, error)
	Save(ctx context.Context, role *Role) error
}
