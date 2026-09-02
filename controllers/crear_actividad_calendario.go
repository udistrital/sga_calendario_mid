package controllers

import (
	"github.com/astaxie/beego"
	"github.com/udistrital/sga_calendario_mid/services"
	"github.com/udistrital/utils_oas/errorhandler"
	"github.com/udistrital/utils_oas/requestresponse"
)

type ActividadCalendarioController struct {
	beego.Controller
}

func (c *ActividadCalendarioController) URLMapping() {
	c.Mapping("PostActividadCalendario", c.PostActividadCalendario)
	c.Mapping("UpdateActividadResponsables", c.UpdateActividadResponsables)
	c.Mapping("PostExtensionActividad", c.PostExtensionActividad)
	c.Mapping("PutExtensionActividad", c.PutExtensionActividad)
	c.Mapping("AnularExtensionActividad", c.AnularExtensionActividad)
	c.Mapping("DeleteExtensionActividad", c.DeleteExtensionActividad)
	c.Mapping("GetExtensionesActividad", c.GetExtensionesActividad)
	c.Mapping("GetRangoActividadDependencia", c.GetRangoActividadDependencia)
}

// PostActividadCalendario ...
// @Title PostActividadCalendario
// @Description Crea una actividad de calendario (calendario_evento) y sus relaciones de público dirigido (calendario_evento_tipo_publico).
// @Param	body		body 	models.GenericPayload	true		"Payload con Actividad y responsable[]"
// @Success 200 {}
// @Failure 404 recurso no encontrado o solicitud inválida
// @router / [post]
func (c *ActividadCalendarioController) PostActividadCalendario() {
	defer errorhandler.HandlePanic(&c.Controller)

	data := c.Ctx.Input.RequestBody

	resultado, err := services.PostActividadCalendario(data)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()
}

// UpdateActividadResponsables ...
// @Title UpdateActividadResponsables
// @Description Actualiza las relaciones de público dirigido (calendario_evento_tipo_publico) de una actividad.
// @Param	id		path 	string	true		"Id de la actividad"
// @Param	body		body 	models.GenericPayload	true		"Payload con responsable[] y responsables eliminados cuando aplique"
// @Success 200 {}
// @Failure 404 recurso no encontrado o solicitud inválida
// @router /calendario/actividad/:id [put]
func (c *ActividadCalendarioController) UpdateActividadResponsables() {
	defer errorhandler.HandlePanic(&c.Controller)

	idStr := c.Ctx.Input.Param(":id")
	data := c.Ctx.Input.RequestBody

	resultado, err := services.UpdateActividadResponsables(idStr, data)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()
}

// PostExtensionActividad ...
// @Title PostExtensionActividad
// @Description Crea una extensión de fecha fin para una actividad y los programas académicos autorizados.
// @Param id path string true "Id de la actividad"
// @Param body body models.SolicitudExtensionActividadRequest true "Solicitud de extensión: FechaFin, DocumentoId, Descripcion y Dependencias"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /:id/extension [post]
func (c *ActividadCalendarioController) PostExtensionActividad() {
	defer errorhandler.HandlePanic(&c.Controller)
	resultado, err := services.PostExtensionActividad(c.Ctx.Input.Param(":id"), c.Ctx.Input.RequestBody)
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// PutExtensionActividad ...
// @Title PutExtensionActividad
// @Description Actualiza datos editables de una extensión de actividad y sus programas académicos autorizados.
// @Param id path string true "Id de la actividad"
// @Param extension path string true "Id de la extensión"
// @Param body body models.SolicitudExtensionActividadRequest true "Solicitud de extensión: FechaFin, DocumentoId, Descripcion y Dependencias"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /:id/extension/:extension [put]
func (c *ActividadCalendarioController) PutExtensionActividad() {
	defer errorhandler.HandlePanic(&c.Controller)
	resultado, err := services.PutExtensionActividad(c.Ctx.Input.Param(":id"), c.Ctx.Input.Param(":extension"), c.Ctx.Input.RequestBody)
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// AnularExtensionActividad ...
// @Title AnularExtensionActividad
// @Description Inactiva una extensión de actividad y sus relaciones usando PUT para evitar restricciones de gateway con DELETE.
// @Param id path string true "Id de la actividad"
// @Param extension path string true "Id de la extensión"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /:id/extension/:extension/anular [put]
func (c *ActividadCalendarioController) AnularExtensionActividad() {
	defer errorhandler.HandlePanic(&c.Controller)
	resultado, err := services.DeleteExtensionActividad(c.Ctx.Input.Param(":id"), c.Ctx.Input.Param(":extension"))
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// DeleteExtensionActividad ...
// @Title DeleteExtensionActividad
// @Description Inactiva una extensión de actividad y sus relaciones.
// @Param id path string true "Id de la actividad"
// @Param extension path string true "Id de la extensión"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /:id/extension/:extension [delete]
func (c *ActividadCalendarioController) DeleteExtensionActividad() {
	defer errorhandler.HandlePanic(&c.Controller)
	resultado, err := services.DeleteExtensionActividad(c.Ctx.Input.Param(":id"), c.Ctx.Input.Param(":extension"))
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// GetExtensionesActividad ...
// @Title GetExtensionesActividad
// @Description Consulta extensiones activas e históricas de una actividad.
// @Param id path string true "Id de la actividad"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /:id/extension [get]
func (c *ActividadCalendarioController) GetExtensionesActividad() {
	defer errorhandler.HandlePanic(&c.Controller)
	resultado, err := services.GetExtensionesActividad(c.Ctx.Input.Param(":id"))
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}

// GetRangoActividadDependencia ...
// @Title GetRangoActividadDependencia
// @Description Consulta el rango de fechas permitido para una actividad y un programa académico, considerando extensiones vigentes.
// @Param id path string true "Id de la actividad"
// @Param dependencia path string true "Id de la dependencia"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /:id/rango-dependencia/:dependencia [get]
func (c *ActividadCalendarioController) GetRangoActividadDependencia() {
	defer errorhandler.HandlePanic(&c.Controller)
	resultado, err := services.GetRangoActividadDependencia(c.Ctx.Input.Param(":id"), c.Ctx.Input.Param(":dependencia"))
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}
	c.ServeJSON()
}
