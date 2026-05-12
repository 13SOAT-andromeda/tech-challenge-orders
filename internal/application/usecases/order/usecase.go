/*package order

import (
	"context"
	"fmt"
	"time"

	orderMaintenanceModel "github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/database/model/order_maintenance"
	orderProductModel "github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/database/model/order_product"
	appmetrics "github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/metrics"
	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/application/ports"
	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/domain"
)

type UseCase struct {
	orderService               ports.OrderService
	productService             ports.ProductService
	maintenanceService         ports.MaintenanceService
	customerService            ports.CustomerService
	userService                ports.UserService
	employeeService            ports.EmployeeService
	emailService               ports.Email
	orderRepository            ports.OrderRepository
	orderProductRepository     ports.OrderProductRepository
	orderMaintenanceRepository ports.OrderMaintenanceRepository
	apiUrl                     string
	metrics                    ports.OrderMetrics
}

func NewOrderUseCase(
	orderService ports.OrderService,
	productsService ports.ProductService,
	maintenanceService ports.MaintenanceService,
	customerService ports.CustomerService,
	userService ports.UserService,
	employeeService ports.EmployeeService,
	emailService ports.Email,
	orderRepository ports.OrderRepository,
	orderProductRepository ports.OrderProductRepository,
	orderMaintenanceRepository ports.OrderMaintenanceRepository,
	apiUrl string,
	metrics ports.OrderMetrics,
) *UseCase {
	if metrics == nil {
		metrics = appmetrics.NoopOrderMetrics{}
	}
	return &UseCase{
		orderService:               orderService,
		productService:             productsService,
		maintenanceService:         maintenanceService,
		customerService:            customerService,
		userService:                userService,
		employeeService:            employeeService,
		emailService:               emailService,
		orderRepository:            orderRepository,
		orderProductRepository:     orderProductRepository,
		orderMaintenanceRepository: orderMaintenanceRepository,
		apiUrl:                     apiUrl,
		metrics:                    metrics,
	}
}

func (uc *UseCase) resolveEmployeeID(ctx context.Context, userID uint) (uint, error) {
	user, err := uc.userService.GetByID(ctx, userID)
	if err != nil || user == nil {
		return 0, fmt.Errorf("employee not found for user %d", userID)
	}

	employee, err := uc.employeeService.GetByPersonID(ctx, user.PersonID)
	if err != nil || employee == nil {
		return 0, fmt.Errorf("employee not found for user %d", userID)
	}

	return employee.ID, nil
}

func referenceTimeForTransition(last *time.Time, dateIn time.Time) time.Time {
	if last != nil && !last.IsZero() {
		return *last
	}
	return dateIn
}

func (uc *UseCase) CreateOrder(ctx context.Context, userID uint, input ports.CreateOrderInput) (*domain.Order, error) {
	employeeID, err := uc.resolveEmployeeID(ctx, userID)
	if err != nil {
		return nil, err
	}

	order := domain.Order{
		DateIn:            time.Now(),
		DateOut:           nil,
		Status:            domain.OrderStatuses.RECEIVED,
		VehicleKilometers: input.VehicleKilometers,
		Note:              input.Note,
		Price:             nil,
		CustomerVehicleID: input.CustomerVehicleID,
		EmployeeID:        employeeID,
		CompanyID:         input.CompanyID,
	}

	created, err := uc.orderService.Create(ctx, order)
	if err != nil {
		return nil, err
	}

	uc.metrics.OrderCreated(ctx)

	return created, nil
}

func (uc *UseCase) AssignOrder(ctx context.Context, orderID uint, userID uint) error {
	order, err := uc.orderService.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	if order == nil {
		return domain.ErrOrderNotFound
	}

	employeeID, err := uc.resolveEmployeeID(ctx, userID)
	if err != nil {
		return err
	}

	from := order.Status
	ref := referenceTimeForTransition(order.LastStatusAt, order.DateIn)
	now := time.Now()
	durationMin := now.Sub(ref).Minutes()

	order.EmployeeID = employeeID
	order.Status = domain.OrderStatuses.IN_ANALYSIS
	order.LastStatusAt = &now

	err = uc.orderService.Update(ctx, *order)

	if err != nil {
		return err
	}

	uc.metrics.OrderStatusTransition(ctx, from, domain.OrderStatuses.IN_ANALYSIS, durationMin)
	return nil
}

func (uc *UseCase) CompleteOrderAnalysis(ctx context.Context, id uint, userID uint, input ports.CreateCompleteOrderAnalysisInput) error {
	order, err := uc.orderService.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("order with Id %d not found", id)
	}

	if order.Status != domain.OrderStatuses.IN_ANALYSIS {
		return fmt.Errorf("order cannot complete analysis. Current status: %s", order.Status)
	}

	productIds := make([]uint, 0, len(input.Products))
	productQuantities := make(map[uint]int, len(input.Products))

	for _, item := range input.Products {
		productIds = append(productIds, item.ID)
		productQuantities[item.ID] = int(item.Quantity)
	}

	products, err := uc.productService.GetByIds(ctx, productIds)
	if err != nil {
		return err
	}

	maintenances, err := uc.maintenanceService.GetByIDs(ctx, input.Maintenances)
	if err != nil {
		return err
	}

	totalPrice := 0.0

	orderProducts := make([]domain.OrderProduct, 0, len(products))
	for _, product := range products {
		quantity := productQuantities[product.ID]
		totalPrice += float64(product.Price) * float64(quantity)

		orderProducts = append(orderProducts, domain.OrderProduct{
			Quantity:  uint(quantity),
			OrderId:   id,
			ProductId: product.ID,
		})
	}

	orderMaintenances := make([]domain.OrderMaintenance, 0, len(maintenances))
	for _, maintenance := range maintenances {
		totalPrice += float64(maintenance.Price)

		orderMaintenances = append(orderMaintenances, domain.OrderMaintenance{
			OrderId:       id,
			MaintenanceId: maintenance.ID,
		})
	}

	for _, op := range orderProducts {
		model := orderProductModel.Model{}
		model.FromDomain(&op)

		_, err := uc.orderProductRepository.Create(ctx, &model)
		if err != nil {
			return fmt.Errorf("failed to associate product %d with order %d: %w", op.ProductId, id, err)
		}
	}

	for _, om := range orderMaintenances {
		model := orderMaintenanceModel.Model{}
		model.FromDomain(&om)
		_, err := uc.orderMaintenanceRepository.Create(ctx, &model)
		if err != nil {
			return fmt.Errorf("failed to associate maintenance %d with order %d: %w", om.MaintenanceId, id, err)
		}
	}

	employeeID, err := uc.resolveEmployeeID(ctx, userID)
	if err != nil {
		return err
	}

	from := order.Status
	ref := referenceTimeForTransition(order.LastStatusAt, order.DateIn)
	now := time.Now()
	durationMin := now.Sub(ref).Minutes()

	order.DiagnosticNote = input.DiagnosticNote
	order.Status = domain.OrderStatuses.ANALYSIS_FINISHED
	order.Price = &totalPrice
	order.EmployeeID = employeeID
	order.LastStatusAt = &now

	if err := uc.orderService.Update(ctx, *order); err != nil {
		return fmt.Errorf("failed to complete order analysis: %w", err)
	}

	uc.metrics.OrderStatusTransition(ctx, from, domain.OrderStatuses.ANALYSIS_FINISHED, durationMin)
	return nil
}

func (uc *UseCase) ApproveOrder(ctx context.Context, id uint) error {
	existentOrder, err := uc.orderRepository.FindByID(ctx, id)

	if err != nil {
		return fmt.Errorf("order with Id %d not found", id)
	}

	if domain.OrderStatus(existentOrder.Status) != domain.OrderStatuses.AWAITING_APPROVAL {
		return fmt.Errorf("order cannot be approved. Current status: %s", existentOrder.Status)
	}

	from := domain.OrderStatus(existentOrder.Status)
	ref := referenceTimeForTransition(existentOrder.LastStatusAt, existentOrder.DateIn)
	now := time.Now()
	durationMin := now.Sub(ref).Minutes()

	existentOrder.Status = string(domain.OrderStatuses.APPROVED)
	existentOrder.DateApproved = &now
	existentOrder.LastStatusAt = &now

	if err := uc.orderRepository.Update(ctx, existentOrder); err != nil {
		return fmt.Errorf("failed to approve order: %w", err)
	}

	uc.metrics.OrderStatusTransition(ctx, from, domain.OrderStatuses.APPROVED, durationMin)
	uc.metrics.OrderApproved(ctx)
	return nil
}

func (uc *UseCase) RejectOrder(ctx context.Context, id uint) error {
	existentOrder, err := uc.orderRepository.FindByID(ctx, id)

	if err != nil {
		return fmt.Errorf("order with Id %d not found", id)
	}

	if domain.OrderStatus(existentOrder.Status) != domain.OrderStatuses.AWAITING_APPROVAL {
		return fmt.Errorf("order cannot be reject. Current status: %s", existentOrder.Status)
	}

	from := domain.OrderStatus(existentOrder.Status)
	ref := referenceTimeForTransition(existentOrder.LastStatusAt, existentOrder.DateIn)
	now := time.Now()
	durationMin := now.Sub(ref).Minutes()

	existentOrder.Status = string(domain.OrderStatuses.FINISHED)
	existentOrder.DateRejected = &now
	existentOrder.LastStatusAt = &now

	if err := uc.orderRepository.Update(ctx, existentOrder); err != nil {
		return fmt.Errorf("failed to reject order: %w", err)
	}

	uc.metrics.OrderStatusTransition(ctx, from, domain.OrderStatuses.FINISHED, durationMin)
	uc.metrics.OrderRejected(ctx)
	return nil
}

func (uc *UseCase) ArchiveOrder(ctx context.Context, id uint) error {
	existentOrder, err := uc.orderRepository.FindByID(ctx, id)

	if err != nil {
		return fmt.Errorf("order with Id %d not found", id)
	}

	if domain.OrderStatus(existentOrder.Status) != domain.OrderStatuses.FINISHED {
		return fmt.Errorf("order cannot be archived. Current status: %s", existentOrder.Status)
	}

	from := domain.OrderStatus(existentOrder.Status)
	ref := referenceTimeForTransition(existentOrder.LastStatusAt, existentOrder.DateIn)
	now := time.Now()
	durationMin := now.Sub(ref).Minutes()

	existentOrder.Status = string(domain.OrderStatuses.DELIVERED)
	existentOrder.DateOut = &now
	existentOrder.LastStatusAt = &now

	if err := uc.orderRepository.Update(ctx, existentOrder); err != nil {
		return fmt.Errorf("failed to archive order: %w", err)
	}

	uc.metrics.OrderStatusTransition(ctx, from, domain.OrderStatuses.DELIVERED, durationMin)
	return nil
}

func (uc *UseCase) RequestApproval(ctx context.Context, id uint) error {
	existentOrder, err := uc.orderRepository.FindOrderByID(ctx, id)

	if err != nil {
		return fmt.Errorf("order with Id %d not found", id)
	}

	if domain.OrderStatus(existentOrder.Status) != domain.OrderStatuses.ANALYSIS_FINISHED {
		return fmt.Errorf("notification cannot be sent. Order should be in %s status. Current status: %s", domain.OrderStatuses.ANALYSIS_FINISHED, existentOrder.Status)
	}

	c, err := uc.customerService.GetByID(ctx, existentOrder.CustomerVehicle.CustomerID)

	if err != nil {
		return fmt.Errorf("error on find order's customer: %w", err)
	}

	html, err := uc.orderService.GetApprovalTemplate(*existentOrder.ToDomain(), *c, uc.apiUrl)

	if err != nil {
		return fmt.Errorf("failed to parse mail template: %w", err)
	}

	customerName := ""
	customerEmail := ""
	if c.Person != nil {
		customerName = c.Person.Name
		customerEmail = c.Person.Email
	}

	err = uc.emailService.Send(customerName, customerEmail, "Aprovação de Ordem de Serviço", html)

	if err != nil {
		return fmt.Errorf("failed to send approval notification: %w", err)
	}

	from := domain.OrderStatus(existentOrder.Status)
	ref := referenceTimeForTransition(existentOrder.LastStatusAt, existentOrder.DateIn)
	now := time.Now()
	durationMin := now.Sub(ref).Minutes()

	existentOrder.Status = string(domain.OrderStatuses.AWAITING_APPROVAL)
	existentOrder.LastStatusAt = &now

	if err := uc.orderRepository.Update(ctx, existentOrder); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	uc.metrics.OrderStatusTransition(ctx, from, domain.OrderStatuses.AWAITING_APPROVAL, durationMin)
	return nil
}

func (uc *UseCase) StartWorkOrder(ctx context.Context, id uint) error {
	order, err := uc.orderService.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get order with id %d: %w", id, err)
	}

	if order.Status != domain.OrderStatuses.APPROVED {
		return fmt.Errorf("order cannot start work. current status: %s", order.Status)
	}

	productItems := make([]domain.StockItem, 0, len(*order.Products))
	operation := domain.StockOperationDecrease

	for _, item := range *order.Products {
		orderedQuantity := *item.Quantity
		available, err := uc.productService.CheckAvailability(ctx, item.ID, orderedQuantity)
		if err != nil {
			return fmt.Errorf("failed to check availability for product %d: %w", item.ID, err)
		}

		if !available {
			return fmt.Errorf("cannot start work: product ID %d is not available in quantity %d", item.ID, orderedQuantity)
		}

		productItems = append(productItems, domain.StockItem{
			ID:        item.ID,
			Quantity:  orderedQuantity,
			Operation: &operation,
		})
	}

	err = uc.productService.UpdateStock(ctx, productItems)
	if err != nil {
		return fmt.Errorf("failed to decrement stock: %w", err)
	}

	from := order.Status
	ref := referenceTimeForTransition(order.LastStatusAt, order.DateIn)
	now := time.Now()
	durationMin := now.Sub(ref).Minutes()

	order.Status = domain.OrderStatuses.IN_PROGRESS
	order.LastStatusAt = &now

	if err = uc.orderService.Update(ctx, *order); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	uc.metrics.OrderStatusTransition(ctx, from, domain.OrderStatuses.IN_PROGRESS, durationMin)
	return nil
}

func (uc *UseCase) CompleteWorkOrder(ctx context.Context, id uint) error {
	order, err := uc.orderService.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get order with id %d: %w", id, err)
	}

	if order.Status != domain.OrderStatuses.IN_PROGRESS {
		return fmt.Errorf("order cannot complete work. current status: %s", order.Status)
	}

	from := order.Status
	ref := referenceTimeForTransition(order.LastStatusAt, order.DateIn)
	now := time.Now()
	durationMin := now.Sub(ref).Minutes()

	order.Status = domain.OrderStatuses.FINISHED
	order.LastStatusAt = &now

	if err = uc.orderService.Update(ctx, *order); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	uc.metrics.OrderStatusTransition(ctx, from, domain.OrderStatuses.FINISHED, durationMin)
	return nil
}
*/