package models

type PerfilConfiguracion struct {
	Id                int    `json:"Id"`
	Nombre            string `json:"Nombre"`
	CodigoAbreviacion string `json:"CodigoAbreviacion,omitempty"`
	Activo            bool   `json:"Activo"`
}

type AplicacionConfiguracion struct {
	Id     int    `json:"Id"`
	Nombre string `json:"Nombre,omitempty"`
	Alias  string `json:"Alias,omitempty"`
	Activo bool   `json:"Activo,omitempty"`
}
