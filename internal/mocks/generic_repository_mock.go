package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockGenericRepository[T any] struct {
	mock.Mock
}

func (m *MockGenericRepository[T]) FindByID(ctx context.Context, id uint) (*T, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*T), args.Error(1)
}

func (m *MockGenericRepository[T]) FindAll(ctx context.Context, includeDeleted bool) ([]T, error) {
	args := m.Called(ctx, false)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]T), args.Error(1)
}

func (m *MockGenericRepository[T]) Create(ctx context.Context, entity *T) (*T, error) {
	args := m.Called(ctx, entity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*T), args.Error(1)
}

func (m *MockGenericRepository[T]) Update(ctx context.Context, entity *T) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockGenericRepository[T]) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
