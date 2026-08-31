package models

type DependenciaEvento struct {
	Proyectos []int                     `json:"proyectos"`
	Fechas    []FechaParticularPrograma `json:"fechas"`
}

type FechaParticularPrograma struct {
	Id           int    `json:"Id"`
	Inicio       string `json:"Inicio"`
	Fin          string `json:"Fin"`
	Modificacion string `json:"Modificacion"`
	Activo       *bool  `json:"Activo,omitempty"`
}

func (f FechaParticularPrograma) ActivoPtr() bool {
	return f.Activo == nil || *f.Activo
}
