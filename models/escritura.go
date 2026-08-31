package models

type EstadoActivoRequest struct {
	Activo    *bool `json:"Activo"`
	TerceroId int   `json:"TerceroId"`
}

type CalendarioDependenciasRequest struct {
	DependenciaId string `json:"DependenciaId"`
	Forzar        bool   `json:"Forzar"`
	TerceroId     int    `json:"TerceroId"`
}

type ProcesoPeriodicidadRequest struct {
	TipoRecurrenciaId RelacionID `json:"TipoRecurrenciaId"`
}

type ActividadDependenciasRequest struct {
	DependenciaId string `json:"DependenciaId"`
	TerceroId     int    `json:"TerceroId"`
}

type ActividadesProgramasMasivoRequest struct {
	ProgramaIds  []int  `json:"ProgramaIds"`
	ActividadIds []int  `json:"ActividadIds"`
	Operacion    string `json:"Operacion"`
	TerceroId    int    `json:"TerceroId"`
}
