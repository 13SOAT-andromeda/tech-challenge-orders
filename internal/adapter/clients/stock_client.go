package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/application/ports"
	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
)

type StockHTTPClient struct {
	baseClient
	baseURL string
}

func NewStockHTTPClient(baseURL string, timeoutMs int) *StockHTTPClient {
	return &StockHTTPClient{
		baseClient: newBaseClient("stock-client", time.Duration(timeoutMs)*time.Millisecond),
		baseURL:    strings.TrimRight(baseURL, "/"),
	}
}

// API request/response types (unexported)

type checkBatchRequest struct {
	Items []stockItemReq `json:"items"`
}

type stockItemReq struct {
	ID  int64 `json:"id"`
	Qty uint  `json:"qty"`
}

type checkBatchResponse struct {
	Available bool          `json:"available"`
	Items     []productResp `json:"items"`
}

type productResp struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Quantity       uint   `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
}

func (c *StockHTTPClient) CheckProductsBatch(ctx context.Context, items []domain.StockItem) ([]domain.ItemSnapshot, error) {
	if len(items) == 0 {
		return nil, nil
	}

	reqItems := make([]stockItemReq, 0, len(items))
	for _, it := range items {
		reqItems = append(reqItems, stockItemReq{ID: it.ID, Qty: it.Quantity})
	}

	payload, err := json.Marshal(checkBatchRequest{Items: reqItems})
	if err != nil {
		return nil, fmt.Errorf("marshal stock request: %w", err)
	}

	url := c.baseURL + "/v1/catalog/products/check-batch"
	resp, err := c.execute(ctx, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	})
	if err != nil {
		return nil, fmt.Errorf("check products batch: %w", err)
	}
	defer resp.Body.Close()

	var body checkBatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode stock response: %w", err)
	}

	snapshots := make([]domain.ItemSnapshot, 0, len(body.Items))
	for _, p := range body.Items {
		snapshots = append(snapshots, domain.ItemSnapshot{
			ID:             p.ID,
			Kind:           domain.ItemKindProduct,
			Name:           p.Name,
			Quantity:       p.Quantity,
			UnitPriceCents: p.UnitPriceCents,
		})
	}
	return snapshots, nil
}

var _ ports.StockClient = (*StockHTTPClient)(nil)
