package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
)

type UsersHTTPClient struct {
	baseClient
	baseURL string
}

func NewUsersHTTPClient(baseURL string, timeoutMs int) *UsersHTTPClient {
	return &UsersHTTPClient{
		baseClient: newBaseClient("users-client", time.Duration(timeoutMs)*time.Millisecond),
		baseURL:    strings.TrimRight(baseURL, "/"),
	}
}

// API response types (unexported)

type customerVehicleResp struct {
	ID         int64        `json:"id"`
	CustomerID int64        `json:"customer_id"`
	VehicleID  int64        `json:"vehicle_id"`
	Vehicle    vehicleResp  `json:"vehicle,omitempty"`
	Customer   customerResp `json:"customer,omitempty"`
}

type vehicleResp struct {
	ID    int64  `json:"id"`
	Plate string `json:"plate"`
	Name  string `json:"name"`
	Year  int    `json:"year"`
	Brand string `json:"brand"`
	Color string `json:"color"`
}

type customerResp struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type userResp struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type employeeResp struct {
	ID       int64  `json:"id"`
	Position string `json:"position"`
}

type companyResp struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type maintenanceResp struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	PriceCents int64  `json:"price_cents"`
}

func (c *UsersHTTPClient) GetCustomerVehicle(ctx context.Context, id int64) (*ports.CustomerVehicle, error) {
	url := c.baseURL + "/v1/users/customers/" + strconv.FormatInt(id, 10) + "/vehicles"
	resp, err := c.execute(ctx, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("get customer vehicle %d: %w", id, err)
	}
	defer resp.Body.Close()

	var body []customerVehicleResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode customer vehicle: %w", err)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("customer vehicle %d not found", id)
	}

	cv := body[0]
	customer := &ports.Customer{ID: cv.Customer.ID, Name: cv.Customer.Name, Email: cv.Customer.Email}
	return &ports.CustomerVehicle{
		ID:         cv.ID,
		CustomerID: cv.CustomerID,
		VehicleID:  cv.VehicleID,
		Customer:   customer,
		Vehicle: domain.VehicleSnapshot{
			ID:    cv.Vehicle.ID,
			Plate: cv.Vehicle.Plate,
			Name:  cv.Vehicle.Name,
			Year:  cv.Vehicle.Year,
			Brand: cv.Vehicle.Brand,
			Color: cv.Vehicle.Color,
		},
	}, nil
}

func (c *UsersHTTPClient) GetCustomer(ctx context.Context, id int64) (*ports.Customer, error) {
	url := c.baseURL + "/v1/users/customers/" + strconv.FormatInt(id, 10)
	resp, err := c.execute(ctx, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("get customer %d: %w", id, err)
	}
	defer resp.Body.Close()

	var body customerResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode customer: %w", err)
	}

	return &ports.Customer{ID: body.ID, Name: body.Name, Email: body.Email}, nil
}

func (c *UsersHTTPClient) GetUser(ctx context.Context, id int64) (*ports.User, error) {
	url := c.baseURL + "/v1/users/" + strconv.FormatInt(id, 10)
	resp, err := c.execute(ctx, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("get user %d: %w", id, err)
	}
	defer resp.Body.Close()

	var body userResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode user: %w", err)
	}

	return &ports.User{ID: body.ID, Email: body.Email, Role: body.Role}, nil
}

func (c *UsersHTTPClient) GetEmployee(ctx context.Context, id int64) (*ports.Employee, error) {
	url := c.baseURL + "/v1/users/employees/" + strconv.FormatInt(id, 10)
	resp, err := c.execute(ctx, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("get employee %d: %w", id, err)
	}
	defer resp.Body.Close()

	var body employeeResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode employee: %w", err)
	}

	return &ports.Employee{ID: body.ID, Position: body.Position}, nil
}

func (c *UsersHTTPClient) GetEmployeeByUserID(ctx context.Context, userID int64) (*ports.Employee, error) {
	url := c.baseURL + "/v1/users/users/" + strconv.FormatInt(userID, 10) + "/employee"
	resp, err := c.execute(ctx, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("get employee by user %d: %w", userID, err)
	}
	defer resp.Body.Close()

	var body employeeResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode employee: %w", err)
	}

	return &ports.Employee{ID: body.ID, Position: body.Position}, nil
}

func (c *UsersHTTPClient) GetCompany(ctx context.Context, id int64) (*ports.Company, error) {
	url := c.baseURL + "/v1/users/companies/" + strconv.FormatInt(id, 10)
	resp, err := c.execute(ctx, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("get company %d: %w", id, err)
	}
	defer resp.Body.Close()

	var body companyResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode company: %w", err)
	}

	return &ports.Company{ID: body.ID, Name: body.Name}, nil
}

func (c *UsersHTTPClient) GetMaintenancesBatch(ctx context.Context, ids []int64) ([]domain.ItemSnapshot, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.FormatInt(id, 10))
	}
	url := c.baseURL + "/v1/users/maintenances/check-batch?ids=" + strings.Join(parts, ",")
	resp, err := c.execute(ctx, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("get maintenances batch: %w", err)
	}
	defer resp.Body.Close()

	var body []maintenanceResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode maintenances: %w", err)
	}

	items := make([]domain.ItemSnapshot, 0, len(body))
	for _, m := range body {
		items = append(items, domain.ItemSnapshot{
			ID:             m.ID,
			Kind:           domain.ItemKindMaintenance,
			Name:           m.Name,
			Quantity:       1,
			UnitPriceCents: m.PriceCents,
		})
	}
	return items, nil
}

var _ ports.UsersClient = (*UsersHTTPClient)(nil)
