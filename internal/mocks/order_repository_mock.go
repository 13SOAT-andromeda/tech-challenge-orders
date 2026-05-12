package mocks

import (
	"context"

	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/database/model/order"
	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/application/ports"
	"github.com/stretchr/testify/mock"
)

type MockOrderRepository struct {
	mock.Mock
}

var _ ports.OrderRepository = (*MockOrderRepository)(nil)

func (m *MockOrderRepository) FindAll(ctx context.Context, includeDeleted bool) ([]order.Model, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]order.Model), args.Error(1)
}

func (m *MockOrderRepository) FindByID(ctx context.Context, id uint) (*order.Model, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.Model), args.Error(1)
}

func (m *MockOrderRepository) FindOrderByID(ctx context.Context, id uint) (*order.Model, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.Model), args.Error(1)
}

func (m *MockOrderRepository) FindByIDs(ctx context.Context, ids []uint) ([]order.Model, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]order.Model), args.Error(1)
}

func (m *MockOrderRepository) Create(ctx context.Context, entity *order.Model) (*order.Model, error) {
	args := m.Called(ctx, entity)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.Model), args.Error(1)
}

func (m *MockOrderRepository) Update(ctx context.Context, entity *order.Model) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockOrderRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrderRepository) Search(ctx context.Context, params ports.OrderSearch) ([]order.Model, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]order.Model), args.Error(1)
}
