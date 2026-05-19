package domain

type ItemKind string

const (
	ItemKindProduct     ItemKind = "product"
	ItemKindMaintenance ItemKind = "maintenance"
)

type VehicleSnapshot struct {
	ID    int64  `json:"id"`
	Plate string `json:"plate"`
	Name  string `json:"name"`
	Year  int    `json:"year"`
	Brand string `json:"brand"`
	Color string `json:"color"`
}

// ItemSnapshot is an immutable record of a product, service, or maintenance
// item frozen at the time of order analysis.
type ItemSnapshot struct {
	ID             string   `json:"id"`
	Kind           ItemKind `json:"kind"`
	Name           string   `json:"name"`
	Quantity       uint     `json:"qty"`
	UnitPriceCents int64    `json:"unit_price_cents"`
}
