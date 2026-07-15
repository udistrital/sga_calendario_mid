package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/sga_calendario_mid/helpers"
	"github.com/udistrital/sga_calendario_mid/models"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
)

func parseDependenciaEvento(dependencia interface{}) (map[string]interface{}, bool) {
	dependenciaModel, ok := helpers.ParseDependenciaEvento(dependencia)
	if !ok {
		return nil, false
	}
	var dependenciaMap map[string]interface{}
	data, _ := json.Marshal(dependenciaModel)
	if err := json.Unmarshal(data, &dependenciaMap); err != nil {
		logs.Error("error parseando DependenciaId de evento: ", err)
		return nil, false
	}
	return dependenciaMap, true
}

func interfaceToInt(value interface{}) (int, bool) {
	return helpers.InterfaceToInt(value)
}

func dependenciaIncluyeProyecto(dependenciaMap map[string]interface{}, proyectoID int) bool {
	dependenciaModel, ok := dependenciaMapToModel(dependenciaMap)
	if !ok {
		return false
	}
	return helpers.DependenciaIncluyeProyecto(dependenciaModel, proyectoID)
}

func fechaParticularProyecto(dependenciaMap map[string]interface{}, proyectoID int) (map[string]interface{}, bool) {
	dependenciaModel, ok := dependenciaMapToModel(dependenciaMap)
	if !ok {
		return nil, false
	}
	fechaModel, ok := helpers.FechaParticularProyecto(dependenciaModel, proyectoID)
	if !ok {
		return nil, false
	}
	var fechaMap map[string]interface{}
	data, _ := json.Marshal(fechaModel)
	if err := json.Unmarshal(data, &fechaMap); err != nil {
		return nil, false
	}
	return fechaMap, true
}

func dependenciaMapToModel(dependenciaMap map[string]interface{}) (models.DependenciaEvento, bool) {
	if dependenciaMap == nil {
		return models.DependenciaEvento{}, false
	}
	data, _ := json.Marshal(dependenciaMap)
	var dependenciaModel models.DependenciaEvento
	if err := json.Unmarshal(data, &dependenciaModel); err != nil {
		return models.DependenciaEvento{}, false
	}
	return dependenciaModel, true
}

