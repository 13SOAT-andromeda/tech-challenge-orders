package model

import (
	"fmt"
	"time"

	"github.com/13SOAT-andromeda/tech-challenge-orders/internal/domain"
)

const (
	SKMeta    = "META"
	TypeOrder = "Order"
)

type OrderItem struct {
	PK     string `dynamodbav:"PK"`
	SK     string `dynamodbav:"SK"`
	GSI1PK string `dynamodbav:"GSI1PK"`
	GSI1SK string `dynamodbav:"GSI1SK"`
	GSI2PK string `dynamodbav:"GSI2PK"`
	GSI2SK string `dynamodbav:"GSI2SK"`
	Type   string `dynamodbav:"Type"`

	ID                string         `dynamodbav:"ID"`
	Status            string         `dynamodbav:"Status"`
	DateIn            time.Time      `dynamodbav:"DateIn"`
	DateOut           *time.Time     `dynamodbav:"DateOut,omitempty"`
	DateApproved      *time.Time     `dynamodbav:"DateApproved,omitempty"`
	DateRejected      *time.Time     `dynamodbav:"DateRejected,omitempty"`
	LastStatusAt      *time.Time     `dynamodbav:"LastStatusAt,omitempty"`
	VehicleKilometers int            `dynamodbav:"VehicleKilometers"`
	Note              *string        `dynamodbav:"Note,omitempty"`
	DiagnosticNote    *string        `dynamodbav:"DiagnosticNote,omitempty"`
	Price             *float64       `dynamodbav:"Price,omitempty"`
	CustomerVehicleID string         `dynamodbav:"CustomerVehicleID"`
	EmployeeID        string         `dynamodbav:"EmployeeID,omitempty"`
	CompanyID         string         `dynamodbav:"CompanyID"`
	Vehicle           *VehicleAV     `dynamodbav:"Vehicle,omitempty"`
	Items             []OrderItemsAV `dynamodbav:"Items,omitempty"`
}

type VehicleAV struct {
	ID    uint   `dynamodbav:"ID,omitempty"`
	Plate string `dynamodbav:"Plate,omitempty"`
	Name  string `dynamodbav:"Name"`
	Year  int    `dynamodbav:"Year"`
	Brand string `dynamodbav:"Brand"`
	Color string `dynamodbav:"Color"`
}

type ItemAV struct {
	ID    uint   `dynamodbav:"ID"`
	Name  string `dynamodbav:"Name"`
	Price int64  `dynamodbav:"Price"`
	Type  string `dynamodbav:"Type"`
}

type OrderItemsAV struct {
	Quantity uint   `dynamodbav:"Quantity"`
	ItemID   uint   `dynamodbav:"ItemID"`
	Item     ItemAV `dynamodbav:"Item"`
}

func PrimaryKey(orderID string) (pk, sk string) {
	return fmt.Sprintf("ORDER#%s", orderID), SKMeta
}

func statusGSIKeys(status domain.Status, dateIn time.Time, orderID string) (pk, sk string) {
	return fmt.Sprintf("STATUS#%s", status),
		fmt.Sprintf("%s#%s", dateIn.UTC().Format(time.RFC3339Nano), orderID)
}

func customerVehicleGSIKeys(customerVehicleID string, dateIn time.Time, orderID string) (pk, sk string) {
	return fmt.Sprintf("CUSTOMERVEHICLE#%s", customerVehicleID),
		fmt.Sprintf("%s#%s", dateIn.UTC().Format(time.RFC3339Nano), orderID)
}

func FromDomain(o *domain.Order) *OrderItem {
	pk, sk := PrimaryKey(o.ID)
	gsi1pk, gsi1sk := statusGSIKeys(o.Status, o.DateIn, o.ID)
	gsi2pk, gsi2sk := customerVehicleGSIKeys(o.CustomerVehicleID, o.DateIn, o.ID)

	var vehicle *VehicleAV
	if o.Vehicle != nil {
		vehicle = &VehicleAV{
			ID:    o.Vehicle.ID,
			Plate: o.Vehicle.Plate,
			Name:  o.Vehicle.Name,
			Year:  o.Vehicle.Year,
			Brand: o.Vehicle.Brand,
			Color: o.Vehicle.Color,
		}
	}

	var items []OrderItemsAV
	if o.Items != nil {
		items = make([]OrderItemsAV, 0, len(*o.Items))
		for _, oi := range *o.Items {
			items = append(items, OrderItemsAV{
				Quantity: oi.Quantity,
				ItemID:   oi.ItemId,
				Item: ItemAV{
					ID:    oi.Item.ID,
					Name:  oi.Item.Name,
					Price: oi.Item.Price,
					Type:  oi.Item.Type,
				},
			})
		}
	}

	return &OrderItem{
		PK:     pk,
		SK:     sk,
		GSI1PK: gsi1pk,
		GSI1SK: gsi1sk,
		GSI2PK: gsi2pk,
		GSI2SK: gsi2sk,
		Type:   TypeOrder,

		ID:                o.ID,
		Status:            string(o.Status),
		DateIn:            o.DateIn,
		DateOut:           o.DateOut,
		DateApproved:      o.DateApproved,
		DateRejected:      o.DateRejected,
		LastStatusAt:      o.LastStatusAt,
		VehicleKilometers: o.VehicleKilometers,
		Note:              o.Note,
		DiagnosticNote:    o.DiagnosticNote,
		Price:             o.Price,
		CustomerVehicleID: o.CustomerVehicleID,
		EmployeeID:        o.EmployeeID,
		CompanyID:         o.CompanyID,
		Vehicle:           vehicle,
		Items:             items,
	}
}

func (it *OrderItem) ToDomain() *domain.Order {
	var vehicle *domain.Vehicle
	if it.Vehicle != nil {
		vehicle = &domain.Vehicle{
			ID:    it.Vehicle.ID,
			Plate: it.Vehicle.Plate,
			Name:  it.Vehicle.Name,
			Year:  it.Vehicle.Year,
			Brand: it.Vehicle.Brand,
			Color: it.Vehicle.Color,
		}
	}

	var items *[]domain.OrderItems
	if it.Items != nil {
		list := make([]domain.OrderItems, 0, len(it.Items))
		for _, oi := range it.Items {
			list = append(list, domain.OrderItems{
				Quantity: oi.Quantity,
				ItemId:   oi.ItemID,
				Item: domain.Item{
					ID:    oi.Item.ID,
					Name:  oi.Item.Name,
					Price: oi.Item.Price,
					Type:  oi.Item.Type,
				},
			})
		}
		items = &list
	}

	return &domain.Order{
		ID:                it.ID,
		Status:            domain.Status(it.Status),
		DateIn:            it.DateIn,
		DateOut:           it.DateOut,
		DateApproved:      it.DateApproved,
		DateRejected:      it.DateRejected,
		LastStatusAt:      it.LastStatusAt,
		VehicleKilometers: it.VehicleKilometers,
		Note:              it.Note,
		DiagnosticNote:    it.DiagnosticNote,
		Price:             it.Price,
		CustomerVehicleID: it.CustomerVehicleID,
		EmployeeID:        it.EmployeeID,
		CompanyID:         it.CompanyID,
		Vehicle:           vehicle,
		Items:             items,
	}
}
