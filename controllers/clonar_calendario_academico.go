package controllers

import (
	"github.com/astaxie/beego"
	"github.com/udistrital/sga_calendario_mid/services"
	"github.com/udistrital/utils_oas/requestresponse"

	"github.com/udistrital/utils_oas/errorhandler"
)

type ClonarCalendarioController struct {
	beego.Controller
}

func (c *ClonarCalendarioController) URLMapping() {
	c.Mapping("PostCalendario", c.PostCalendario)
	c.Mapping("PostCalendarioPadre", c.PostCalendarioPadre)
}

// PostCalendario ...
// @Title PostCalendario
// @Description Clona procesos, actividades y públicos dirigidos desde el calendario activo del periodo/nivel destino hacia el calendario indicado.
// @Param	body		body 	models.GenericPayload	true		"Datos de clonación: Id, PeriodoIdClone y NivelClone"
// @Success 200 {}
// @Failure 404 recurso no encontrado o solicitud inválida
// @router / [post]
func (c *ClonarCalendarioController) PostCalendario() {
	defer errorhandler.HandlePanic(&c.Controller)

	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	data := c.Ctx.Input.RequestBody

	resultado, err := services.PostCalendario(data, usuario)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()

}

// PostCalendarioPadre ...
// @Title PostCalendarioPadre
// @Description Clona procesos, actividades y públicos dirigidos desde un calendario padre hacia un calendario destino existente o creado por la solicitud.
// @Param	body		body 	models.GenericPayload	true		"Datos de clonación padre: IdPadre y datos del calendario destino"
// @Success 200 {}
// @Failure 404 recurso no encontrado o solicitud inválida
// @router /padre [post]
func (c *ClonarCalendarioController) PostCalendarioPadre() {
	defer errorhandler.HandlePanic(&c.Controller)

	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	data := c.Ctx.Input.RequestBody

	resultado, err := services.PostCalendarioPadre(data, usuario)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()
}
