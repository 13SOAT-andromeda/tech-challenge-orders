package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	ID         string      `json:"id"`
	CustomerID string      `json:"customer_id"`
	Vehicle    vehicleResp `json:"vehicle"`
}

type vehicleResp struct {
	ID    string `json:"id"`
	Plate string `json:"plate"`
	Name  string `json:"name"`
	Year  int    `json:"year"`
	Brand string `json:"brand"`
	Color string `json:"color"`
}

type customerResp struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type userResp struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type maintenanceResp struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PriceCents int64  `json:"price_cents"`
}

func (c *UsersHTTPClient) GetCustomerVehicle(ctx context.Context, id string) (*ports.CustomerVehicle, error) {
	url := c.baseURL + "/customer-vehicles/" + id
	resp, err := c.execute(ctx, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("get customer vehicle %s: %w", id, err)
	}
	defer resp.Body.Close()

	var body customerVehicleResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode customer vehicle: %w", err)
	}

	return &ports.CustomerVehicle{
		ID:         body.ID,
		CustomerID: body.CustomerID,
		Vehicle: domain.VehicleSnapshot{
			ID:    body.Vehicle.ID,
			Plate: body.Vehicle.Plate,
			Name:  body.Vehicle.Name,
			Year:  body.Vehicle.Year,
			Brand: body.Vehicle.Brand,
			Color: body.Vehicle.Color,
		},
	}, nil
}

func (c *UsersHTTPClient) GetCustomer(ctx context.Context, id string) (*ports.Customer, error) {
	url := c.baseURL + "/customers/" + id
	resp, err := c.execute(ctx, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("get customer %s: %w", id, err)
	}
	defer resp.Body.Close()

	var body customerResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode customer: %w", err)
	}

	return &ports.Customer{ID: body.ID, Name: body.Name, Email: body.Email}, nil
}

func (c *UsersHTTPClient) GetUser(ctx context.Context, id string) (*ports.User, error) {
	url := c.baseURL + "/users/" + id
	resp, err := c.execute(ctx, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("get user %s: %w", id, err)
	}
	defer resp.Body.Close()

	var body userResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode user: %w", err)
	}

	return &ports.User{ID: body.ID, Email: body.Email, Role: body.Role}, nil
}

func (c *UsersHTTPClient) GetMaintenancesBatch(ctx context.Context, ids []string) ([]domain.ItemSnapshot, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	url := c.baseURL + "/maintenances/check-batch?ids=" + strings.Join(ids, ",")
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
