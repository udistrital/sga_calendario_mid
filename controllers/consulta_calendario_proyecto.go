package controllers

import (
	"strconv"

	"github.com/astaxie/beego"
	"github.com/udistrital/sga_calendario_mid/services"
	"github.com/udistrital/utils_oas/errorhandler"
	"github.com/udistrital/utils_oas/requestresponse"
)

type ConsultaCalendarioProyectoController struct {
	beego.Controller
}

// URLMapping
func (c *ConsultaCalendarioProyectoController) URLMapping() {
	c.Mapping("GetCalendarByProjectId", c.GetCalendarByProjectId)
	c.Mapping("GetCalendarProject", c.GetCalendarProject)
}

// GetCalendarByProjectId ...
// @Title GetCalendarByProjectId
// @Description get ConsultaCalendarioAcademico by id
// @Param	id		path 	string	true		"The key for staticblock"
// @Success 200
// @Failure 403 :id is empty
// @router /:id [get]
func (c *ConsultaCalendarioProyectoController) GetCalendarByProjectId() {
	defer errorhandler.HandlePanic(&c.Controller)

	idCalendario, _ := strconv.Atoi(c.Ctx.Input.Param(":id"))

	resultado, err := services.GetCalendarByProjectId(idCalendario)

	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(true, 404, nil, err.Error())
	}

	c.ServeJSON()
}

// GetCalendarProject ...
// @Title GetCalendarProject
// @Description get ConsultaCalendarioAcademico & id y Project By Id
// @Param	id-nivel	query	string	false	"Se recivbe parametro Id de el nivel"
// @Param	id-periodo	query	string	false	"Se recivbe parametro Id de el Periodo"
// @Success 200
// @Failure 403 :id is empty
// @router /calendario/proyecto [get]
func (c *ConsultaCalendarioProyectoController) GetCalendarProject() {
	defer errorhandler.HandlePanic(&c.Controller)
	var idPeriodo string
	var idNivel string

	// Id de el periodo
	if v := c.GetString("id-periodo"); v != "" {
		idPeriodo = v
	} else {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "Parametro id periodo vacío")
	}
	// Id de el nivel
	if v := c.GetString("id-nivel"); v != "" {
		idNivel = v
	} else {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 400, nil, "Parametro id nivel vacío")
	}

	resultado, err := services.GetCalendarProject(idNivel, idPeriodo)
	if err == nil {
		c.Ctx.Output.SetStatus(200)
		c.Data["json"] = resultado
	} else {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = requestresponse.APIResponseDTO(false, 404, nil, err.Error())
	}

	c.ServeJSON()
}
