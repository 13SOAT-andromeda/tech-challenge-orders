package domain

type Item struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Price int64  `json:"price"`
	Type  string `json:"type"`
}

// StockItem represents a product quantity used in order analysis and stock operations.
type StockItem struct {
	ID       string `json:"id"`
	Quantity uint   `json:"quantity"`
}
