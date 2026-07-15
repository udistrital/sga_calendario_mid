package models

type AuditoriaEventoPayload struct {
	EntidadTipo   string      `json:"EntidadTipo"`
	Operacion     string      `json:"Operacion"`
	ValorAnterior interface{} `json:"ValorAnterior"`
	ValorNuevo    interface{} `json:"ValorNuevo"`
	TerceroId     interface{} `json:"TerceroId"`
	Endpoint      string      `json:"Endpoint"`
	Activo        bool        `json:"Activo"`
}
