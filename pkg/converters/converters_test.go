package converters

import (
	"net/url"
	"reflect"
	"testing"
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
