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

	ID      string `dynamodbav:"ID"`
	Status  string `dynamodbav:"Status"`
	Version int    `dynamodbav:"Version"`

	DateIn       time.Time  `dynamodbav:"DateIn"`
	DateOut      *time.Time `dynamodbav:"DateOut,omitempty"`
	DateApproved *time.Time `dynamodbav:"DateApproved,omitempty"`
	DateRejected *time.Time `dynamodbav:"DateRejected,omitempty"`
	PaymentDate  *time.Time `dynamodbav:"PaymentDate,omitempty"`
	LastStatusAt *time.Time `dynamodbav:"LastStatusAt,omitempty"`
	CreatedAt    time.Time  `dynamodbav:"CreatedAt"`
	UpdatedAt    time.Time  `dynamodbav:"UpdatedAt"`

	VehicleKilometers int     `dynamodbav:"VehicleKilometers"`
	Note              *string `dynamodbav:"Note,omitempty"`
	DiagnosticNote    *string `dynamodbav:"DiagnosticNote,omitempty"`
	PriceCents        *int64  `dynamodbav:"PriceCents,omitempty"`

	CustomerVehicleID int64 `dynamodbav:"CustomerVehicleID"`
	CustomerID        int64 `dynamodbav:"CustomerID,omitempty"`
	EmployeeID        int64 `dynamodbav:"EmployeeID,omitempty"`
	CompanyID         int64 `dynamodbav:"CompanyID"`

	Vehicle *VehicleAV       `dynamodbav:"Vehicle,omitempty"`
	Items   []ItemSnapshotAV `dynamodbav:"Items,omitempty"`
	History []HistoryEntryAV `dynamodbav:"History,omitempty"`
}

type VehicleAV struct {
	ID    int64  `dynamodbav:"ID"`
	Plate string `dynamodbav:"Plate"`
	Name  string `dynamodbav:"Name"`
	Year  int    `dynamodbav:"Year"`
	Brand string `dynamodbav:"Brand"`
	Color string `dynamodbav:"Color"`
}

type ItemSnapshotAV struct {
	ID             int64  `dynamodbav:"ID"`
	Kind           string `dynamodbav:"Kind"`
	Name           string `dynamodbav:"Name"`
	Quantity       uint   `dynamodbav:"Quantity"`
	UnitPriceCents int64  `dynamodbav:"UnitPriceCents"`
}

type HistoryEntryAV struct {
	From   string    `dynamodbav:"From"`
	To     string    `dynamodbav:"To"`
	At     time.Time `dynamodbav:"At"`
	Actor  string    `dynamodbav:"Actor"`
	Reason string    `dynamodbav:"Reason,omitempty"`
}

func PrimaryKey(orderID string) (pk, sk string) {
	return fmt.Sprintf("ORDER#%s", orderID), SKMeta
}

func statusGSIKeys(status domain.Status, dateIn time.Time, orderID string) (pk, sk string) {
	return fmt.Sprintf("STATUS#%s", status),
		fmt.Sprintf("%s#%s", dateIn.UTC().Format(time.RFC3339Nano), orderID)
}

func customerVehicleGSIKeys(customerVehicleID int64, dateIn time.Time, orderID string) (pk, sk string) {
	return fmt.Sprintf("CUSTOMERVEHICLE#%d", customerVehicleID),
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

	items := make([]ItemSnapshotAV, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, ItemSnapshotAV{
			ID:             it.ID,
			Kind:           string(it.Kind),
			Name:           it.Name,
			Quantity:       it.Quantity,
			UnitPriceCents: it.UnitPriceCents,
		})
	}

	history := make([]HistoryEntryAV, 0, len(o.History))
	for _, h := range o.History {
		history = append(history, HistoryEntryAV{
			From:   string(h.From),
			To:     string(h.To),
			At:     h.At,
			Actor:  h.Actor,
			Reason: h.Reason,
		})
	}

	return &OrderItem{
		PK:     pk,
		SK:     sk,
		GSI1PK: gsi1pk,
		GSI1SK: gsi1sk,
		GSI2PK: gsi2pk,
		GSI2SK: gsi2sk,
		Type:   TypeOrder,

		ID:      o.ID,
		Status:  string(o.Status),
		Version: o.Version,

		DateIn:       o.DateIn,
		DateOut:      o.DateOut,
		DateApproved: o.DateApproved,
		DateRejected: o.DateRejected,
		PaymentDate:  o.PaymentDate,
		LastStatusAt: o.LastStatusAt,
		CreatedAt:    o.CreatedAt,
		UpdatedAt:    o.UpdatedAt,

		VehicleKilometers: o.VehicleKilometers,
		Note:              o.Note,
		DiagnosticNote:    o.DiagnosticNote,
		PriceCents:        o.PriceCents,

		CustomerVehicleID: o.CustomerVehicleID,
		CustomerID:        o.CustomerID,
		EmployeeID:        o.EmployeeID,
		CompanyID:         o.CompanyID,

		Vehicle: vehicle,
		Items:   items,
		History: history,
	}
}

func (it *OrderItem) ToDomain() *domain.Order {
	var vehicle *domain.VehicleSnapshot
	if it.Vehicle != nil {
		vehicle = &domain.VehicleSnapshot{
			ID:    it.Vehicle.ID,
			Plate: it.Vehicle.Plate,
			Name:  it.Vehicle.Name,
			Year:  it.Vehicle.Year,
			Brand: it.Vehicle.Brand,
			Color: it.Vehicle.Color,
		}
	}

	items := make([]domain.ItemSnapshot, 0, len(it.Items))
	for _, av := range it.Items {
		items = append(items, domain.ItemSnapshot{
			ID:             av.ID,
			Kind:           domain.ItemKind(av.Kind),
			Name:           av.Name,
			Quantity:       av.Quantity,
			UnitPriceCents: av.UnitPriceCents,
		})
	}

	history := make([]domain.HistoryEntry, 0, len(it.History))
	for _, av := range it.History {
		history = append(history, domain.HistoryEntry{
			From:   domain.Status(av.From),
			To:     domain.Status(av.To),
			At:     av.At,
			Actor:  av.Actor,
			Reason: av.Reason,
		})
	}

	return &domain.Order{
		ID:      it.ID,
		Status:  domain.Status(it.Status),
		Version: it.Version,

		DateIn:       it.DateIn,
		DateOut:      it.DateOut,
		DateApproved: it.DateApproved,
		DateRejected: it.DateRejected,
		PaymentDate:  it.PaymentDate,
		LastStatusAt: it.LastStatusAt,
		CreatedAt:    it.CreatedAt,
		UpdatedAt:    it.UpdatedAt,

		VehicleKilometers: it.VehicleKilometers,
		Note:              it.Note,
		DiagnosticNote:    it.DiagnosticNote,
		PriceCents:        it.PriceCents,

		CustomerVehicleID: it.CustomerVehicleID,
		CustomerID:        it.CustomerID,
		EmployeeID:        it.EmployeeID,
		CompanyID:         it.CompanyID,

		Vehicle: vehicle,
		Items:   items,
		History: history,
	}
}
