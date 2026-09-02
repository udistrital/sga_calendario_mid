package controllers

import (
	"github.com/astaxie/beego"
	"github.com/udistrital/sga_calendario_mid/services"

	"github.com/udistrital/utils_oas/errorhandler"
	"github.com/udistrital/utils_oas/requestresponse"
)

type EventoController struct {
	beego.Controller
}

// URLMapping ...
func (c *EventoController) URLMapping() {
	c.Mapping("PostEvento", c.PostEvento)
	c.Mapping("PutEvento", c.PutEvento)
	c.Mapping("GetEvento", c.GetEvento)
	c.Mapping("DeleteEvento", c.DeleteEvento)
}

// PostEvento ...
// @Title PostEvento
// @Description Crea un evento compuesto usando tr_evento del CRUD de eventos.
// @Param	body	body	models.GenericPayload	true	"Payload con Evento, EncargadosEvento y TiposPublico"
// @Success 200 {}
// @Failure 404 recurso no encontrado o solicitud inválida
// @router / [post]
func (c *EventoController) PostEvento() {
	defer errorhandler.HandlePanic(&c.Controller)

	data := c.Ctx.Input.RequestBody

	resultado, err := services.PostEvento(data)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()
}

// PutEvento ...
// @Title PutEvento
// @Description Actualiza un evento compuesto usando tr_evento del CRUD de eventos.
// @Param	id		path 	string	true		"Id del evento a modificar"
// @Param	body	body	models.GenericPayload	true	"Payload con Evento, EncargadosEvento, TiposPublico y elementos borrados cuando aplique"
// @Success 200 {}
// @Failure 404 recurso no encontrado o solicitud inválida
// @router /:id [put]
func (c *EventoController) PutEvento() {
	defer errorhandler.HandlePanic(&c.Controller)

	idStr := c.Ctx.Input.Param(":id")
	data := c.Ctx.Input.RequestBody

	resultado, err := services.PutEvento(idStr, data)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()
}

// GetEvento ...
// @Title GetEvento
// @Description Consulta eventos asociados a una persona.
// @Param	persona	path	string	true	"Identificador de la persona"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /evento/persona/:persona [get]
func (c *EventoController) GetEvento() {
	defer errorhandler.HandlePanic(&c.Controller)

	persona := c.Ctx.Input.Param(":persona")

	resultado, err := services.GetEvento(persona)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()
}

// DeleteEvento ...
// @Title DeleteEvento
// @Description Elimina o inactiva un evento compuesto por id usando el servicio de eventos.
// @Param	id	path	string	true	"Id del evento"
// @Success 200 {}
// @Failure 404 recurso no encontrado
// @router /:id [delete]
func (c *EventoController) DeleteEvento() {
	defer errorhandler.HandlePanic(&c.Controller)

	id := c.Ctx.Input.Param(":id")

	resultado, err := services.DeleteEvento(id)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()
}
