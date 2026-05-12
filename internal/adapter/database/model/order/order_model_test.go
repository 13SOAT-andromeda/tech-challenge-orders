package order

import (
	"testing"
	"time"

	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/database/model/company"
	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/database/model/customer_vehicle"
	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/database/model/employee"
	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/domain"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestOrderModelInitialization(t *testing.T) {
	now := time.Now()
	vehicleKm := 10000
	note := "Test note"
	diagnosticNote := "Test diagnostic note"
	price := 150.75

	lastAt := now.Add(-time.Hour)
	o := Model{
		DateIn:            now,
		DateOut:           &now,
		LastStatusAt:      &lastAt,
		Status:            "pending",
		VehicleKilometers: vehicleKm,
		Note:              &note,
		DiagnosticNote:    &diagnosticNote,
		Price:             &price,
	}

	assert.NotNil(t, o)
	assert.Equal(t, now, o.DateIn)
	assert.Equal(t, &now, o.DateOut)
	assert.Equal(t, &lastAt, o.LastStatusAt)
	assert.Equal(t, "pending", o.Status)
	assert.Equal(t, vehicleKm, o.VehicleKilometers)
	assert.Equal(t, &note, o.Note)
	assert.Equal(t, &diagnosticNote, o.DiagnosticNote)
	assert.Equal(t, &price, o.Price)
}

func TestOrderModel_ToFromDomain(t *testing.T) {
	now := time.Now()
	vehicleKm := 10000
	note := "Test note"
	diagnosticNote := "Test diagnostic note"
	price := 150.75

	modelOrder := Model{
		DateIn:            now,
		DateOut:           &now,
		Status:            "Recebida",
		VehicleKilometers: vehicleKm,
		Note:              &note,
		DiagnosticNote:    &diagnosticNote,
		Price:             &price,
		Employee:          employee.Model{},
		CustomerVehicle:   customer_vehicle.Model{},
		Company:           company.Model{},
	}

	domainOrder := modelOrder.ToDomain()

	assert.Equal(t, modelOrder.ID, domainOrder.ID)
	assert.Equal(t, modelOrder.DateIn, domainOrder.DateIn)
	assert.Equal(t, modelOrder.DateOut, domainOrder.DateOut)

	assert.NotNil(t, domainOrder.Status)
	assert.Equal(t, modelOrder.Status, string(domainOrder.Status))

	assert.Equal(t, modelOrder.VehicleKilometers, domainOrder.VehicleKilometers)
	assert.Equal(t, modelOrder.Note, domainOrder.Note)
	assert.Equal(t, modelOrder.DiagnosticNote, domainOrder.DiagnosticNote)
	assert.Equal(t, modelOrder.Price, domainOrder.Price)
}

func TestOrderModel_ToDomain_LastStatusAt_nil_uses_DateIn(t *testing.T) {
	now := time.Date(2025, 3, 19, 12, 0, 0, 0, time.UTC)
	m := Model{
		DateIn:       now,
		LastStatusAt: nil,
		Status:       "Recebida",
	}

	d := m.ToDomain()
	assert.NotNil(t, d.LastStatusAt)
	assert.True(t, d.LastStatusAt.Equal(now), "fallback deve ser DateIn quando LastStatusAt é nil no model")
}

func TestOrderModel_ToDomain_LastStatusAt_set(t *testing.T) {
	dateIn := time.Date(2025, 3, 19, 10, 0, 0, 0, time.UTC)
	last := time.Date(2025, 3, 19, 15, 30, 0, 0, time.UTC)
	m := Model{
		DateIn:       dateIn,
		LastStatusAt: &last,
		Status:       "Em execução",
	}

	d := m.ToDomain()
	assert.NotNil(t, d.LastStatusAt)
	assert.True(t, d.LastStatusAt.Equal(last))
	assert.False(t, d.LastStatusAt.Equal(dateIn))
}

func TestOrderModel_FromDomain_LastStatusAt_create(t *testing.T) {
	dateIn := time.Date(2025, 3, 19, 9, 0, 0, 0, time.UTC)
	d := &domain.Order{
		DateIn: dateIn,
		Status: domain.RECEIVED,
	}

	var m Model
	m.FromDomain(d)

	assert.NotNil(t, m.LastStatusAt)
	assert.True(t, m.LastStatusAt.Equal(dateIn))
}

func TestOrderModel_FromDomain_LastStatusAt_explicit(t *testing.T) {
	dateIn := time.Date(2025, 3, 19, 9, 0, 0, 0, time.UTC)
	last := time.Date(2025, 3, 19, 16, 0, 0, 0, time.UTC)
	d := &domain.Order{
		ID:           1,
		DateIn:       dateIn,
		LastStatusAt: &last,
		Status:       domain.IN_PROGRESS,
	}

	var m Model
	m.FromDomain(d)

	assert.NotNil(t, m.LastStatusAt)
	assert.True(t, m.LastStatusAt.Equal(last))
}

func TestOrderModel_FromDomain_LastStatusAt_update_sem_domain_nao_sobrescreve(t *testing.T) {
	dateIn := time.Date(2025, 3, 19, 9, 0, 0, 0, time.UTC)
	existing := time.Date(2025, 3, 19, 14, 0, 0, 0, time.UTC)
	m := Model{
		Model:        gorm.Model{ID: 5},
		DateIn:       dateIn,
		LastStatusAt: &existing,
		Status:       string(domain.RECEIVED),
	}

	d := &domain.Order{
		ID:           5,
		DateIn:       dateIn,
		LastStatusAt: nil,
		Status:       domain.IN_ANALYSIS,
	}
	m.FromDomain(d)

	assert.NotNil(t, m.LastStatusAt)
	assert.True(t, m.LastStatusAt.Equal(existing), "update com LastStatusAt nil no domain não deve limpar o valor já carregado no model")
}
