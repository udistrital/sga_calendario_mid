package controllers

import (
	"github.com/astaxie/beego"
	"github.com/udistrital/sga_calendario_mid/services"
	"github.com/udistrital/utils_oas/errorhandler"
	"github.com/udistrital/utils_oas/requestresponse"
)

// ConsultaCalendarioAcademicoController operations for Consulta_calendario_academico
type ConsultaCalendarioAcademicoController struct {
	beego.Controller
}

// URLMapping ...
func (c *ConsultaCalendarioAcademicoController) URLMapping() {
	c.Mapping("GetAll", c.GetAll)
	c.Mapping("GetOnePorId", c.GetOnePorId)
	c.Mapping("Put", c.PutInhabilitarCalendario)
	c.Mapping("PostCalendarioHijo", c.PostCalendarioHijo)
	c.Mapping("GetCalendarInfo", c.GetCalendarInfo)
	c.Mapping("PutCalendarioEstado", c.PutCalendarioEstado)
	c.Mapping("PutCalendarioDependencias", c.PutCalendarioDependencias)
	c.Mapping("PostProcesoCalendario", c.PostProcesoCalendario)
	c.Mapping("PutProcesoPeriodicidad", c.PutProcesoPeriodicidad)
	c.Mapping("PutProcesoEstado", c.PutProcesoEstado)
	c.Mapping("PutActividadDependencias", c.PutActividadDependencias)
	c.Mapping("PostValidarActividadesProgramasMasivo", c.PostValidarActividadesProgramasMasivo)
	c.Mapping("PostActividadesProgramasMasivo", c.PostActividadesProgramasMasivo)
	c.Mapping("GetFacultadesSecretario", c.GetFacultadesSecretario)
	c.Mapping("GetFacultadesDecano", c.GetFacultadesDecano)
	c.Mapping("GetEventosCrud", c.GetEventosCrud)
	c.Mapping("GetEventosCrudId", c.GetEventosCrudId)
	c.Mapping("PostEventosCrud", c.PostEventosCrud)
	c.Mapping("PutEventosCrud", c.PutEventosCrud)
	c.Mapping("DeleteEventosCrud", c.DeleteEventosCrud)
}

// GetAll ...
// @Title GetAll
// @Description Consulta todos los calendarios académicos junto con su periodo y metadatos asociados.
// @Param	query	query	string	false	"Filter. e.g. col1:v1,col2:v2 ..."
// @Param	fields	query	string	false	"Fields returned. e.g. col1,col2 ..."
// @Param	sortby	query	string	false	"Sorted-by fields. e.g. col1,col2 ..."
// @Param	order	query	string	false	"Order corresponding to each sortby field, if single value, apply to all sortby fields. e.g. desc,asc ..."
// @Param	limit	query	string	false	"Limit the size of result set. Must be an integer"
// @Param	offset	query	string	false	"Start position of result set. Must be an integer"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router / [get]
func (c *ConsultaCalendarioAcademicoController) GetAll() {
	defer errorhandler.HandlePanic(&c.Controller)

	resultado, err := services.GetAll()

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()
}

