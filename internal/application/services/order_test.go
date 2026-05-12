package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/database/model/order"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createTestOrderModel(id uint) *order.Model {
	d := time.Now()
	om := &order.Model{
		DateIn:            d,
		Status:            string(domain.Statuses.RECEIVED),
		VehicleKilometers: 100,
		EmployeeID:        1,
		CustomerVehicleID: 1,
		CompanyID:         1,
	}
	om.ID = id
	return om
}

func TestOrderService_Create_Success(t *testing.T) {
	mockRepo := new(mocks.MockOrderRepository)
	service := NewOrderService(mockRepo)
	ctx := context.Background()

	domainOrder := domain.Order{
		DateIn:            time.Now(),
		Status:            domain.Statuses.RECEIVED,
		VehicleKilometers: 100,
		EmployeeID:        1,
		CustomerVehicleID: 1,
		CompanyID:         1,
	}

	createdModel := createTestOrderModel(1)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*order.Model")).Return(createdModel, nil)

	result, err := service.Create(ctx, domainOrder)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.Statuses.RECEIVED, result.Status)
	mockRepo.AssertExpectations(t)
}

func TestOrderService_Create_RepositoryError(t *testing.T) {
	mockRepo := new(mocks.MockOrderRepository)
	service := NewOrderService(mockRepo)
	ctx := context.Background()

	domainOrder := domain.Order{Status: domain.Statuses.FINISHED}
	expectedErr := errors.New("create failed")
	mockRepo.On("Create", ctx, mock.AnythingOfType("*order.Model")).Return(nil, expectedErr)

	result, err := service.Create(ctx, domainOrder)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestOrderService_GetByID_Success(t *testing.T) {
	mockRepo := new(mocks.MockOrderRepository)
	service := NewOrderService(mockRepo)
	ctx := context.Background()

	om := createTestOrderModel(1)
	mockRepo.On("FindOrderByID", ctx, uint(1)).Return(om, nil)

	result, err := service.GetByID(ctx, 1)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	mockRepo.AssertExpectations(t)
}

func TestOrderService_GetByID_NotFound(t *testing.T) {
	mockRepo := new(mocks.MockOrderRepository)
	service := NewOrderService(mockRepo)
	ctx := context.Background()

	mockRepo.On("FindOrderByID", ctx, uint(999)).Return(nil, errors.New("not found"))

	result, err := service.GetByID(ctx, 999)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestOrderService_GetAll_Success(t *testing.T) {
	mockRepo := new(mocks.MockOrderRepository)
	service := NewOrderService(mockRepo)
	ctx := context.Background()

	models := []order.Model{*createTestOrderModel(1), *createTestOrderModel(2)}
	mockRepo.On("Search", ctx, mock.Anything).Return(models, nil)

	result, err := service.GetAll(ctx, map[string]interface{}{})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, *result, 2)
	mockRepo.AssertExpectations(t)
}

func TestOrderService_GetAll_RepositoryError(t *testing.T) {
	mockRepo := new(mocks.MockOrderRepository)
	service := NewOrderService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Search", ctx, mock.Anything).Return(nil, errors.New("search error"))

	result, err := service.GetAll(ctx, map[string]interface{}{})

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestOrderService_Delete_Success(t *testing.T) {
	mockRepo := new(mocks.MockOrderRepository)
	service := NewOrderService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Delete", ctx, uint(1)).Return(nil)

	err := service.Delete(ctx, 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestOrderService_Delete_DeleteError(t *testing.T) {
	mockRepo := new(mocks.MockOrderRepository)
	service := NewOrderService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Delete", ctx, uint(1)).Return(errors.New("delete error"))

	err := service.Delete(ctx, 1)

	assert.Error(t, err)
	assert.EqualError(t, err, "An error occurred while trying to delete the order")
	mockRepo.AssertExpectations(t)
}

func TestOrderService_Update_Success(t *testing.T) {
	mockRepo := new(mocks.MockOrderRepository)
	service := NewOrderService(mockRepo)
	ctx := context.Background()

	domainOrder := domain.Order{ID: 1, VehicleKilometers: 200}
	mockRepo.On("Update", ctx, mock.AnythingOfType("*order.Model")).Return(nil)

	err := service.Update(ctx, domainOrder)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestOrderService_Update_RepositoryError(t *testing.T) {
	mockRepo := new(mocks.MockOrderRepository)
	service := NewOrderService(mockRepo)
	ctx := context.Background()

	domainOrder := domain.Order{ID: 1}
	mockRepo.On("Update", ctx, mock.AnythingOfType("*order.Model")).Return(errors.New("update failed"))

	err := service.Update(ctx, domainOrder)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestOrderService_GetAll_FilterFinishedAndDelivered(t *testing.T) {
	mockRepo := new(mocks.MockOrderRepository)
	service := NewOrderService(mockRepo)
	ctx := context.Background()

	received := createTestOrderModel(1)
	received.Status = string(domain.Statuses.RECEIVED)

	finished := createTestOrderModel(2)
	finished.Status = string(domain.FINISHED)

	delivered := createTestOrderModel(3)
	delivered.Status = string(domain.DELIVERED)

	models := []order.Model{
		*received,
		*finished,
		*delivered,
	}

	mockRepo.On("Search", ctx, mock.Anything).Return(models, nil)

	result, err := service.GetAll(ctx, map[string]interface{}{})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, *result, 1)
	assert.Equal(t, domain.Statuses.RECEIVED, (*result)[0].Status)

	mockRepo.AssertExpectations(t)
}

func TestOrderService_GetAll_SortByStatusPriority(t *testing.T) {
	mockRepo := new(mocks.MockOrderRepository)
	service := NewOrderService(mockRepo)
	ctx := context.Background()

	o1 := createTestOrderModel(1)
	o1.Status = string(domain.Statuses.RECEIVED)

	o2 := createTestOrderModel(2)
	o2.Status = string(domain.Statuses.IN_PROGRESS)

	o3 := createTestOrderModel(3)
	o3.Status = string(domain.Statuses.AWAITING_APPROVAL)

	models := []order.Model{*o1, *o2, *o3}

	mockRepo.On("Search", ctx, mock.Anything).Return(models, nil)

	result, err := service.GetAll(ctx, map[string]interface{}{})

	assert.NoError(t, err)
	assert.Len(t, *result, 3)

	// valida que está ordenado por prioridade
	for i := 0; i < len(*result)-1; i++ {
		p1 := getPriorityStatus((*result)[i].Status)
		p2 := getPriorityStatus((*result)[i+1].Status)
		assert.LessOrEqual(t, p1, p2)
	}

	mockRepo.AssertExpectations(t)
}

func TestOrderService_GetAll_WithStatusParam(t *testing.T) {
	mockRepo := new(mocks.MockOrderRepository)
	service := NewOrderService(mockRepo)
	ctx := context.Background()

	model := createTestOrderModel(1)
	model.Status = string(domain.Statuses.RECEIVED)

	mockRepo.
		On("Search", ctx, mock.MatchedBy(func(search any) bool {
			s := search.(ports.OrderSearch)
			return s.Status == string(domain.Statuses.RECEIVED)
		})).
		Return([]order.Model{*model}, nil)

	params := map[string]interface{}{
		"status": string(domain.Statuses.RECEIVED),
	}

	result, err := service.GetAll(ctx, params)

	assert.NoError(t, err)
	assert.Len(t, *result, 1)

	mockRepo.AssertExpectations(t)
}
