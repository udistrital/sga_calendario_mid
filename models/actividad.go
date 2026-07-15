package models

type CalendarioEventoTipoPublicoPayload struct {
	Activo             bool       `json:"Activo"`
	PerfilId           int        `json:"PerfilId"`
	CalendarioEventoId RelacionID `json:"CalendarioEventoId"`
}
