package helpers

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const formatoFechaHoraCalendario = "2006-01-02T15:04:05"

var GMTMinus5Location = time.FixedZone("GMT-5", -5*60*60)

func ParseFecha(fecha string) (time.Time, error) {
	fecha = limpiarFechaGo(fecha)
	if fechaLocal, ok := fechaHoraLocal(fecha); ok {
		parsed, err := time.ParseInLocation(formatoFechaHoraCalendario, fechaLocal, GMTMinus5Location)
		if err == nil {
			return parsed, nil
		}
	}
	formatosConZona := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05Z07:00", "2006-01-02 15:04:05 -0700"}
	formatosSinZona := []string{"2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02"}
	var ultimo error
	for _, formato := range formatosConZona {
		parsed, err := time.Parse(formato, fecha)
		if err == nil {
			return parsed.In(GMTMinus5Location), nil
		}
		ultimo = err
	}
	for _, formato := range formatosSinZona {
		parsed, err := time.ParseInLocation(formato, fecha, GMTMinus5Location)
		if err == nil {
			return parsed, nil
		}
		ultimo = err
	}
	return time.Time{}, ultimo
}

func fechaHoraLocal(fecha string) (string, bool) {
	fecha = strings.TrimSpace(fecha)
	if len(fecha) < len(formatoFechaHoraCalendario) {
		return "", false
	}
	fecha = fecha[:len(formatoFechaHoraCalendario)]
	if fecha[10] == ' ' {
		fecha = strings.Replace(fecha, " ", "T", 1)
	}
	if fecha[10] != 'T' {
		return "", false
	}
	return fecha, true
}

func limpiarFechaGo(fecha string) string {
	partes := strings.Fields(fecha)
	if len(partes) == 4 && strings.HasPrefix(partes[2], "+") && strings.HasPrefix(partes[3], "+") {
		return strings.Join(partes[:3], " ")
	}
	return fecha
}

func FechaTimeParaCRUD(value interface{}) string {
	fecha := limpiarFechaGo(strings.TrimSpace(fmt.Sprintf("%v", value)))
	if fecha == "" || fecha == "<nil>" {
		return fecha
	}
	if fechaLocal, ok := fechaHoraLocal(fecha); ok {
		return fechaLocal + "Z"
	}
	parsed, err := ParseFecha(fecha)
	if err != nil {
		return fecha
	}
	return parsed.In(GMTMinus5Location).Format("2006-01-02T15:04:05") + "Z"
}

func FechaTimeParaModelo(value interface{}) (time.Time, error) {
	fecha := strings.TrimSuffix(FechaTimeParaCRUD(value), "Z")
	return time.ParseInLocation(formatoFechaHoraCalendario, fecha, time.UTC)
}

func FechaDependenciaActividad(value interface{}) string {
	fecha := strings.TrimSpace(fmt.Sprintf("%v", value))
	if fecha == "" || fecha == "<nil>" {
		return ""
	}
	fechaCRUD := FechaTimeParaCRUD(fecha)
	if fechaCRUD != fecha {
		return strings.TrimSuffix(fechaCRUD, "Z")
	}
	fecha = strings.Replace(fecha, " ", "T", 1)
	if len(fecha) >= len(formatoFechaHoraCalendario) {
		return fecha[:len(formatoFechaHoraCalendario)]
	}
	return fecha
}

func FormatFechaGMTMinus5(fecha time.Time) string {
	fechaLocal := time.Date(fecha.Year(), fecha.Month(), fecha.Day(), fecha.Hour(), fecha.Minute(), fecha.Second(), fecha.Nanosecond(), GMTMinus5Location)
	return fechaLocal.Format("2006-01-02T15:04:05-07:00")
}

func ValidarRangoFechas(fechaInicio string, fechaFin string) error {
	fechaInicio = strings.TrimSpace(fechaInicio)
	fechaFin = strings.TrimSpace(fechaFin)
	if fechaInicio == "" || fechaFin == "" {
		return errors.New("fecha inicio y fecha fin son requeridas")
	}
	inicio, err := ParseFecha(fechaInicio)
	if err != nil {
		return errors.New("fecha inicio inválida")
	}
	fin, err := ParseFecha(fechaFin)
	if err != nil {
		return errors.New("fecha fin inválida")
	}
	if fin.Before(inicio) {
		return errors.New("la fecha fin no puede ser menor que la fecha inicio")
	}
	return nil
}