// PutCalendarioEstado ...
// @Title PutCalendarioEstado
// @Description Activa o inactiva un calendario académico. Si no se envía Activo, alterna el estado actual; al inactivar también inactiva procesos y actividades asociadas.
// @Param	id		path 	string	true	"Id del calendario"
// @Param	body	body	models.EstadoActivoRequest	true	"Estado objetivo del calendario; Activo es opcional"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /calendario/:id/estado [put]
func (c *ConsultaCalendarioAcademicoController) PutCalendarioEstado() {
	defer errorhandler.HandlePanic(&c.Controller)
	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	resultado, err := services.PutCalendarioEstado(c.Ctx.Input.Param(":id"), c.Ctx.Input.RequestBody, usuario)
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// PutCalendarioDependencias ...
// @Title PutCalendarioDependencias
// @Description Actualiza los programas académicos asociados a un calendario. Bloquea retiros con fechas particulares modificadas o extensiones vigentes.
// @Param	id		path 	string	true	"Id del calendario"
// @Param	body	body	models.CalendarioDependenciasRequest	true	"DependenciaId contiene JSON serializado con proyectos y fechas del calendario"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @Failure 409 operación bloqueada por impactos de desasociación
// @router /calendario/:id/dependencias [put]
func (c *ConsultaCalendarioAcademicoController) PutCalendarioDependencias() {
	defer errorhandler.HandlePanic(&c.Controller)
	authHeader := c.Ctx.Input.Header("Authorization")
	usuario := services.ExtraerUsuario(authHeader)
	resultado, err := services.PutCalendarioDependencias(c.Ctx.Input.Param(":id"), c.Ctx.Input.RequestBody, usuario, authHeader)
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else if impacto, ok := err.(*services.ImpactoDesasociacionCalendarioError); ok {
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 409, impacto.Data(), impacto.Error())
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// PostProcesoCalendario ...
// @Title PostProcesoCalendario
// @Description Crea un proceso asociado a un calendario académico, validando fechas y registrando auditoría.
// @Param	body	body	models.GenericPayload	true	"Datos del proceso para eventos_crud/proceso"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /proceso [post]
func (c *ConsultaCalendarioAcademicoController) PostProcesoCalendario() {
	defer errorhandler.HandlePanic(&c.Controller)
	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	resultado, err := services.PostProcesoCalendario(c.Ctx.Input.RequestBody, usuario)
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// PutProcesoPeriodicidad ...
// @Title PutProcesoPeriodicidad
// @Description Actualiza la periodicidad (TipoRecurrenciaId) de un proceso y registra auditoría.
// @Param	id		path 	string	true	"Id del proceso"
// @Param	body	body	models.ProcesoPeriodicidadRequest	true	"Periodicidad objetivo con TipoRecurrenciaId.Id"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /proceso/:id/periodicidad [put]
func (c *ConsultaCalendarioAcademicoController) PutProcesoPeriodicidad() {
	defer errorhandler.HandlePanic(&c.Controller)
	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	resultado, err := services.PutProcesoPeriodicidad(c.Ctx.Input.Param(":id"), c.Ctx.Input.RequestBody, usuario)
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// PutProcesoEstado ...
// @Title PutProcesoEstado
// @Description Activa o inactiva un proceso. Si no se envía Activo, alterna el estado actual; al inactivar también inactiva actividades asociadas.
// @Param	id		path 	string	true	"Id del proceso"
// @Param	body	body	models.EstadoActivoRequest	true	"Estado objetivo del proceso; Activo es opcional"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /proceso/:id/estado [put]
func (c *ConsultaCalendarioAcademicoController) PutProcesoEstado() {
	defer errorhandler.HandlePanic(&c.Controller)
	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	resultado, err := services.PutProcesoEstado(c.Ctx.Input.Param(":id"), c.Ctx.Input.RequestBody, usuario)
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// PutActividadDependencias ...
// @Title PutActividadDependencias
// @Description Actualiza los programas académicos asociados a una actividad (calendario_evento), validando fechas particulares y extensiones vigentes.
// @Param	id		path 	string	true	"Id de la actividad"
// @Param	body	body	models.ActividadDependenciasRequest	true	"DependenciaId contiene JSON serializado con proyectos y fechas particulares de la actividad"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @Failure 409 operación bloqueada por impactos de desasociación
// @router /actividad/:id/dependencias [put]
func (c *ConsultaCalendarioAcademicoController) PutActividadDependencias() {
	defer errorhandler.HandlePanic(&c.Controller)
	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	resultado, err := services.PutActividadDependencias(c.Ctx.Input.Param(":id"), c.Ctx.Input.RequestBody, usuario)
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// PostValidarActividadesProgramasMasivo ...
// @Title PostValidarActividadesProgramasMasivo
// @Description Prevalida asociaciones o desasociaciones masivas entre programas académicos y actividades del calendario sin persistir cambios.
// @Param	id		path 	string	true	"Id del calendario"
// @Param	body	body	models.ActividadesProgramasMasivoRequest	true	"Programas y actividades a validar. Operacion puede ser asociar o desasociar."
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /calendario/:id/actividades-programas/masivo/validar [post]
func (c *ConsultaCalendarioAcademicoController) PostValidarActividadesProgramasMasivo() {
	defer errorhandler.HandlePanic(&c.Controller)
	resultado, err := services.PostValidarActividadesProgramasMasivo(c.Ctx.Input.Param(":id"), c.Ctx.Input.RequestBody)
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// PostActividadesProgramasMasivo ...
// @Title PostActividadesProgramasMasivo
// @Description Asocia o desasocia programas académicos a múltiples actividades del calendario. La operación es atómica: si hay impactos bloqueantes no actualiza ninguna actividad.
// @Param	id		path 	string	true	"Id del calendario"
// @Param	body	body	models.ActividadesProgramasMasivoRequest	true	"Operacion: asociar o desasociar; ProgramaIds y ActividadIds son arreglos de enteros."
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @Failure 409 operación bloqueada por impactos
// @router /calendario/:id/actividades-programas/masivo [post]
func (c *ConsultaCalendarioAcademicoController) PostActividadesProgramasMasivo() {
	defer errorhandler.HandlePanic(&c.Controller)
	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	resultado, err := services.PostActividadesProgramasMasivo(c.Ctx.Input.Param(":id"), c.Ctx.Input.RequestBody, usuario)
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else if impacto, ok := err.(*services.ImpactoDesasociacionCalendarioError); ok {
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 409, impacto.Data(), impacto.Error())
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// GetFacultadesSecretario ...
// @Title GetFacultadesSecretario
// @Description Consulta facultades Oikos asociadas al secretario académico usando WSO2 académico y homologación
// @Param documento path string true "Documento del secretario académico"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /secretario-academico/:documento/facultades [get]
func (c *ConsultaCalendarioAcademicoController) GetFacultadesSecretario() {
	defer errorhandler.HandlePanic(&c.Controller)
	resultado, err := services.FacultadesOikosSecretario(c.Ctx.Input.Param(":documento"))
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// GetFacultadesDecano ...
// @Title GetFacultadesDecano
// @Description Consulta facultades Oikos asociadas al decano usando WSO2 académico y homologación
// @Param documento path string true "Documento del decano"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /decano/:documento/facultades [get]
func (c *ConsultaCalendarioAcademicoController) GetFacultadesDecano() {
	defer errorhandler.HandlePanic(&c.Controller)
	resultado, err := services.FacultadesOikosDecano(c.Ctx.Input.Param(":documento"))
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// GetEventosCrud ...
// @Title GetEventosCrud
// @Description Proxy de consulta GET a recursos del CRUD de eventos (calendario, proceso, calendario_evento, catálogos y relaciones autorizadas).
// @Param	recurso		path 	string	true	"Nombre del recurso (calendario, proceso, calendario_evento, evento_catalogo, tipo_publico, etc.)"
// @Param	query		query	string	false	"Filtro en formato eventos_crud. e.g. col1:v1,col2:v2 ..."
// @Param	fields		query	string	false	"Campos retornados. e.g. col1,col2 ..."
// @Param	sortby		query	string	false	"Campos de ordenamiento. e.g. col1,col2 ..."
// @Param	order		query	string	false	"Orden por campo. e.g. desc,asc ..."
// @Param	limit		query	string	false	"Limit the size of result set. Must be an integer"
// @Param	offset		query	string	false	"Start position of result set. Must be an integer"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /eventos/:recurso [get]
func (c *ConsultaCalendarioAcademicoController) GetEventosCrud() {
	defer errorhandler.HandlePanic(&c.Controller)
	resultado, err := services.GetEventosCrud(c.Ctx.Input.Param(":recurso"), "", c.Ctx.Request.URL.RawQuery)
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// GetEventosCrudId ...
// @Title GetEventosCrudId
// @Description Proxy de consulta GET por ID a recursos del CRUD de eventos.
// @Param	recurso		path 	string	true	"Nombre del recurso"
// @Param	id			path 	string	true	"Id del recurso"
// @Param	query		query	string	false	"Parámetros adicionales enviados al CRUD de eventos"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /eventos/:recurso/:id [get]
func (c *ConsultaCalendarioAcademicoController) GetEventosCrudId() {
	defer errorhandler.HandlePanic(&c.Controller)
	resultado, err := services.GetEventosCrud(c.Ctx.Input.Param(":recurso"), c.Ctx.Input.Param(":id"), c.Ctx.Request.URL.RawQuery)
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// PostEventosCrud ...
// @Title PostEventosCrud
// @Description Proxy de creación POST a recursos permitidos del CRUD de eventos, con registro de auditoría.
// @Param	recurso		path 	string	true	"Nombre del recurso"
// @Param	body	body	models.GenericPayload	true	"Datos del recurso a crear en eventos_crud"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /eventos/:recurso [post]
func (c *ConsultaCalendarioAcademicoController) PostEventosCrud() {
	defer errorhandler.HandlePanic(&c.Controller)
	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	resultado, err := services.PostEventosCrud(c.Ctx.Input.Param(":recurso"), c.Ctx.Input.RequestBody, usuario)
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// PutEventosCrud ...
// @Title PutEventosCrud
// @Description Proxy de actualización PUT a recursos permitidos del CRUD de eventos, con registro de auditoría de la entidad actualizada.
// @Param	recurso		path 	string	true	"Nombre del recurso"
// @Param	id			path 	string	true	"Id del recurso"
// @Param	body	body	models.GenericPayload	true	"Datos actualizados del recurso en eventos_crud"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /eventos/:recurso/:id [put]
func (c *ConsultaCalendarioAcademicoController) PutEventosCrud() {
	defer errorhandler.HandlePanic(&c.Controller)
	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	resultado, err := services.PutEventosCrud(c.Ctx.Input.Param(":recurso"), c.Ctx.Input.Param(":id"), c.Ctx.Input.RequestBody, usuario)
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// DeleteEventosCrud ...
// @Title DeleteEventosCrud
// @Description Proxy de eliminación DELETE a recursos permitidos del CRUD de eventos, con registro de auditoría.
// @Param	recurso		path 	string	true	"Nombre del recurso"
// @Param	id			path 	string	true	"Id del recurso"
// @Param	body	body	models.GenericPayload	true	"Payload con TerceroId para registrar la operación"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /eventos/:recurso/:id [delete]
func (c *ConsultaCalendarioAcademicoController) DeleteEventosCrud() {
	defer errorhandler.HandlePanic(&c.Controller)
	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	resultado, err := services.DeleteEventosCrud(c.Ctx.Input.Param(":recurso"), c.Ctx.Input.Param(":id"), c.Ctx.Input.RequestBody, usuario)
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// GetOnePorId ...
// @Title GetOnePorId
// @Description Consulta un calendario académico por id, incluyendo procesos, actividades y metadatos relacionados.
// @Param	id		path 	string	true		"Id del calendario académico"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /:id [get]
func (c *ConsultaCalendarioAcademicoController) GetOnePorId() {
	defer errorhandler.HandlePanic(&c.Controller)

	idCalendario := c.Ctx.Input.Param(":id")

	resultado, err := services.GetOnePorId(idCalendario)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()
}

// PutInhabilitarCalendario ...
// @Title PutInhabilitarCalendario
// @Description Inhabilita un calendario académico y sus procesos y actividades asociadas.
// @Param	id		path 	string	true		"Id del calendario académico a inhabilitar"
// @Param	body	body	models.GenericPayload	true	"Payload recibido desde cliente; se usa para validar solicitud y registrar operación"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /calendario/academico/:id/inhabilitar [put]
func (c *ConsultaCalendarioAcademicoController) PutInhabilitarCalendario() {
	defer errorhandler.HandlePanic(&c.Controller)

	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	idCalendario := c.Ctx.Input.Param(":id")
	data := c.Ctx.Input.RequestBody

	resultado, err := services.PutInhabilitarCalendario(idCalendario, data, usuario)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()
}

// PostCalendarioHijo ...
// @Title PostCalendarioHijo
// @Description Crea un calendario académico independiente/hijo para periodo y nivel, validando que no exista otro activo equivalente.
// @Param	body	body	models.GenericPayload	true	"Datos del calendario: Nombre, DocumentoId, PeriodoId, Nivel y Activo"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /padre [post]
func (c *ConsultaCalendarioAcademicoController) PostCalendarioHijo() {
	defer errorhandler.HandlePanic(&c.Controller)

	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	data := c.Ctx.Input.RequestBody

	resultado, err := services.PostCalendarioHijo(data, usuario)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()
}

// GetCalendarInfo ...
// @Title GetCalendarInfo
// @Description Consulta información resumida de un calendario académico por id.
// @Param	id		path 	string	true		"Id de calendario"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /v2/:id [get]
func (c *ConsultaCalendarioAcademicoController) GetCalendarInfo() {
	defer errorhandler.HandlePanic(&c.Controller)

	idCalendario := c.Ctx.Input.Param(":id")

	resultado, err := services.GetCalendarInfo(idCalendario)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()

}
