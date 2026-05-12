package services

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/database/model/order"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/adapter/email"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
	"github.com/13SOAT-andromeda/tech-challenge-orders/pkg/monetary"
)

type OrderService struct {
	repo ports.OrderRepository
}

func NewOrderService(repo ports.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) Create(ctx context.Context, o domain.Order) (*domain.Order, error) {
	model := order.Model{}
	model.FromDomain(&o)

	_, err := s.repo.Create(ctx, &model)

	if err != nil {
		return nil, err
	}

	created := model.ToDomain()

	return created, nil
}

func (s *OrderService) GetByID(ctx context.Context, id uint) (*domain.Order, error) {
	result, err := s.repo.FindOrderByID(ctx, id)

	if err != nil {
		return nil, err
	}
	o := result.ToDomain()

	return o, nil
}

func getPriorityStatus(status domain.Status) int {
	switch status {
	case domain.IN_PROGRESS:
		return 0
	case domain.AWAITING_APPROVAL:
		return 1
	case domain.IN_ANALYSIS:
		return 2
	case domain.RECEIVED:
		return 3
	default:
		return 999
	}
}

func (s *OrderService) GetAll(ctx context.Context, params map[string]interface{}) (*[]domain.Order, error) {
	oSearch := ports.OrderSearch{Status: "", Enabled: true, OrderBy: "date_in", SortDesc: false}

	if params["status"] != nil {
		oSearch.Status = params["status"].(string)
	}

	if params["enabled"] != nil {
		oSearch.Enabled = params["enabled"].(bool)
	}

	if params["orderby"] != nil {
		oSearch.OrderBy = params["orderby"].(string)
	}

	if params["sortdesc"] != nil {
		if sortDescStr, ok := params["sortdesc"].(string); ok {
			oSearch.SortDesc = sortDescStr == "true" || sortDescStr == "1"
		}
	}

	orders, err := s.repo.Search(ctx, oSearch)

	if err != nil {
		return nil, err
	}

	ordersD := make([]domain.Order, 0, len(orders))

	for _, item := range orders {
		if item.Status != string(domain.FINISHED) && item.Status != string(domain.DELIVERED) {
			ordersD = append(ordersD, *item.ToDomain())
		}
	}

	sort.Slice(ordersD, func(i, j int) bool {
		pI := getPriorityStatus(ordersD[i].Status)
		pJ := getPriorityStatus(ordersD[j].Status)
		return pI < pJ
	})

	return &ordersD, nil
}

func (s *OrderService) Delete(ctx context.Context, id uint) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return errors.New("An error occurred while trying to delete the order")
	}

	return nil
}

func (s *OrderService) Update(ctx context.Context, o domain.Order) error {
	model := order.Model{}
	model.FromDomain(&o)

	err := s.repo.Update(ctx, &model)
	if err != nil {
		return err
	}

	return nil
}

func (s *OrderService) GetApprovalTemplate(order domain.Order, customer domain.Customer, apiUrl string) (string, error) {
	html, err := email.LoadTemplate(email.ORDER_APPROVAL)

	if err != nil {
		return "", err
	}

	dateIn := order.DateIn.Format("02/01/2006")
	price := ""
	diagnosticNote := ""
	note := ""
	id := strconv.FormatUint(uint64(order.ID), 10)

	if order.DiagnosticNote != nil {
		diagnosticNote = *order.DiagnosticNote
	}

	if order.Price != nil {
		price = monetary.FormatPtBrCurrency(*order.Price)
	}

	if order.Note != nil {
		note = *order.Note
	}

	customerName := ""
	if customer.Person != nil {
		customerName = customer.Person.Name
	}
	html = strings.ReplaceAll(html, "$Name", customerName)
	html = strings.ReplaceAll(html, "$ID", id)
	html = strings.ReplaceAll(html, "$DateIn", dateIn)
	html = strings.ReplaceAll(html, "$DiagnosticNote", diagnosticNote)
	html = strings.ReplaceAll(html, "$Value", price)
	html = strings.ReplaceAll(html, "$Note", note)
	html = strings.ReplaceAll(html, "$Approval_url", apiUrl+"/orders/"+id+"/approve")
	html = strings.ReplaceAll(html, "$Repprove_url", apiUrl+"/orders/"+id+"/reject")

	return html, nil
}
