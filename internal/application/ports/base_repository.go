package ports

import "context"

type Repository[T any] interface {
	FindByID(ctx context.Context, id uint) (*T, error)
	FindAll(ctx context.Context, includeDeleted bool) ([]T, error)
	Create(ctx context.Context, entity *T) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uint) error
}
