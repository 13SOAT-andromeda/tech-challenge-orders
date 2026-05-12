package domain

import (
	"time"

	"errors"
)

var (
	ErrOrderNotFound = errors.New("order not found")
)

type Order struct {
	ID                string        `json:"id"`
	Status            Status        `json:"status"`
	DateIn            time.Time     `json:"date_in"`
	DateOut           *time.Time    `json:"date_out"`
	DateApproved      *time.Time    `json:"date_approved"`
	DateRejected      *time.Time    `json:"date_rejected"`
	LastStatusAt      *time.Time    `json:"last_status_at,omitempty"`
	VehicleKilometers int           `json:"vehicle_kilometers"`
	Note              *string       `json:"note"`
	DiagnosticNote    *string       `json:"diagnostic_note"`
	Price             *float64      `json:"price"`
	CustomerVehicleID string        `json:"customer_vehicle_id"`
	EmployeeID        string        `json:"employee_id"`
	CompanyID         string        `json:"company_id"`
	Vehicle           *Vehicle      `json:"vehicle,omitempty"`
	Items             *[]OrderItems `json:"items,omitempty"`
}
