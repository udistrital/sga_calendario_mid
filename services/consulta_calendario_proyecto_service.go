package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
)

func GetCalendarByProjectId(idCalendario int) (interface{}, error) {
	var calendarios []map[string]interface{}
	var CalendarioId string = "0"
	var Calendario map[string]interface{}

	errCalendarios := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario?query=Activo:true&limit=0&sortby=Id&order=desc", &calendarios)
	if errCalendarios == nil && fmt.Sprintf("%v", calendarios[0]["Nombre"]) != "map[]" {
		for _, calendario := range calendarios {
			AplicaExtension := calendario["AplicaExtension"].(bool)
			if AplicaExtension {
				DependenciaParticularId := calendario["DependenciaParticularId"].(string)
				if DependenciaParticularId != "{}" && DependenciaParticularId != "" {
					var listaProyectos map[string][]int
					json.Unmarshal([]byte(DependenciaParticularId), &listaProyectos)
					for _, Id := range listaProyectos["proyectos"] {
						if Id == idCalendario {
							CalendarioId = strconv.FormatFloat(calendario["Id"].(float64), 'f', 0, 64)
							break
						}
					}
				}
			} else {
				DependenciaId := calendario["DependenciaId"].(string)
				if DependenciaId != "{}" {
					var listaProyectos map[string][]int
					json.Unmarshal([]byte(DependenciaId), &listaProyectos)
					for _, Id := range listaProyectos["proyectos"] {
						if Id == idCalendario {
							CalendarioId = strconv.FormatFloat(calendario["Id"].(float64), 'f', 0, 64)
							break
						}
					}
				}
			}
			if CalendarioId != "0" {
				break
			}
		}
		Calendario = map[string]interface{}{
			"CalendarioId": CalendarioId,
		}
		return requestresponse.APIResponseDTO(true, 200, Calendario), nil
	} else {
		logs.Error(errCalendarios.Error())
		return nil, errors.New("error del servicio GetCalendarByProjectId: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
	}
}

func GetCalendarProject(idNiv string, idPer string) (interface{}, error) {
	var calendarios []map[string]interface{}
	var calendarioEventos []map[string]interface{}
	var proyectos []map[string]interface{}
	var proyectosP []map[string]interface{}
	var proyectosH []map[string]interface{}
	var CalendarioId string = "0"
	var proyectosArrMap []map[string]interface{}

	// list proyectos padres
	errProyectosP := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+"proyecto_academico_institucion?query=Activo:true,NivelFormacionId.Id:"+fmt.Sprintf("%v", idNiv)+"&sortby=Nombre&order=asc&limit=0&fields=Id,Nombre", &proyectosP)
	if errProyectosP == nil {
		if fmt.Sprintf("%v", proyectosP) != "[map[]]" {
			proyectos = append(proyectos, proyectosP...)
		}
		// list proyectos hijos
		errProyectosH := request.GetJson("http://"+beego.AppConfig.String("ProyectoAcademicoService")+"proyecto_academico_institucion?query=Activo:true,NivelFormacionId.NivelFormacionPadreId.Id:"+fmt.Sprintf("%v", idNiv)+"&sortby=Nombre&order=asc&limit=0&fields=Id,Nombre", &proyectosH)
		if errProyectosH == nil {
			if fmt.Sprintf("%v", proyectosH) != "[map[]]" {
				proyectos = append(proyectos, proyectosH...)
			}

			if len(proyectos) > 0 {
				errCalendarios := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario?query=Activo:true,Nivel:"+fmt.Sprintf("%v", idNiv)+",PeriodoId:"+fmt.Sprintf("%v", idPer)+"&limit=0&sortby=Id&order=desc", &calendarios)
				if errCalendarios == nil && fmt.Sprintf("%v", calendarios) != "[map[]]" {

					for _, proyecto := range proyectos {
						IdPro := int(proyecto["Id"].(float64))
						CalendarioId = "0"
						for _, calendario := range calendarios {
							AplicaExtension := calendario["AplicaExtension"].(bool)
							if AplicaExtension {
								DependenciaParticularId := calendario["DependenciaParticularId"].(string)
								if DependenciaParticularId != "{}" && DependenciaParticularId != "" {
									var listaProyectos map[string][]int
									json.Unmarshal([]byte(DependenciaParticularId), &listaProyectos)
									for _, Id := range listaProyectos["proyectos"] {
										if Id == IdPro {
											CalendarioId = strconv.FormatFloat(calendario["Id"].(float64), 'f', 0, 64)
											break
										}
									}
								}
							} else {
								DependenciaId := calendario["DependenciaId"].(string)
								if DependenciaId != "{}" {
									var listaProyectos map[string][]int
									json.Unmarshal([]byte(DependenciaId), &listaProyectos)
									for _, Id := range listaProyectos["proyectos"] {
										if Id == IdPro {
											CalendarioId = strconv.FormatFloat(calendario["Id"].(float64), 'f', 0, 64)
											break
										}
									}
								}
							}
							if CalendarioId != "0" {
								proyectoInfo := map[string]interface{}{
									"ProyectoId":          IdPro,
									"NombreProyecto":      proyecto["Nombre"],
									"CalendarioID":        CalendarioId,
									"CalendarioExtension": AplicaExtension,
									"Evento":              nil,
									"EventoInscripcion":   nil,
								}
								proyectosArrMap = append(proyectosArrMap, proyectoInfo)
								break
							}
						}
					}

					if len(proyectosArrMap) > 0 {
						for i := range proyectosArrMap {
							errEvento := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario_evento/?query=TipoEventoId__CalendarioID__Id:"+proyectosArrMap[i]["CalendarioID"].(string)+",Activo:true&limit=0", &calendarioEventos)
							if errEvento == nil && fmt.Sprintf("%v", calendarioEventos) != "[map[]]" {

								var lista_eventos []map[string]interface{}
								for _, Evento := range calendarioEventos {
									nombreEvento := strings.ToUpper(Evento["Nombre"].(string))
									codAbrEvento := Evento["TipoEventoId"].(map[string]interface{})["CodigoAbreviacion"].(string)
									pago := strings.Contains(nombreEvento, "PAGO")
									var aplicaParticular bool = false
									if fmt.Sprintf("%v", Evento["DependenciaId"]) != "" && fmt.Sprintf("%v", Evento["DependenciaId"]) != "{}" {
										var listaProyectos map[string]interface{}
										json.Unmarshal([]byte(Evento["DependenciaId"].(string)), &listaProyectos)
										for _, project := range listaProyectos["fechas"].([]interface{}) {
											if int(project.(map[string]interface{})["Id"].(float64)) == proyectosArrMap[i]["ProyectoId"].(int) {
												if project.(map[string]interface{})["Activo"].(bool) {
													// datos_respuesta := map[string]interface{}{
													evento_x := map[string]interface{}{
														"ActividadParticular": true,
														"NombreEvento":        Evento["Descripcion"],
														"FechaInicioEvento":   project.(map[string]interface{})["Inicio"],
														"FechaFinEvento":      project.(map[string]interface{})["Fin"],
														"CodigoAbreviacion":   codAbrEvento,
														"Pago":                pago,
													}
													lista_eventos = append(lista_eventos, evento_x)
												}
												aplicaParticular = true
												break
											}
										}
									}
									if !aplicaParticular {
										evento_x := map[string]interface{}{
											"ActividadParticular": false,
											"NombreEvento":        Evento["Descripcion"],
											"FechaInicioEvento":   Evento["FechaInicio"],
											"FechaFinEvento":      Evento["FechaFin"],
											"CodigoAbreviacion":   codAbrEvento,
											"Pago":                pago,
										}
										lista_eventos = append(lista_eventos, evento_x)
									}
								}
								proyectosArrMap[i]["Evento"] = lista_eventos
							}
						}
					}
					return requestresponse.APIResponseDTO(true, 200, proyectosArrMap), nil

				} else {
					logs.Error(errCalendarios.Error())
					return nil, errors.New("error del servicio GetCalendarProject: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
				}

			} else {
				proyectos = []map[string]interface{}{}
				return requestresponse.APIResponseDTO(true, 200, proyectos), nil
			}

		} else {
			logs.Error(errProyectosH.Error())
			return nil, errors.New("error del servicio GetCalendarProject: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
		}
	} else {
		logs.Error(errProyectosP.Error())
		return nil, errors.New("error del servicio GetCalendarProject: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
	}

}
