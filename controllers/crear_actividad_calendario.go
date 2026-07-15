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
// @Description Agregar actividad calendario, tipo_publico y tabla de rompimiento calendario_evento_tipo_publico
// @Param	body		body 	{}	true		"body Agregar Actividad calendario content"
// @Success 200 {}
// @Failure 403 body is empty
// @router / [post]
func (c *ActividadCalendarioController) PostActividadCalendario() {
	defer errorhandler.HandlePanic(&c.Controller)

	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	data := c.Ctx.Input.RequestBody

	resultado, err := services.PostActividadCalendario(data, usuario)

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
// @Description Actualiza tabla de rompimiento calendario_evento_tipo_publico segun los responsables de una Actividad
// @Param	body		body 	{}	true		"body Actualizar responsables de una Actividad content"
// @Success 200 {}
// @Failure 403 body is empty
// @router /calendario/actividad/:id [put]
func (c *ActividadCalendarioController) UpdateActividadResponsables() {
	defer errorhandler.HandlePanic(&c.Controller)

	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	idStr := c.Ctx.Input.Param(":id")
	data := c.Ctx.Input.RequestBody

	resultado, err := services.UpdateActividadResponsables(idStr, data, usuario)

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
// @Description Crea una extensión de fecha fin para una actividad y programas autorizados
// @Param id path string true "Id de la actividad"
// @Param body body {} true "body extensión"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /:id/extension [post]
func (c *ActividadCalendarioController) PostExtensionActividad() {
	defer errorhandler.HandlePanic(&c.Controller)
	authHeader := c.Ctx.Input.Header("Authorization")
	usuario := services.ExtraerUsuario(authHeader)
	resultado, err := services.PostExtensionActividad(c.Ctx.Input.Param(":id"), c.Ctx.Input.RequestBody, usuario, authHeader)
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
// @Description Actualiza datos editables de una extensión de actividad
// @Param id path string true "Id de la actividad"
// @Param extension path string true "Id de la extensión"
// @Param body body {} true "body extensión"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /:id/extension/:extension [put]
func (c *ActividadCalendarioController) PutExtensionActividad() {
	defer errorhandler.HandlePanic(&c.Controller)
	authHeader := c.Ctx.Input.Header("Authorization")
	usuario := services.ExtraerUsuario(authHeader)
	resultado, err := services.PutExtensionActividad(c.Ctx.Input.Param(":id"), c.Ctx.Input.Param(":extension"), c.Ctx.Input.RequestBody, usuario, authHeader)
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
// @Description Inactiva una extensión de actividad y sus relaciones usando PUT para evitar restricciones de gateway con DELETE
// @Param id path string true "Id de la actividad"
// @Param extension path string true "Id de la extensión"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /:id/extension/:extension/anular [put]
func (c *ActividadCalendarioController) AnularExtensionActividad() {
	defer errorhandler.HandlePanic(&c.Controller)
	authHeader := c.Ctx.Input.Header("Authorization")
	usuario := services.ExtraerUsuario(authHeader)
	resultado, err := services.DeleteExtensionActividad(c.Ctx.Input.Param(":id"), c.Ctx.Input.Param(":extension"), usuario, authHeader)
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
// @Description Inactiva una extensión de actividad y sus relaciones
// @Param id path string true "Id de la actividad"
// @Param extension path string true "Id de la extensión"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /:id/extension/:extension [delete]
func (c *ActividadCalendarioController) DeleteExtensionActividad() {
	defer errorhandler.HandlePanic(&c.Controller)
	authHeader := c.Ctx.Input.Header("Authorization")
	usuario := services.ExtraerUsuario(authHeader)
	resultado, err := services.DeleteExtensionActividad(c.Ctx.Input.Param(":id"), c.Ctx.Input.Param(":extension"), usuario, authHeader)
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
// @Description Consulta extensiones de una actividad
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
// @Description Consulta rango permitido para una actividad y dependencia
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
