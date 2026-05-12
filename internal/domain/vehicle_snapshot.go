package domain

type Vehicle struct {
	ID    uint   `json:"id,omitempty"`
	Plate string `json:"plate,omitempty"`
	Name  string `json:"name"`
	Year  int    `json:"year"`
	Brand string `json:"brand"`
	Color string `json:"color"`
}
