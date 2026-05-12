package converters

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/domain"
)

func TestMergeStructs(t *testing.T) {
	type TestStruct struct {
		Name   string
		Email  string
		Age    int
		Height float64
		Weight float64
		Type   uint
	}

	existing := TestStruct{
		Name:   "João",
		Email:  "joao@email.com",
		Age:    30,
		Height: 1.80,
		Weight: 70.0,
		Type:   1,
	}

	update := TestStruct{
		Name:   "João Silva",
		Email:  "",
		Age:    0,
		Height: 1.80,
		Type:   0,
	}

	result := MergeStructs(existing, update).(TestStruct)

	expected := TestStruct{
		Name:   "João Silva",
		Email:  "joao@email.com",
		Age:    30,
		Height: 1.80,
		Weight: 70.0,
		Type:   1,
	}

	if result != expected {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

func TestMergeStructsWithDomainUser(t *testing.T) {
	existing := domain.User{
		ID:        1,
		Role:      "user",
		PersonID:  1,
		DeletedAt: nil,
	}

	deletedAt := time.Now()
	update := domain.User{
		ID:        1,
		Role:      "admin",
		DeletedAt: &deletedAt,
	}

	result := MergeStructs(existing, update).(domain.User)

	expected := domain.User{
		ID:        1,
		Role:      "admin",
		PersonID:  1,
		DeletedAt: &deletedAt,
	}

	if result.Role != expected.Role {
		t.Errorf("Role: expected %s, got %s", expected.Role, result.Role)
	}
	if result.PersonID != expected.PersonID {
		t.Errorf("PersonID: expected %d, got %d", expected.PersonID, result.PersonID)
	}
	if (result.DeletedAt == nil) != (expected.DeletedAt == nil) {
		t.Errorf("DeletedAt: expected %v, got %v", expected.DeletedAt, result.DeletedAt)
	}
}

func TestParamsToMap(t *testing.T) {
	params := url.Values{
		"name":  {"Jon Snow"},
		"email": {"jon@winterfell.com"},
		"age":   {"30"},
	}

	result := ParamsToMap(params)

	expected := map[string]interface{}{
		"name":  "Jon Snow",
		"email": "jon@winterfell.com",
		"age":   "30",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}
