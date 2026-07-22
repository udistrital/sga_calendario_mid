package models

type EstadoActivoRequest struct {
	Activo *bool `json:"Activo"`
}

type CalendarioDependenciasRequest struct {
	DependenciaId string `json:"DependenciaId"`
	Forzar        bool   `json:"Forzar"`
}

type ProcesoPeriodicidadRequest struct {
	TipoRecurrenciaId RelacionID `json:"TipoRecurrenciaId"`
}

type ActividadDependenciasRequest struct {
	DependenciaId string `json:"DependenciaId"`
}

type ActividadesProgramasMasivoRequest struct {
	ProgramaIds  []int  `json:"ProgramaIds"`
	ActividadIds []int  `json:"ActividadIds"`
	Operacion    string `json:"Operacion"`
}
