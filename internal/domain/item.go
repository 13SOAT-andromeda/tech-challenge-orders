package domain

type Item struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Price int64  `json:"price"`
	Type  string `json:"type"`
}
