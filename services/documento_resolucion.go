package services

import (
	"encoding/json"

	"github.com/astaxie/beego/logs"
	"github.com/udistrital/sga_calendario_mid/models"
)

func resolucionDesdeDocumento(documento map[string]interface{}) map[string]interface{} {
	if documento == nil {
		return nil
	}

	metadatoJSON, ok := documento["Metadatos"].(string)
	if !ok || metadatoJSON == "" {
		return nil
	}

	var metadato models.Metadatos
	if err := json.Unmarshal([]byte(metadatoJSON), &metadato); err != nil {
		logs.Warn("Metadatos de documento invalidos: %v", err)
		return nil
	}

	return map[string]interface{}{
		"Id":         documento["Id"],
		"Enlace":     documento["Enlace"],
		"Resolucion": metadato.Resolucion,
		"Anno":       metadato.Anno,
		"Nombre":     documento["Nombre"],
	}
}
