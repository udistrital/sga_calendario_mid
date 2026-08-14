package services

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/sga_calendario_mid/helpers"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
	"github.com/udistrital/utils_oas/time_bogota"
)

const fechaGenericaClonacion = "2000-01-01T00:00:00-05:00"
const dependenciaVaciaEvento = `{"proyectos":[],"fechas":[]}`

func PostCalendario(data []byte, usuario string) (interface{}, error) {
	var calendario map[string]interface{}
	var calendarioParam []map[string]interface{}
	var proceso []map[string]interface{}

	var dataPost map[string]interface{}
	if err := json.Unmarshal(data, &dataPost); err != nil {
		return nil, errors.New("error del servicio PostCalendario: solicitud inválida")
	}
	terceroID, err := helpers.TerceroIDFromPayload(dataPost)
	if err != nil {
		return nil, errors.New("error del servicio PostCalendario: " + err.Error())
	}
	idCalendario, ok := helpers.IDToString(dataPost["Id"])
	if !ok {
		return nil, errors.New("error del servicio PostCalendario: calendario destino inválido")
	}
	idPeriodo, ok := helpers.IDToString(dataPost["PeriodoIdClone"])
	if !ok {
		return nil, errors.New("error del servicio PostCalendario: periodo origen inválido")
	}
	idNivel, ok := helpers.IDToString(dataPost["NivelClone"])
	if !ok {
		return nil, errors.New("error del servicio PostCalendario: nivel origen inválido")
	}

	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario/"+idCalendario, &calendario); err != nil || calendario == nil || calendario["Type"] == "error" || calendario["Id"] == nil {
		return nil, errors.New("error del servicio PostCalendario: no se encontró calendario destino")
	}

	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario?query=Activo:true,PeriodoId:"+idPeriodo+",Nivel:"+idNivel+"&sortby=Id&order=desc", &calendarioParam); err != nil {
		return nil, errors.New("error del servicio PostCalendario: no fue posible consultar calendario origen")
	}
	if len(calendarioParam) == 0 || calendarioParam[0]["Id"] == nil {
		return nil, errors.New("error del servicio PostCalendario: no se encontró calendario origen activo para el periodo y nivel seleccionados")
	}
	idCalendarioParam, ok := helpers.IDToString(calendarioParam[0]["Id"])
	if !ok {
		return nil, errors.New("error del servicio PostCalendario: calendario origen inválido")
	}

	if err := request.GetJson(beego.AppConfig.String("EventoService")+"proceso?query=CalendarioID__Id:"+idCalendarioParam+"&limit=0", &proceso); err != nil {
		return nil, errors.New("error del servicio PostCalendario: no fue posible consultar procesos del calendario origen")
	}
	if len(proceso) == 0 || proceso[0]["Id"] == nil {
		return nil, errors.New("error del servicio PostCalendario: no se encontraron procesos en el calendario origen")
	}

	for _, procesoOrigen := range proceso {
		idProcesoOrigen, ok := helpers.IDToString(procesoOrigen["Id"])
		if !ok {
			return nil, errors.New("error del servicio PostCalendario: proceso origen inválido")
		}

		procesoOrigen["Id"] = 0
		procesoOrigen["CalendarioID"] = calendario

		var procesoClonado map[string]interface{}
		if err := request.SendJson(beego.AppConfig.String("EventoService")+"/proceso", "POST", &procesoClonado, procesoOrigen); err != nil || helpers.EventosPostResponseInvalid(procesoClonado) {
			return nil, errors.New("error del servicio PostCalendario: no fue posible crear un proceso clonado")
		}
		procesoOrigen["Id"] = procesoClonado["Id"]

		var actividades []map[string]interface{}
		if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento?query=ProcesoId__Id:"+idProcesoOrigen+"&limit=0", &actividades); err != nil {
			return nil, errors.New("error del servicio PostCalendario: no fue posible consultar actividades del proceso origen")
		}
		if len(actividades) == 0 || actividades[0]["Id"] == nil {
			continue
		}

		for _, actividadOrigen := range actividades {
			idActividadOrigen, ok := helpers.IDToString(actividadOrigen["Id"])
			if !ok {
				return nil, errors.New("error del servicio PostCalendario: actividad origen inválida")
			}

			actividadOrigen["Id"] = 0
			actividadOrigen["ProcesoId"] = procesoOrigen
			actividadOrigen["FechaInicio"] = fechaGenericaClonacion
			actividadOrigen["FechaFin"] = fechaGenericaClonacion
			actividadOrigen["DependenciaId"] = dependenciaVaciaEvento
			helpers.SetTerceroID(actividadOrigen, terceroID)

			var actividadClonada map[string]interface{}
			if err := request.SendJson(beego.AppConfig.String("EventoService")+"/calendario_evento", "POST", &actividadClonada, actividadOrigen); err != nil || helpers.EventosPostResponseInvalid(actividadClonada) {
				return nil, errors.New("error del servicio PostCalendario: no fue posible crear una actividad clonada")
			}

			var publicos []map[string]interface{}
			if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento_tipo_publico?query=CalendarioEventoId__Id:"+idActividadOrigen+"&limit=0", &publicos); err != nil {
				return nil, errors.New("error del servicio PostCalendario: no fue posible consultar públicos dirigidos de la actividad origen")
			}
			for _, publicoOrigen := range publicos {
				if publicoOrigen["Id"] == nil {
					continue
				}
				publicoOrigen["Id"] = 0
				publicoOrigen["CalendarioEventoId"] = actividadClonada

				var publicoClonado map[string]interface{}
				if err := request.SendJson(beego.AppConfig.String("EventoService")+"/calendario_evento_tipo_publico", "POST", &publicoClonado, publicoOrigen); err != nil || helpers.EventosPostResponseInvalid(publicoClonado) {
					return nil, errors.New("error del servicio PostCalendario: no fue posible crear un público dirigido clonado")
				}
			}
		}
	}

	return requestresponse.APIResponseDTO(true, 200, calendario), nil
}