func GetCalendarByProjectId(idCalendario int, idPeriodo string) (interface{}, error) {
	var calendarios []map[string]interface{}
	var CalendarioId string = "0"
	var Calendario map[string]interface{}

	query := "Activo:true"
	if strings.TrimSpace(idPeriodo) != "" {
		query += ",PeriodoId:" + strings.TrimSpace(idPeriodo)
	}
	errCalendarios := request.GetJson(beego.AppConfig.String("EventoService")+"calendario?query="+query+"&limit=0&sortby=Id&order=desc", &calendarios)
	if errCalendarios == nil {
		for _, calendario := range calendarios {
			dependencia, ok := parseDependenciaEvento(calendario["DependenciaId"])
			if ok && dependenciaIncluyeProyecto(dependencia, idCalendario) {
				CalendarioId = fmt.Sprintf("%v", calendario["Id"])
				if id, ok := interfaceToInt(calendario["Id"]); ok {
					CalendarioId = strconv.Itoa(id)
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
		if errCalendarios != nil {
			logs.Error(errCalendarios.Error())
		}
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
	errProyectosP := request.GetJson(beego.AppConfig.String("ProyectoAcademicoService")+"proyecto_academico_institucion?query=Activo:true,NivelFormacionId.Id:"+fmt.Sprintf("%v", idNiv)+"&sortby=Nombre&order=asc&limit=0&fields=Id,Nombre", &proyectosP)
	if errProyectosP == nil {
		if fmt.Sprintf("%v", proyectosP) != "[map[]]" {
			proyectos = append(proyectos, proyectosP...)
		}
		// list proyectos hijos
		errProyectosH := request.GetJson(beego.AppConfig.String("ProyectoAcademicoService")+"proyecto_academico_institucion?query=Activo:true,NivelFormacionId.NivelFormacionPadreId.Id:"+fmt.Sprintf("%v", idNiv)+"&sortby=Nombre&order=asc&limit=0&fields=Id,Nombre", &proyectosH)
		if errProyectosH == nil {
			if fmt.Sprintf("%v", proyectosH) != "[map[]]" {
				proyectos = append(proyectos, proyectosH...)
			}

			if len(proyectos) > 0 {
				errCalendarios := request.GetJson(beego.AppConfig.String("EventoService")+"calendario?query=Activo:true,Nivel:"+fmt.Sprintf("%v", idNiv)+",PeriodoId:"+fmt.Sprintf("%v", idPer)+"&limit=0&sortby=Id&order=desc", &calendarios)
				if errCalendarios == nil && fmt.Sprintf("%v", calendarios) != "[map[]]" {

					for _, proyecto := range proyectos {
						IdPro := int(proyecto["Id"].(float64))
						CalendarioId = "0"
						for _, calendario := range calendarios {
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
							if CalendarioId != "0" {
								proyectoInfo := map[string]interface{}{
									"ProyectoId":          IdPro,
									"NombreProyecto":      proyecto["Nombre"],
									"CalendarioID":        CalendarioId,
									"CalendarioExtension": false,
									"EventoInscripcion":   nil,
								}
								proyectosArrMap = append(proyectosArrMap, proyectoInfo)
								break
							}
						}
					}

					if len(proyectosArrMap) > 0 {
						for i := range proyectosArrMap {
							proyectosArrMap[i]["Proceso"] = []map[string]interface{}{}
							errEvento := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento/?query=ProcesoId__CalendarioID__Id:"+proyectosArrMap[i]["CalendarioID"].(string)+",Activo:true&limit=0", &calendarioEventos)
							if errEvento == nil && fmt.Sprintf("%v", calendarioEventos) != "[map[]]" {

								procesosPorId := make(map[string]map[string]interface{})
								var lista_procesos []map[string]interface{}
								for _, Evento := range calendarioEventos {
									dependenciaEvento, ok := parseDependenciaEvento(Evento["DependenciaId"])
									if ok && !dependenciaIncluyeProyecto(dependenciaEvento, proyectosArrMap[i]["ProyectoId"].(int)) {
										continue
									}

									nombreCatalogo, descripcionCatalogo, codAbrEvento := datosEventoCatalogoCompleto(Evento["EventoCatalogoId"])
									nombreEvento := strings.ToUpper(nombreCatalogo)
									nombreProceso, _, codAbrProceso := datosProcesoCatalogo(Evento["ProcesoId"])
									pago := strings.Contains(nombreEvento, "PAGO")
									procesoId := ""
									if procesoMap, ok := Evento["ProcesoId"].(map[string]interface{}); ok {
										if id, ok := procesoMap["Id"].(float64); ok {
											procesoId = fmt.Sprintf("%.f", id)
										}
									}
									if fechaParticular, ok := fechaParticularProyecto(dependenciaEvento, proyectosArrMap[i]["ProyectoId"].(int)); ok {
										evento_x := map[string]interface{}{
											"ActividadParticular": true,
											"EventoId":            Evento["Id"],
											"EventoCatalogoId":    Evento["EventoCatalogoId"],
											"ProcesoId":           Evento["ProcesoId"],
											"NombreProceso":       nombreProceso,
											"CodigoProceso":       codAbrProceso,
											"NombreEvento":        descripcionCatalogo,
											"FechaInicioEvento":   fechaParticular["Inicio"],
											"FechaFinEvento":      fechaParticular["Fin"],
											"CodigoAbreviacion":   codAbrEvento,
											"Pago":                pago,
										}
										if procesoId != "" {
											if _, existe := procesosPorId[procesoId]; !existe {
												procesosPorId[procesoId] = map[string]interface{}{
													"ProcesoId":         procesoId,
													"Proceso":           Evento["ProcesoId"],
													"NombreProceso":     nombreProceso,
													"CodigoAbreviacion": codAbrProceso,
													"Eventos":           []map[string]interface{}{},
												}
												lista_procesos = append(lista_procesos, procesosPorId[procesoId])
											}
											procesosPorId[procesoId]["Eventos"] = append(procesosPorId[procesoId]["Eventos"].([]map[string]interface{}), evento_x)
										}
									}
								}
								proyectosArrMap[i]["Proceso"] = lista_procesos
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
