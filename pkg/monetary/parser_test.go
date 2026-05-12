package monetary

import "testing"

func TestConvertToMinorUnitInt64_BrazilianExamples(t *testing.T) {
	cases := []struct {
		name     string
		amount   float64
		decimals uint
		want     int64
	}{
		{"R$ 1.234,56", 1234.56, 2, 123456},
		{"R$ 0,99", 0.99, 2, 99},
		{"R$ 0,005 (3 decimals)", 0.005, 3, 5},
		{"R$ 0,00 (zero)", 0.0, 2, 0},
		{"R$ 10,00", 10.0, 2, 1000},
		{"R$ 999,99", 999.99, 2, 99999},
		{"R$ 0,01", 0.01, 2, 1},
		{"R$ 1.000,00", 1000.0, 2, 100000},
		{"-R$ 12,34 (negative)", -12.34, 2, -1234},
		{"R$ 0,999 (rounding up)", 0.999, 2, 100},
		{"R$ 0,994 (rounding down)", 0.994, 2, 99},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ConvertToMinorUnitInt64(tc.amount, tc.decimals)
			if got != tc.want {
				t.Fatalf("%s: ConvertToMinorUnitInt64(%v, %d) = %d; want %d", tc.name, tc.amount, tc.decimals, got, tc.want)
			}
		})
	}
}

func TestFormatPtBrCurrency(t *testing.T) {
	cases := []struct {
		name  string
		value float64
		want  string
	}{
		{"Valor zero", 0.0, "R$ 0,00"},
		{"Valor pequeno", 0.99, "R$ 0,99"},
		{"Valor sem milhares", 123.45, "R$ 123,45"},
		{"Valor com milhares", 1234.56, "R$ 1.234,56"},
		{"Valor grande", 1234567.89, "R$ 1.234.567,89"},
		{"Valor muito grande", 1234567890.12, "R$ 1.234.567.890,12"},
		{"Valor negativo", -1234.56, "R$ -1.234,56"},
		{"Valor exato", 10.0, "R$ 10,00"},
		{"Centavos apenas", 0.01, "R$ 0,01"},
		{"Milhar exato", 1000.0, "R$ 1.000,00"},
		{"Milhão", 1000000.0, "R$ 1.000.000,00"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatPtBrCurrency(tc.value)
			if got != tc.want {
				t.Fatalf("%s: FormatPtBrCurrency(%v) = %s; want %s", tc.name, tc.value, got, tc.want)
			}
		})
	}
}
