package helpers

import (
	"testing"
	"time"
)

func TestParseFechaPreservaHoraVisibleConZona(t *testing.T) {
	fecha, err := ParseFecha("2026-07-23T14:30:00Z")
	if err != nil {
		t.Fatalf("ParseFecha retorno error: %v", err)
	}

	if got := fecha.Format(formatoFechaHoraCalendario); got != "2026-07-23T14:30:00" {
		t.Fatalf("hora visible cambiada: got %s", got)
	}
	if fecha.Location().String() != GMTMinus5Location.String() {
		t.Fatalf("ubicacion inesperada: got %s", fecha.Location())
	}
}

func TestFechaTimeParaCRUDPreservaHoraVisible(t *testing.T) {
	casos := []struct {
		nombre string
		fecha  interface{}
		want   string
	}{
		{"string con Z", "2026-07-23T14:30:00Z", "2026-07-23T14:30:00Z"},
		{"string con offset", "2026-07-23T14:30:00-05:00", "2026-07-23T14:30:00Z"},
		{"time UTC", time.Date(2026, 7, 23, 14, 30, 0, 0, time.UTC), "2026-07-23T14:30:00Z"},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if got := FechaTimeParaCRUD(caso.fecha); got != caso.want {
				t.Fatalf("FechaTimeParaCRUD() = %s, want %s", got, caso.want)
			}
		})
	}
}

func TestFechaTimeParaModeloPreservaHoraVisibleComoTransporteUTC(t *testing.T) {
	fecha, err := FechaTimeParaModelo("2026-07-23T14:30:00-05:00")
	if err != nil {
		t.Fatalf("FechaTimeParaModelo retorno error: %v", err)
	}

	if got := fecha.Format(time.RFC3339); got != "2026-07-23T14:30:00Z" {
		t.Fatalf("fecha de transporte inesperada: got %s", got)
	}
}

func TestFormatFechaGMTMinus5NoDesplazaHoraVisible(t *testing.T) {
	fecha := time.Date(2026, 7, 23, 14, 30, 0, 0, time.UTC)

	if got := FormatFechaGMTMinus5(fecha); got != "2026-07-23T14:30:00-05:00" {
		t.Fatalf("FormatFechaGMTMinus5() = %s", got)
	}
}

func TestValidarRangoFechasComparaHoraVisible(t *testing.T) {
	err := ValidarRangoFechas("2026-07-23T14:00:00-05:00", "2026-07-23T14:30:00Z")
	if err != nil {
		t.Fatalf("ValidarRangoFechas retorno error: %v", err)
	}
}
