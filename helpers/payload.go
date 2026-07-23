package helpers

import "fmt"

func EventosPostResponseInvalid(resultado map[string]interface{}) bool {
	return resultado == nil || fmt.Sprintf("%v", resultado["System"]) == "map[]" || resultado["Id"] == nil || fmt.Sprintf("%v", resultado["Type"]) == "error" || fmt.Sprintf("%v", resultado["Status"]) == "400"
}

func NormalizeFechasTimeCalendarioEvento(recurso string, payload interface{}) {
	if recurso != "calendario_evento" {
		return
	}
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return
	}
	for _, campo := range []string{"FechaInicio", "FechaFin"} {
		if value, ok := payloadMap[campo]; ok && value != nil {
			payloadMap[campo] = FechaTimeParaCRUD(value)
		}
	}
}

func ActivoFromPayload(payload interface{}) (bool, bool) {
	m, ok := payload.(map[string]interface{})
	if !ok {
		return false, false
	}
	activo, ok := m["Activo"].(bool)
	return activo, ok
}

func IntSliceUnicos(valores []int) []int {
	vistos := make(map[int]bool)
	resultado := make([]int, 0)
	for _, valor := range valores {
		if valor <= 0 || vistos[valor] {
			continue
		}
		vistos[valor] = true
		resultado = append(resultado, valor)
	}
	return resultado
}