func PostCalendarioPadre(data []byte, usuario string) (interface{}, error) {
	var calendario map[string]interface{}
	var calendarioParam []map[string]interface{}
	var proceso []map[string]interface{}
	var calendarioEvento []map[string]interface{}
	var calendarioEventoTipoPublico []map[string]interface{}
	var resultadoPost map[string]interface{}
	var resultadoPostResponsable map[string]interface{}
	var resultado map[string]interface{}
	var errCalendarioParam = errors.New("")
	var errorGetAll bool

	var dataPost map[string]interface{}
	if err := json.Unmarshal(data, &dataPost); err == nil {
		terceroID, terceroErr := helpers.TerceroIDFromPayload(dataPost)
		if terceroErr != nil {
			return nil, errors.New("error del servicio PostCalendarioPadre: " + terceroErr.Error())
		}
		idCalendario, err := calendarioDestinoClonacion(dataPost, usuario)
		if err != nil {
			return nil, err
		}
		idCalendarioPadre := fmt.Sprintf("%.f", dataPost["IdPadre"].(map[string]interface{})["Id"])
		errCalendario := request.GetJson(beego.AppConfig.String("EventoService")+"calendario/"+idCalendario, &calendario)
		if errCalendario == nil {
			if calendario != nil {
				if dataPost["Nivel"].(float64) == calendario["Nivel"].(float64) {
					errCalendarioParam = request.GetJson(beego.AppConfig.String("EventoService")+"calendario?query=Id:"+idCalendarioPadre, &calendarioParam)
				} else {
					errCalendarioParam = request.GetJson(beego.AppConfig.String("EventoService")+"calendario?query=Id:"+idCalendarioPadre, &calendarioParam)
				}

				if errCalendarioParam == nil {
					if len(calendarioParam) > 0 && calendarioParam[0]["Id"] != nil {
						idCalendarioParam := fmt.Sprintf("%.f", calendarioParam[0]["Id"].(float64))

						// persistir procesos si el calendario que se esta clonando los tiene
						errProceso := request.GetJson(beego.AppConfig.String("EventoService")+"proceso?query=CalendarioID__Id:"+idCalendarioParam+"&limit=0", &proceso)
						if errProceso == nil {
							resultado = map[string]interface{}{
								"Id": idCalendario,
							}
							if len(proceso) > 0 && proceso[0]["Id"] != nil {
								for _, procesoOrigen := range proceso {
									idOld := fmt.Sprintf("%.f", procesoOrigen["Id"].(float64))
									procesoOrigen["Id"] = 0
									procesoOrigen["CalendarioID"] = calendario

									errProcesoPost := request.SendJson(beego.AppConfig.String("EventoService")+"/proceso", "POST", &resultadoPost, procesoOrigen)
									if errProcesoPost == nil && fmt.Sprintf("%v", resultadoPost["System"]) != "map[]" && resultadoPost["Id"] != nil {
										if resultadoPost["Status"] != 400 {
											procesoOrigen["Id"] = resultadoPost["Id"]

											// persistir calendario_evento si el proceso que se esta clonando esta asociado en el campo proceso_id del calendario_evento
											errCalendarioEvento := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento?query=ProcesoId__Id:"+idOld+"&limit=0", &calendarioEvento)
											if errCalendarioEvento == nil {
												if len(calendarioEvento) > 0 && calendarioEvento[0]["Id"] != nil {
													for _, cEvento := range calendarioEvento {
														idCalendarioEventoOld := fmt.Sprintf("%.f", cEvento["Id"].(float64))
														cEvento["Id"] = 0
														cEvento["ProcesoId"] = procesoOrigen
														cEvento["FechaInicio"] = fechaGenericaClonacion
														cEvento["FechaFin"] = fechaGenericaClonacion
														cEvento["DependenciaId"] = dependenciaVaciaEvento
														helpers.SetTerceroID(cEvento, terceroID)

														errCalendarioEventoPost := request.SendJson(beego.AppConfig.String("EventoService")+"/calendario_evento", "POST", &resultadoPost, cEvento)
														if errCalendarioEventoPost == nil && fmt.Sprintf("%v", resultadoPost["System"]) != "map[]" && resultadoPost["Id"] != nil {
															if resultadoPost["Status"] != 400 {
																errCalendarioEventoTipoPublico := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento_tipo_publico?query=CalendarioEventoId__Id:"+idCalendarioEventoOld+"&limit=0", &calendarioEventoTipoPublico)
																if errCalendarioEventoTipoPublico == nil {
																	for _, cEventoTipoPublico := range calendarioEventoTipoPublico {
																		cEventoTipoPublico["Id"] = 0
																		cEventoTipoPublico["CalendarioEventoId"] = resultadoPost
																		if err := request.SendJson(beego.AppConfig.String("EventoService")+"/calendario_evento_tipo_publico", "POST", &resultadoPostResponsable, cEventoTipoPublico); err != nil {
																			errorGetAll = true
																			logs.Error(err.Error())
																		}
																	}
																} else {
																	errorGetAll = true
																	logs.Error(errCalendarioEventoTipoPublico.Error())
																}
															} else {
																errorGetAll = true
																logs.Error(errCalendarioEventoPost.Error())
															}
														} else {
															errorGetAll = true
														}
													}
												}
											} else {
												errorGetAll = true
												logs.Error(errCalendarioEvento.Error())
											}
										} else {
											errorGetAll = true
											logs.Error(errProcesoPost.Error())
										}
									} else {
										errorGetAll = true
									}
								}
							}
						} else {
							errorGetAll = true
							logs.Error(errProceso.Error())
						}
					} else {
						errorGetAll = true
					}
				} else {
					errorGetAll = true
					logs.Error(errCalendario.Error())
				}
			} else {
				errorGetAll = true
			}
		} else {
			errorGetAll = true
			logs.Error(errCalendario.Error())
		}
	} else {
		errorGetAll = true
		logs.Error(err.Error())
	}

	if !errorGetAll {
		return requestresponse.APIResponseDTO(true, 200, resultado), nil
	} else {
		return nil, errors.New("error del servicio PostCalendarioPadre: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
	}
}

func calendarioDestinoClonacion(dataPost map[string]interface{}, usuario string) (string, error) {
	if id, ok := dataPost["Id"].(float64); ok && id > 0 {
		return fmt.Sprintf("%.f", id), nil
	}
	periodoID := fmt.Sprintf("%.f", dataPost["PeriodoId"].(float64))
	nivelID := fmt.Sprintf("%.f", dataPost["Nivel"].(float64))
	existe, err := calendarioActivoPorPeriodoNivel(periodoID, nivelID)
	if err != nil {
		return "", errors.New("error del servicio PostCalendarioPadre: no fue posible validar calendario existente")
	}
	if existe {
		return "", errors.New("Ya existe un calendario activo para el periodo y nivel seleccionados")
	}
	calendarioNuevo := map[string]interface{}{
		"Nombre":            dataPost["Nombre"],
		"DependenciaId":     `{"proyectos":[]}`,
		"DocumentoId":       dataPost["DocumentoId"],
		"PeriodoId":         dataPost["PeriodoId"],
		"AplicacionId":      0,
		"Nivel":             dataPost["Nivel"],
		"Activo":            dataPost["Activo"],
		"FechaCreacion":     time_bogota.TiempoBogotaFormato(),
		"FechaModificacion": time_bogota.TiempoBogotaFormato(),
	}
	var resultado map[string]interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario", "POST", &resultado, calendarioNuevo); err != nil {
		return "", errors.New("error del servicio PostCalendarioPadre: no fue posible crear calendario destino")
	}
	id, ok := resultado["Id"].(float64)
	if !ok || id <= 0 || resultado["Status"] == 400 || resultado["Type"] == "error" {
		return "", errors.New("error del servicio PostCalendarioPadre: no fue posible crear calendario destino")
	}
	RegistrarAuditoria("calendario", int(id), "POST", nil, resultado, usuario, "PostCalendarioPadre/calendario")
	return fmt.Sprintf("%.f", id), nil
}
