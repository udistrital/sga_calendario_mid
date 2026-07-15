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
// @Description get todos los calendarios académicos junto a sus periodos correspondientes
// @Param	query	query	string	false	"Filter. e.g. col1:v1,col2:v2 ..."
// @Param	fields	query	string	false	"Fields returned. e.g. col1,col2 ..."
// @Param	sortby	query	string	false	"Sorted-by fields. e.g. col1,col2 ..."
// @Param	order	query	string	false	"Order corresponding to each sortby field, if single value, apply to all sortby fields. e.g. desc,asc ..."
// @Param	limit	query	string	false	"Limit the size of result set. Must be an integer"
// @Param	offset	query	string	false	"Start position of result set. Must be an integer"
// @Success 200 {object} models.ConsultaCalendarioAcademico
// @Failure 404
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
// @Description Cambiar estado (Activo/Inactivo) de un calendario académico
// @Param	id		path 	string	true	"Id del calendario"
// @Param   body    body    {}      true	"body { Activo: bool }"
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
// @Description Actualizar DependenciaId (proyectos asociados) de un calendario
// @Param	id		path 	string	true	"Id del calendario"
// @Param   body    body    {}      true	"body { DependenciaId: string }"
// @Success 200 {}
// @Failure 404 recurso no encontrado
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
// @Description Crear un nuevo proceso asociado a un calendario académico
// @Param   body    body    {}  true	"body datos del proceso"
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
// @Description Actualizar el TipoRecurrenciaId de un proceso
// @Param	id		path 	string	true	"Id del proceso"
// @Param   body    body    {}      true	"body { TipoRecurrenciaId: { Id: int } }"
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
// @Description Cambiar estado (Activo/Inactivo) de un proceso
// @Param	id		path 	string	true	"Id del proceso"
// @Param   body    body    {}      true	"body { Activo: bool }"
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
// @Description Actualizar DependenciaId de una actividad (calendario_evento)
// @Param	id		path 	string	true	"Id de la actividad"
// @Param   body    body    {}      true	"body { DependenciaId: string }"
// @Success 200 {}
// @Failure 404 recurso no encontrado
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
// @Description Proxy de consulta GET a recursos del CRUD de eventos (calendario, proceso, calendario_evento, catálogos)
// @Param	recurso		path 	string	true	"Nombre del recurso (calendario, proceso, calendario_evento, evento_catalogo, tipo_publico, etc.)"
// @Param	query		query	string	false	"Filter. e.g. col1:v1,col2:v2 ..."
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
// @Description Proxy de consulta GET por ID a recursos del CRUD de eventos
// @Param	recurso		path 	string	true	"Nombre del recurso"
// @Param	id			path 	string	true	"Id del recurso"
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
// @Description Proxy de creación POST a recursos del CRUD de eventos
// @Param	recurso		path 	string	true	"Nombre del recurso"
// @Param   body    	body    {}    	true	"body datos del recurso a crear"
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
// @Description Proxy de actualización PUT a recursos del CRUD de eventos
// @Param	recurso		path 	string	true	"Nombre del recurso"
// @Param	id			path 	string	true	"Id del recurso"
// @Param   body    	body    {}    	true	"body datos actualizados del recurso"
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
// @Description Proxy de eliminación DELETE a recursos del CRUD de eventos
// @Param	recurso		path 	string	true	"Nombre del recurso"
// @Param	id			path 	string	true	"Id del recurso"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /eventos/:recurso/:id [delete]
func (c *ConsultaCalendarioAcademicoController) DeleteEventosCrud() {
	defer errorhandler.HandlePanic(&c.Controller)
	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	resultado, err := services.DeleteEventosCrud(c.Ctx.Input.Param(":recurso"), c.Ctx.Input.Param(":id"), usuario)
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
// @Description get obtener calendario académico por id
// @Param	id		path 	string	true		"The key for staticblock"
// @Success 200 {}
// @Failure 403 :id is empty
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
// @Description Inhabilitar Calendario
// @Param	id		path 	string	true		"el id del calendario a inhabilitar"
// @Param   body        body    {}  true        "body Inhabilitar calendario content"
// @Success 200 {}
// @Failure 403 :id is empty
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
// @Description  Proyecto obtener el Id de calendario padre, crear el nuevo calendario (hijo) e inactivar el calendario padre
// @Param   body        body    {}  true        "body crear calendario hijo content"
// @Success 200 {}
// @Failure 403 :body is empty
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
// @Description get obtener información calendario académico por id
// @Param	id		path 	string	true		"Id de calendario"
// @Success 200 {}
// @Failure 404 not found resource
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
