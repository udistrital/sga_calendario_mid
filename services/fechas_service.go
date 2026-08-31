package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/udistrital/sga_calendario_mid/helpers"
	"github.com/udistrital/utils_oas/request"
)

func validarFechasPayload(payload map[string]interface{}, actual map[string]interface{}, contexto string) error {
	fechaInicio, tieneInicio := stringCampo(payload, "FechaInicio")
	fechaFin, tieneFin := stringCampo(payload, "FechaFin")
	if tieneInicio || tieneFin {
		if !tieneInicio && actual != nil {
			fechaInicio, tieneInicio = stringCampo(actual, "FechaInicio")
		}
		if !tieneFin && actual != nil {
			fechaFin, tieneFin = stringCampo(actual, "FechaFin")
		}
		if !tieneInicio || !tieneFin {
			return fmt.Errorf("%s: fecha inicio y fecha fin son requeridas para validar el rango", contexto)
		}
		if err := helpers.ValidarRangoFechas(fechaInicio, fechaFin); err != nil {
			return fmt.Errorf("%s: %w", contexto, err)
		}
	}

	if dependencia, ok := payload["DependenciaId"]; ok {
		if err := helpers.ValidarFechasDependenciaEvento(dependencia); err != nil {
			return fmt.Errorf("%s: %w", contexto, err)
		}
	}
	return nil
}

func validarFechasPayloadEventos(recurso string, id string, payload interface{}) error {
	payloadMap, ok := payload.(map[string]interface{})
	if !ok || !recursoSoportaFechas(recurso) || !payloadContieneFechas(payloadMap) {
		return nil
	}
	var actual map[string]interface{}
	if id != "" && payloadRequiereFechaActual(payloadMap) {
		if err := request.GetJson(beego.AppConfig.String("EventoService")+recurso+"/"+id, &actual); err != nil || actual == nil || actual["Type"] == "error" {
			return errors.New("no fue posible consultar el recurso para validar fechas")
		}
	}
	return validarFechasPayload(payloadMap, actual, recurso)
}

func recursoSoportaFechas(recurso string) bool {
	switch recurso {
	case "proceso", "calendario_evento", "calendario":
		return true
	default:
		return false
	}
}

func payloadContieneFechas(payload map[string]interface{}) bool {
	_, tieneInicio := payload["FechaInicio"]
	_, tieneFin := payload["FechaFin"]
	_, tieneDependencia := payload["DependenciaId"]
	return tieneInicio || tieneFin || tieneDependencia
}

func payloadRequiereFechaActual(payload map[string]interface{}) bool {
	_, tieneInicio := payload["FechaInicio"]
	_, tieneFin := payload["FechaFin"]
	return tieneInicio != tieneFin
}

func stringCampo(payload map[string]interface{}, campo string) (string, bool) {
	value, ok := payload[campo]
	if !ok || value == nil {
		return "", false
	}
	texto, ok := value.(string)
	if !ok {
		return "", false
	}
	texto = strings.TrimSpace(texto)
	return texto, texto != ""
}
