package converters

import (
	"net/url"
	"reflect"
)

func ParamsToMap(params url.Values) map[string]interface{} {
	paramsMap := make(map[string]interface{})
	for key, value := range params {
		paramsMap[key] = value[0]
	}

	return paramsMap
}

// MergeStructs faz merge entre dois structs, preservando valores do target quando source é zero value
// Para campos específicos como bool, considera apenas string vazia, 0 para números e nil para ponteiros como "não fornecido"
func MergeStructs(target, source interface{}) interface{} {
	targetValue := reflect.ValueOf(target)
	sourceValue := reflect.ValueOf(source)

	// Se target é ponteiro, pega o valor apontado
	if targetValue.Kind() == reflect.Ptr {
		targetValue = targetValue.Elem()
	}

	// Se source é ponteiro, pega o valor apontado
	if sourceValue.Kind() == reflect.Ptr {
		sourceValue = sourceValue.Elem()
	}

	// Cria uma nova instância do tipo target
	result := reflect.New(targetValue.Type()).Elem()

	// Copia todos os campos do target para o result
	for i := 0; i < targetValue.NumField(); i++ {
		fieldName := targetValue.Type().Field(i).Name
		targetField := targetValue.Field(i)
		sourceField := sourceValue.FieldByName(fieldName)

		if sourceField.IsValid() && !isZeroValueForMerge(sourceField) {
			// Se o campo no source não é zero value, usa o valor do source
			result.Field(i).Set(sourceField)
		} else {
			// Se o campo no source é zero value, mantém o valor do target
			result.Field(i).Set(targetField)
		}
	}

	return result.Interface()
}

// isZeroValueForMerge verifica se um valor deve ser considerado como "não fornecido" para merge
// Para bool, considera apenas quando é o zero value do tipo (false), mas isso pode ser problemático
// Para este caso específico, vamos considerar apenas string vazia, 0 para números e nil para ponteiros
func isZeroValueForMerge(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		// Para bool, não consideramos false como "não fornecido" pois é um valor válido
		// Retornamos false para que o valor seja sempre usado
		return false
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
		return v.IsNil()
	default:
		return v.IsZero()
	}
}
