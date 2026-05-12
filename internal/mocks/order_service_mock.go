package mocks

import (
	"context"

	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/application/ports"
	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockOrderService struct {
	mock.Mock
}

var _ ports.OrderService = (*MockOrderService)(nil)

func (m *MockOrderService) Create(ctx context.Context, input domain.Order) (*domain.Order, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderService) GetByID(ctx context.Context, id uint) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderService) GetAll(ctx context.Context, params map[string]interface{}) (*[]domain.Order, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]domain.Order), args.Error(1)
}

func (m *MockOrderService) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrderService) GetApprovalTemplate(order domain.Order, customer domain.Customer, apiUrl string) (string, error) {
	args := m.Called(order, customer, apiUrl)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

func (m *MockOrderService) Update(ctx context.Context, input domain.Order) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}
