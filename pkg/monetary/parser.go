package monetary

import (
	"fmt"
	"math"
	"strings"
)

func ConvertToMinorUnitInt64(amount float64, decimals uint) int64 {
	scaleFactor := math.Pow(10, float64(decimals))

	// Scale and round
	scaled := amount * scaleFactor
	rounded := math.Round(scaled)

	// Convert the rounded float64 to int64.
	return int64(rounded)
}

func FormatPtBrCurrency(value float64) string {
	// Converte para string com 2 casas decimais
	formatted := fmt.Sprintf("%.2f", value)

	// Separa a parte inteira e decimal
	parts := strings.Split(formatted, ".")
	integerPart := parts[0]
	decimalPart := parts[1]

	// Adiciona separadores de milhares (pontos)
	if len(integerPart) > 3 {
		var result strings.Builder
		for i, digit := range integerPart {
			if i > 0 && (len(integerPart)-i)%3 == 0 {
				result.WriteString(".")
			}
			result.WriteRune(digit)
		}
		integerPart = result.String()
	}

	return fmt.Sprintf("R$ %s,%s", integerPart, decimalPart)
}
