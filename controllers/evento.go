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
// @Description Agregar Evento
// @Param   body        body    {}  true        "body Agregar Evento content"
// @Success 200 {}
// @Failure 403 body is empty
// @router / [post]
func (c *EventoController) PostEvento() {
	defer errorhandler.HandlePanic(&c.Controller)

	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	data := c.Ctx.Input.RequestBody

	resultado, err := services.PostEvento(data, usuario)

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
// @Description Modificar Evento
// @Param	id		path 	string	true		"el id del evento a modificar"
// @Param   body        body    {}  true        "body Modificar Evento content"
// @Success 200 {}
// @Failure 403 :id is empty
// @router /:id [put]
func (c *EventoController) PutEvento() {
	defer errorhandler.HandlePanic(&c.Controller)

	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	idStr := c.Ctx.Input.Param(":id")
	data := c.Ctx.Input.RequestBody

	resultado, err := services.PutEvento(idStr, data, usuario)

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
// @Description consultar Evento por persona
// @Param   persona      path    string  true        "Persona"
// @Success 200 {}
// @Failure 403 :persona is empty
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
// @Description eliminar Evento por id
// @Param   id      path    string  true        "Id del Evento"
// @Success 200 {}
// @Failure 403 :id is empty
// @router /:id [delete]
func (c *EventoController) DeleteEvento() {
	defer errorhandler.HandlePanic(&c.Controller)

	usuario := services.ExtraerUsuario(c.Ctx.Input.Header("Authorization"))
	id := c.Ctx.Input.Param(":id")

	resultado, err := services.DeleteEvento(id, usuario)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()
}
