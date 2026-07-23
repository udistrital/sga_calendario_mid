package models

import (
	"encoding/json"
	"strings"
	"time"
)

type SolicitudExtensionActividad struct {
	FechaFin     time.Time     `json:"FechaFin"`
	DocumentoId  interface{}   `json:"DocumentoId"`
	Descripcion  string        `json:"Descripcion"`
	Dependencias []interface{} `json:"Dependencias"`
}

type SolicitudExtensionActividadRequest struct {
	FechaFin     string `json:"FechaFin"`
	DocumentoId  int    `json:"DocumentoId"`
	Descripcion  string `json:"Descripcion"`
	Dependencias []int  `json:"Dependencias"`
}

func (s *SolicitudExtensionActividad) UnmarshalJSON(data []byte) error {
	type solicitudAlias SolicitudExtensionActividad
	aux := struct {
		FechaFin interface{} `json:"FechaFin"`
		*solicitudAlias
	}{solicitudAlias: (*solicitudAlias)(s)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	fecha, err := parseFechaSolicitud(aux.FechaFin)
	if err != nil {
		return err
	}
	s.FechaFin = fecha
	return nil
}

func parseFechaSolicitud(value interface{}) (time.Time, error) {
	if value == nil {
		return time.Time{}, nil
	}
	fecha, ok := value.(string)
	if !ok {
		return time.Time{}, nil
	}
	fecha = strings.TrimSpace(fecha)
	if fecha == "" {
		return time.Time{}, nil
	}
	if parsed, err := time.Parse(time.RFC3339Nano, fecha); err == nil {
		return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), parsed.Hour(), parsed.Minute(), parsed.Second(), parsed.Nanosecond(), time.UTC), nil
	}
	if len(fecha) >= len("2006-01-02T15:04:05") {
		fecha = fecha[:len("2006-01-02T15:04:05")]
		fecha = strings.Replace(fecha, " ", "T", 1)
	}
	return time.Parse(time.RFC3339, fecha+"Z")
}

type CalendarioEventoExtensionPayload struct {
	Id                 int         `json:"Id,omitempty"`
	CalendarioEventoId RelacionID  `json:"CalendarioEventoId"`
	FechaFin           time.Time   `json:"FechaFin"`
	DocumentoId        interface{} `json:"DocumentoId"`
	Descripcion        string      `json:"Descripcion"`
	NumeroExtension    int         `json:"NumeroExtension,omitempty"`
	Activo             bool        `json:"Activo"`
	FechaCreacion      time.Time   `json:"FechaCreacion,omitempty"`
}

type CalendarioEventoExtensionProgramaPayload struct {
	Id                          int         `json:"Id,omitempty"`
	CalendarioEventoId          RelacionID  `json:"CalendarioEventoId"`
	CalendarioEventoExtensionId RelacionID  `json:"CalendarioEventoExtensionId"`
	ExtensionPadreId            *RelacionID `json:"ExtensionPadreId,omitempty"`
	DependenciaId               int         `json:"DependenciaId"`
	Vigente                     bool        `json:"Vigente"`
	Activo                      bool        `json:"Activo"`
	FechaCreacion               time.Time   `json:"FechaCreacion,omitempty"`
}

type RangoFechas struct {
	FechaInicio time.Time `json:"FechaInicio"`
	FechaFin    time.Time `json:"FechaFin"`
}

type RangoActividadDependencia struct {
	CalendarioEventoId string      `json:"CalendarioEventoId"`
	DependenciaId      int         `json:"DependenciaId"`
	RangoOriginal      RangoFechas `json:"RangoOriginal"`
	RangoPermitido     RangoFechas `json:"RangoPermitido"`
	ExtensionVigenteId interface{} `json:"ExtensionVigenteId"`
}
