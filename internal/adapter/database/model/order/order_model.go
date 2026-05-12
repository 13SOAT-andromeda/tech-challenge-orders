package order

import (
	"time"

	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/database/model/company"
	customerVehicle "github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/database/model/customer_vehicle"
	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/database/model/employee"
	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/database/model/order_maintenance"
	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/database/model/order_product"
	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/domain"
	"gorm.io/gorm"
)

type Model struct {
	gorm.Model
	DateIn            time.Time `gorm:"not null"`
	DateOut           *time.Time
	DateApproved      *time.Time
	DateRejected      *time.Time
	LastStatusAt      *time.Time
	Status            string
	VehicleKilometers int
	Note              *string
	DiagnosticNote    *string
	Price             *float64
	EmployeeID        uint
	CustomerVehicleID uint
	CompanyID         uint
	Employee          employee.Model            `gorm:"foreignKey:EmployeeID;references:ID"`
	Company           company.Model             `gorm:"foreignKey:CompanyID;references:ID"`
	CustomerVehicle   customerVehicle.Model     `gorm:"foreignKey:CustomerVehicleID;references:ID"`
	OrderProducts     []order_product.Model     `gorm:"foreignKey:OrderId;references:ID"`
	OrderMaintenances []order_maintenance.Model `gorm:"foreignKey:OrderId;references:ID"`
}

func (*Model) TableName() string {
	return "Orders"
}

func (m *Model) ToDomain() *domain.Order {
	var products []domain.Product
	for _, orderProduct := range m.OrderProducts {
		if product := orderProduct.Product.ToDomain(); product != nil {
			quantity := orderProduct.Quantity
			product.Quantity = &quantity
			product.Stock = nil
			products = append(products, *product)
		}
	}

	var maintenances []domain.Maintenance
	for _, orderMaintenance := range m.OrderMaintenances {
		if maintenance := orderMaintenance.Maintenance.ToDomain(); maintenance != nil {
			maintenances = append(maintenances, *maintenance)
		}
	}

	var productsPtr *[]domain.Product
	if len(products) > 0 {
		productsPtr = &products
	}

	var maintenancesPtr *[]domain.Maintenance
	if len(maintenances) > 0 {
		maintenancesPtr = &maintenances
	}

	var vehicle *domain.Vehicle
	if m.CustomerVehicle.Vehicle.ID != 0 {
		fullVehicle := m.CustomerVehicle.Vehicle.ToDomain()
		if fullVehicle != nil {
			vehicle = &domain.Vehicle{
				Plate: fullVehicle.Plate,
				Name:  fullVehicle.Name,
				Year:  fullVehicle.Year,
				Brand: fullVehicle.Brand,
				Color: fullVehicle.Color,
			}
		}
	}

	last := m.LastStatusAt
	if last == nil {
		t := m.DateIn
		last = &t
	}

	return &domain.Order{
		ID:                m.ID,
		DateIn:            m.DateIn,
		DateOut:           m.DateOut,
		DateApproved:      m.DateApproved,
		DateRejected:      m.DateRejected,
		LastStatusAt:      last,
		Status:            domain.OrderStatus(m.Status),
		VehicleKilometers: m.VehicleKilometers,
		Note:              m.Note,
		DiagnosticNote:    m.DiagnosticNote,
		Price:             m.Price,
		CustomerVehicleID: m.CustomerVehicleID,
		EmployeeID:        m.EmployeeID,
		CompanyID:         m.CompanyID,
		Vehicle:           vehicle,
		Products:          productsPtr,
		Maintenances:      maintenancesPtr,
	}
}

func (m *Model) FromDomain(d *domain.Order) {
	m.ID = d.ID
	m.DateIn = d.DateIn
	m.DateOut = d.DateOut
	m.DateApproved = d.DateApproved
	m.DateRejected = d.DateRejected
	if d.LastStatusAt != nil {
		m.LastStatusAt = d.LastStatusAt
	} else if d.ID == 0 {
		t := d.DateIn
		m.LastStatusAt = &t
	}
	m.Status = string(d.Status)
	m.VehicleKilometers = d.VehicleKilometers
	m.Note = d.Note
	m.DiagnosticNote = d.DiagnosticNote
	m.Price = d.Price
	m.EmployeeID = d.EmployeeID
	m.CustomerVehicleID = d.CustomerVehicleID
	m.CompanyID = d.CompanyID
}
