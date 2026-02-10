package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/sga_calendario_mid/models"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
	"github.com/udistrital/utils_oas/time_bogota"
	"golang.org/x/sync/errgroup"
)

func GetAll() (interface{}, error) {
	var resultados []map[string]interface{}
	var calendarios []map[string]interface{}
	var errorGetAll bool
	var message string
	wge := new(errgroup.Group)

	errCalendario := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario?limit=0&sortby=Id&order=desc", &calendarios)
	if errCalendario == nil {
		if len(calendarios[0]) > 0 && fmt.Sprintf("%v", calendarios[0]["Nombre"]) != "map[]" {
			fmt.Println(len(calendarios))
			//Limitación de la cantidad de hilos a utilizar, valores negativas representan sin limite
			wge.SetLimit(10)
			for _, calendario := range calendarios {

				calendario := calendario
				//Declaración función anonima
				wge.Go(func() error {
					var ListarCalendario bool = false
					var periodo map[string]interface{}
					var errPeriodo error

					if calendario["CalendarioPadreId"] == nil {
						ListarCalendario = true
					} else if calendario["Activo"].(bool) == true && calendario["CalendarioPadreId"].(map[string]interface{})["Activo"].(bool) == false {
						ListarCalendario = true
					} else {
						ListarCalendario = false
					}
					if calendario["AplicaExtension"].(bool) == false {

						if ListarCalendario {
							periodoID := fmt.Sprintf("%.f", calendario["PeriodoId"].(float64))
							errPeriodo = request.GetJson("http://"+beego.AppConfig.String("ParametroService")+"periodo/"+periodoID, &periodo)
							if errPeriodo == nil {
								periodoNombre := ""
								if periodo["Status"] == "200" {
									periodoNombre = periodo["Data"].(map[string]interface{})["Nombre"].(string)
								}
								resultado := map[string]interface{}{
									"Id":      calendario["Id"].(float64),
									"Nombre":  calendario["Nombre"].(string),
									"Nivel":   calendario["Nivel"].(float64),
									"Activo":  calendario["Activo"].(bool),
									"Periodo": periodoNombre,
								}
								resultados = append(resultados, resultado)
							} else {
								return errPeriodo
							}
						}

					}
					return nil
				})

			}
			//Si existe error, se realiza
			if err := wge.Wait(); err != nil {
				errorGetAll = true
			}
		} else {
			errorGetAll = false
			message += "No data found"
		}
	} else {
		errorGetAll = true
		message += errCalendario.Error()
	}

	if !errorGetAll {
		return requestresponse.APIResponseDTO(true, 200, resultados), nil
	} else {
		return nil, errors.New("error del servicio GetAll: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
	}
}

func GetOnePorId(idCalendario string) (interface{}, error) {
	var resultado map[string]interface{}
	var resultados []map[string]interface{}
	var versionCalendario map[string]interface{}
	var versionCalendarioResultado []map[string]interface{}
	var calendarioPadreID map[string]interface{}
	var documento map[string]interface{}
	var resolucion map[string]interface{}
	var procesoArr []string
	var proceso map[string]interface{}
	var procesoResultado []map[string]interface{}
	var actividad map[string]interface{}
	var procesoAdd map[string]interface{}
	var responsableTipoP map[string]interface{}
	var responsableList []map[string]interface{}

	if resultado["Type"] != "error" {
		// consultar calendario evento por tipo evento
		var calendarios []map[string]interface{}
		errcalendario := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario_evento?query=TipoEventoId__Id.CalendarioID__Id:"+idCalendario, &calendarios)

		if errcalendario == nil {
			if calendarios[0]["Id"] != nil {

				// ver si el calendario esta ligado a un padre
				if calendarios[0]["TipoEventoId"].(map[string]interface{})["CalendarioID"].(map[string]interface{})["CalendarioPadreId"] != nil {

					calendarioPadreID = calendarios[0]["TipoEventoId"].(map[string]interface{})["CalendarioID"].(map[string]interface{})["CalendarioPadreId"].(map[string]interface{})
					padreID := fmt.Sprintf("%.f", calendarioPadreID["Id"].(float64))

					// obtener informacion calendario padre si existe
					if padreID != "" {

						// versionCalendario = map[string]interface{}{
						// 	"Id":     padreID,
						// 	"Nombre": calendarios[0]["TipoEventoId"].(map[string]interface{})["CalendarioID"].(map[string]interface{})["CalendarioPadreId"].(map[string]interface{})["Nombre"],
						// }
						// versionCalendarioResultado = append(versionCalendarioResultado, versionCalendario)

						var calendariosPadre map[string]interface{}
						errcalendarioPadre := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario/"+padreID, &calendariosPadre)
						if calendariosPadre != nil {
							if errcalendarioPadre == nil {
								versionCalendario = map[string]interface{}{
									"Id":     padreID,
									"Nombre": calendariosPadre["Nombre"],
								}
								versionCalendarioResultado = append(versionCalendarioResultado, versionCalendario)
							} else {
								logs.Error(errcalendarioPadre.Error())
							}
						}
					} else {
						versionCalendario = map[string]interface{}{
							"Id":     "",
							"Nombre": "",
						}
						versionCalendarioResultado = append(versionCalendarioResultado, versionCalendario)
					}
				}

				documento = calendarios[0]["TipoEventoId"].(map[string]interface{})["CalendarioID"].(map[string]interface{})
				documentoID := fmt.Sprintf("%.f", documento["DocumentoId"].(float64))
				var documentos map[string]interface{}
				errdocumento := request.GetJson("http://"+beego.AppConfig.String("DocumentosService")+"documento/"+documentoID, &documentos)

				if errdocumento == nil {
					if documentos != nil {
						metadatoJSON := documentos["Metadatos"].(string)
						var metadato models.Metadatos
						json.Unmarshal([]byte(metadatoJSON), &metadato)

						resolucion = map[string]interface{}{
							"Id":         documentos["Id"],
							"Enlace":     documentos["Enlace"],
							"Resolucion": metadato.Resolucion,
							"Anno":       metadato.Anno,
							"Nombre":     documentos["Nombre"],
						}
					} else {
						return requestresponse.APIResponseDTO(true, 200, documentos), nil
					}

				} else {
					logs.Error(errdocumento.Error())
				}

				// recorrer el calendario para agrupar las actividades por proceso
				for _, calendario := range calendarios {
					proceso = nil
					proceso = map[string]interface{}{
						"NombreProceso": calendario["TipoEventoId"].(map[string]interface{})["Id"].(float64),
					}

					procesoResultado = append(procesoResultado, proceso)
				}

				for _, procesoList := range procesoResultado {

					procesoArr = append(procesoArr, fmt.Sprintf("%.f", procesoList["NombreProceso"].(float64)))

				}

				procesoResultado = nil

				m := make(map[string]bool)
				arr := make([]string, 0)

				// eliminar procesos duplicados
				for curIndex := 0; curIndex < len((*&procesoArr)); curIndex++ {
					curValue := (*&procesoArr)[curIndex]
					if has := m[curValue]; !has {
						m[curValue] = true
						arr = append(arr, curValue)
					}
				}
				*&procesoArr = arr

				wge := new(errgroup.Group)
				var mutex sync.Mutex // Mutex para

				wge.SetLimit(10)
				for _, procesoList := range arr {

					procesoList := procesoList

					wge.Go(func() error {
						var actividadResultado []map[string]interface{}
						var procesos []map[string]interface{}
						errproceso := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario_evento?query=TipoEventoId.Id:"+procesoList+"&TipoEventoId__Id.CalendarioID__Id:"+idCalendario, &procesos)

						if errproceso == nil {
							if procesos != nil {
								for _, proceso := range procesos {

									// consultar responsables
									// var responsableString = ""
									responsableTipoP = nil
									for _, responsable := range procesos {

										calendarioResponsableID := fmt.Sprintf("%.f", responsable["Id"].(float64))
										var responsables []map[string]interface{}
										errresponsable := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario_evento_tipo_publico?query=CalendarioEventoId__Id:"+calendarioResponsableID, &responsables)

										if errresponsable == nil {
											if responsables != nil {
												responsableList = nil
												for _, listRresponsable := range responsables {
													var responsablesID map[string]interface{} = listRresponsable["TipoPublicoId"].(map[string]interface{})
													// responsableID := fmt.Sprintf(responsablesID["Nombre"].(string))
													// responsableString = responsableID + ", " + responsableString

													responsableTipoP = map[string]interface{}{
														"responsableID": responsablesID["Id"].(float64),
														"Nombre":        fmt.Sprintf("%s", responsablesID["Nombre"].(string)),
													}
													responsableList = append(responsableList, responsableTipoP)
												}
											}
										} else {
											logs.Error(errresponsable.Error())
										}
									}

									actividad = nil
									actividad = map[string]interface{}{
										"actividadId":   proceso["Id"].(float64),
										"Nombre":        proceso["Nombre"].(string),
										"Descripcion":   proceso["Descripcion"].(string),
										"FechaInicio":   proceso["FechaInicio"].(string),
										"FechaFin":      proceso["FechaFin"].(string),
										"Activo":        proceso["Activo"].(bool),
										"TipoEventoId":  proceso["TipoEventoId"].(map[string]interface{}),
										"EventoPadreId": proceso["EventoPadreId"],
										"Responsable":   responsableList,
									}

									actividadResultado = append(actividadResultado, actividad)

								}

								procesoAdd = nil
								procesoAdd = map[string]interface{}{
									"Proceso":     procesos[0]["TipoEventoId"].(map[string]interface{})["Nombre"].(string),
									"Actividades": actividadResultado,
								}

								mutex.Lock()
								procesoResultado = append(procesoResultado, procesoAdd)
								mutex.Unlock()

							} else {
								return nil
							}

						} else {
							logs.Error(errproceso.Error())
							return errproceso
						}

						return nil
					})
				}
				//Si existe error, se realiza
				if err := wge.Wait(); err != nil {
					return requestresponse.APIResponseDTO(false, 400, nil, err), err
				}

				calendarioAux := calendarios[0]["TipoEventoId"].(map[string]interface{})["CalendarioID"].(map[string]interface{})
				resultado = map[string]interface{}{
					"Id":              idCalendario,
					"Nombre":          calendarioAux["Nombre"].(string),
					"PeriodoId":       calendarioAux["PeriodoId"].(float64),
					"Activo":          calendarioAux["Activo"].(bool),
					"Nivel":           calendarioAux["Nivel"].(float64),
					"ListaCalendario": versionCalendarioResultado,
					"resolucion":      resolucion,
					"proceso":         procesoResultado,
				}
				resultados = append(resultados, resultado)

				return requestresponse.APIResponseDTO(true, 200, resultados), nil

			} else {
				var calendario map[string]interface{}
				errcalendario := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario/"+idCalendario, &calendario)
				if errcalendario == nil {
					if calendario["Id"] != nil {

						if calendario["CalendarioPadreId"] != nil {
							padreID := fmt.Sprintf("%.f", calendario["CalendarioPadreId"].(map[string]interface{})["Id"].(float64))
							versionCalendario = map[string]interface{}{
								"Id":     padreID,
								"Nombre": calendario["CalendarioPadreId"].(map[string]interface{})["Nombre"],
							}
							versionCalendarioResultado = append(versionCalendarioResultado, versionCalendario)
						}

						documentoID := fmt.Sprintf("%.f", calendario["DocumentoId"].(float64))
						var documentos map[string]interface{}

						errdocumento := request.GetJson("http://"+beego.AppConfig.String("DocumentosService")+"documento/"+documentoID, &documentos)

						if errdocumento == nil {

							if documentos != nil {

								metadatoJSON := documentos["Metadatos"].(string)
								var metadato models.Metadatos
								json.Unmarshal([]byte(metadatoJSON), &metadato)

								resolucion = map[string]interface{}{
									"Id":         documentos["Id"],
									"Enlace":     documentos["Enlace"],
									"Resolucion": metadato.Resolucion,
									"Anno":       metadato.Anno,
									"Nombre":     documentos["Nombre"],
								}
							} else {
								return requestresponse.APIResponseDTO(true, 200, documentos), nil
							}

						} else {
							logs.Error(errdocumento.Error())
						}

						resultado = map[string]interface{}{
							"Id":              idCalendario,
							"Nombre":          calendario["Nombre"].(string),
							"PeriodoId":       calendario["PeriodoId"].(float64),
							"Activo":          calendario["Activo"].(bool),
							"Nivel":           calendario["Nivel"].(float64),
							"ListaCalendario": versionCalendarioResultado,
							"resolucion":      resolucion,
							"proceso":         procesoResultado,
						}
						resultados = append(resultados, resultado)

						return requestresponse.APIResponseDTO(true, 200, resultados), nil
					}

				} else {
					return requestresponse.APIResponseDTO(true, 200, calendarios), nil
				}

			}

		} else {
			logs.Error(errcalendario.Error())
		}

	} else {
		if resultado["Body"] == "<QuerySeter> no row found" {
			return nil, errors.New("error del servicio GetOnePorId: <QuerySeter> no row found")
		} else {
			return nil, errors.New("error del servicio GetOnePorId: <QuerySeter> no row found")
		}
	}
	return nil, errors.New("error del servicio GetOnePorId: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")

}

func PutInhabilitarCalendario(idCalendario string, data []byte) (interface{}, error) {
	var calendario map[string]interface{}
	var tipoEvento []map[string]interface{}
	var calendarioEvento []map[string]interface{}
	var calendarioEventoTipoPublico []map[string]interface{}
	var tipoPublico map[string]interface{}
	var resultado map[string]interface{}
	var dataPut map[string]interface{}
	var message string
	var success bool = true
	alertas := []interface{}{"Response:"}
	if err := json.Unmarshal(data, &dataPut); err == nil {

		errCalendario := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario/"+idCalendario, &calendario)
		if errCalendario == nil {
			if calendario != nil {

				if fmt.Sprintf("%v", calendario["DependenciaParticularId"]) == "" {
					calendario["DependenciaParticularId"] = "{}"
				}
				calendario["Activo"] = false

				errCalendario := request.SendJson("http://"+beego.AppConfig.String("EventoService")+"calendario/"+idCalendario, "PUT", &resultado, calendario)
				if resultado["Type"] == "error" || errCalendario != nil || resultado["Status"] == "404" || resultado["Message"] != nil {
					success = false
				} else {

					errCalendario := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"tipo_evento?query=CalendarioID__Id:"+idCalendario, &tipoEvento)
					if errCalendario == nil {
						if tipoEvento != nil && tipoEvento[0] != nil && len(tipoEvento[0]) > 0 {

							for _, tEvento := range tipoEvento {

								idEvento := fmt.Sprintf("%.f", tEvento["Id"].(float64))
								tEvento["Activo"] = false

								errCalendario := request.SendJson("http://"+beego.AppConfig.String("EventoService")+"tipo_evento/"+idEvento, "PUT", &resultado, tEvento)
								if resultado["Type"] == "error" || errCalendario != nil || resultado["Status"] == "404" || resultado["Message"] != nil {
									success = false
								} else {

									errCalendario := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario_evento?query=TipoEventoId__Id:"+idEvento, &calendarioEvento)
									if errCalendario == nil {
										if calendarioEvento != nil && calendarioEvento[0] != nil && len(calendarioEvento[0]) > 0 {

											for _, cEvento := range calendarioEvento {

												idCalendarioEvento := fmt.Sprintf("%.f", cEvento["Id"].(float64))
												cEvento["Activo"] = false

												errCalendario := request.SendJson("http://"+beego.AppConfig.String("EventoService")+"calendario_evento/"+idCalendarioEvento, "PUT", &resultado, cEvento)
												if resultado["Type"] == "error" || errCalendario != nil || resultado["Status"] == "404" || resultado["Message"] != nil {
													success = false
												} else {

													errCalendario := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario_evento_tipo_publico?query=CalendarioEventoId__Id:"+idCalendarioEvento, &calendarioEventoTipoPublico)
													if errCalendario == nil {
														if calendarioEventoTipoPublico != nil && calendarioEventoTipoPublico[0] != nil && len(calendarioEventoTipoPublico[0]) > 0 {

															for _, cEventoTipoPublico := range calendarioEventoTipoPublico {

																idCalendarioEventoTipoPublico := fmt.Sprintf("%.f", cEventoTipoPublico["Id"].(float64))
																cEventoTipoPublico["Activo"] = false

																request.SendJson("http://"+beego.AppConfig.String("EventoService")+"calendario_evento_tipo_publico/"+idCalendarioEventoTipoPublico, "PUT", &resultado, cEventoTipoPublico)
																if resultado["Type"] == "error" || resultado["Status"] == "404" || resultado["Message"] != nil {
																	success = false
																} else {

																	idTipoPublico := fmt.Sprintf("%.f", cEventoTipoPublico["TipoPublicoId"].(map[string]interface{})["Id"].(float64))

																	errCalendario := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"tipo_publico/"+idTipoPublico, &tipoPublico)
																	if errCalendario == nil {
																		if tipoPublico != nil && len(tipoPublico) > 0 {

																			tipoPublico["Activo"] = false

																			errCalendario := request.SendJson("http://"+beego.AppConfig.String("EventoService")+"tipo_publico/"+idTipoPublico, "PUT", &resultado, tipoPublico)
																			if resultado["Type"] == "error" || errCalendario != nil || resultado["Status"] == "404" || resultado["Message"] != nil {
																				success = false
																			}

																		}
																	}

																}

															}

														}
													}

												}

											}

										}
									}

								}

							}

						}
					}
				}
			} else {
				return requestresponse.APIResponseDTO(true, 200, calendario), nil
			}
			logs.Error(calendario)
			return requestresponse.APIResponseDTO(true, 200, calendario), nil
		} else {
			logs.Error(errCalendario)
			return nil, errors.New("error del servicio PutInhabilitarCalendario: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
		}

	} else {
		success = false
		message += err.Error()
	}
	if success {
		return requestresponse.APIResponseDTO(success, 200, alertas), nil
	} else {
		if message != "" {
			return nil, errors.New("error del servicio PutInhabilitarCalendario: " + message)
		} else {
			return nil, errors.New("error del servicio PutInhabilitarCalendario: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
		}
	}
}

func PostCalendarioHijo(data []byte) (interface{}, error) {
	var AuxCalendarioHijo map[string]interface{}
	var calendarioHijoPost map[string]interface{}
	var CalendarioPadreId interface{}
	var CalendarioPadre []map[string]interface{}
	var CalendarioPadrePut map[string]interface{}

	if err := json.Unmarshal(data, &AuxCalendarioHijo); err == nil {

		CalendarioHijo := map[string]interface{}{
			"Nombre":                  AuxCalendarioHijo["Nombre"],
			"DependenciaId":           AuxCalendarioHijo["DependenciaId"],
			"DocumentoId":             AuxCalendarioHijo["DocumentoId"],
			"PeriodoId":               AuxCalendarioHijo["PeriodoId"],
			"MultiplePeriodoId":       AuxCalendarioHijo["MultiplePeriodoId"],
			"AplicacionId":            0,
			"Nivel":                   AuxCalendarioHijo["Nivel"],
			"Activo":                  AuxCalendarioHijo["Activo"],
			"FechaCreacion":           time_bogota.TiempoBogotaFormato(),
			"FechaModificacion":       time_bogota.TiempoBogotaFormato(),
			"CalendarioPadreId":       map[string]interface{}{"Id": AuxCalendarioHijo["CalendarioPadreId"].(map[string]interface{})["Id"].(float64)},
			"DependenciaParticularId": "{}",
		}
		fmt.Println(AuxCalendarioHijo["CalendarioPadreId"].(map[string]interface{})["Id"])
		errCalendarioHijo := request.SendJson("http://"+beego.AppConfig.String("EventoService")+"calendario", "POST", &calendarioHijoPost, CalendarioHijo)
		CalendarioPadreId = calendarioHijoPost["CalendarioPadreId"].(map[string]interface{})["Id"]

		if errCalendarioHijo == nil && fmt.Sprintf("%v", calendarioHijoPost["System"]) != "map[]" && calendarioHijoPost["Id"] != nil {
			if calendarioHijoPost["Status"] != 400 {

				//Se trae el calendario padre con el Id obtenido por el calendario hijo
				IdPadre := fmt.Sprintf("%.f", CalendarioPadreId.(float64))
				errCalendarioPadre := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario?query=Id:"+IdPadre, &CalendarioPadre)
				if errCalendarioPadre == nil {
					if CalendarioPadre[0]["Id"] != nil {

						//Se cambia el estado del calendario Padre a inactivo
						CalendarioPadre[0]["Activo"] = false
						if fmt.Sprintf("%v", CalendarioPadre[0]["DependenciaParticularId"]) == "" {
							CalendarioPadre[0]["DependenciaParticularId"] = "{}"
						}
						CalendarioPadreAux := CalendarioPadre[0]
						errCalendarioPadre := request.SendJson("http://"+beego.AppConfig.String("EventoService")+"calendario/"+IdPadre, "PUT", &CalendarioPadrePut, CalendarioPadreAux)
						if errCalendarioPadre == nil && fmt.Sprintf("%v", CalendarioPadrePut["System"]) != "map[]" && CalendarioPadrePut["Id"] != nil {
							if CalendarioPadrePut["Status"] != 400 {
								return requestresponse.APIResponseDTO(false, 200, calendarioHijoPost), nil
							} else {
								logs.Error(err)
								return nil, errors.New("error del servicio PostCalendarioHijo: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
							}
						} else {
							logs.Error(err)
							return nil, errors.New("error del servicio PostCalendarioHijo: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
						}
					}
				}
			} else {
				logs.Error(err)
				return nil, errors.New("error del servicio PostCalendarioHijo: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
			}

		} else {
			logs.Error(err)
			return nil, errors.New("error del servicio PostCalendarioHijo: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
		}
	} else {
		return nil, errors.New("error del servicio PostCalendarioHijo: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
	}
	return nil, errors.New("error del servicio PostCalendarioHijo: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
}

func GetCalendarInfo(idCalendario string) (interface{}, error) {
	var resultado map[string]interface{}
	var resultados []map[string]interface{}
	var actividadResultado []map[string]interface{}
	var versionCalendario map[string]interface{}
	var versionCalendarioResultado []map[string]interface{}
	var calendarioPadreID map[string]interface{}
	var documento map[string]interface{}
	var resolucion map[string]interface{}
	var procesoArr []string
	var proceso map[string]interface{}
	var procesoResultado []map[string]interface{}
	var actividad map[string]interface{}
	var procesoAdd map[string]interface{}
	var responsableTipoP map[string]interface{}
	var responsableList []map[string]interface{}
	var calendariosExtlist []map[string]interface{}
	var resolucionExt map[string]interface{}

	//var resolucion_ext map[string]interface{}

	if resultado["Type"] != "error" {
		// consultar calendario evento por tipo evento
		var calendarios []map[string]interface{}
		errcalendario := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario_evento?query=TipoEventoId__Id.CalendarioID__Id:"+idCalendario, &calendarios)
		if errcalendario == nil {
			if calendarios[0]["Id"] != nil {

				// ver si el calendario esta ligado a un padre
				if calendarios[0]["TipoEventoId"].(map[string]interface{})["CalendarioID"].(map[string]interface{})["CalendarioPadreId"] != nil {

					calendarioPadreID = calendarios[0]["TipoEventoId"].(map[string]interface{})["CalendarioID"].(map[string]interface{})["CalendarioPadreId"].(map[string]interface{})
					padreID := fmt.Sprintf("%.f", calendarioPadreID["Id"].(float64))

					// obtener informacion calendario padre si existe
					if padreID != "" {

						var calendariosPadre map[string]interface{}
						errcalendarioPadre := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario/"+padreID, &calendariosPadre)
						if calendariosPadre != nil {
							if errcalendarioPadre == nil {
								versionCalendario = map[string]interface{}{
									"Id":     padreID,
									"Nombre": calendariosPadre["Nombre"],
								}
								versionCalendarioResultado = append(versionCalendarioResultado, versionCalendario)
							} else {
								logs.Error(errcalendarioPadre.Error())
							}
						}

					} else {
						versionCalendario = map[string]interface{}{
							"Id":     "",
							"Nombre": "",
						}
						versionCalendarioResultado = append(versionCalendarioResultado, versionCalendario)
					}
				}

				documento = calendarios[0]["TipoEventoId"].(map[string]interface{})["CalendarioID"].(map[string]interface{})
				documentoID := fmt.Sprintf("%.f", documento["DocumentoId"].(float64))

				var documentos map[string]interface{}
				errdocumento := request.GetJson("http://"+beego.AppConfig.String("DocumentosService")+"documento/"+documentoID, &documentos)

				if errdocumento == nil {
					if documentos != nil {
						metadatoJSON := documentos["Metadatos"].(string)
						var metadato models.Metadatos
						json.Unmarshal([]byte(metadatoJSON), &metadato)

						resolucion = map[string]interface{}{
							"Id":         documentos["Id"],
							"Enlace":     documentos["Enlace"],
							"Resolucion": metadato.Resolucion,
							"Anno":       metadato.Anno,
							"Nombre":     documentos["Nombre"],
						}
					} else {
						return requestresponse.APIResponseDTO(true, 200, documentos), nil
					}

				} else {
					logs.Error(errdocumento.Error())
				}

				documentoExtID, ok := documento["DocumentoExtensionId"].(float64)

				if documentoExtID != 0 && ok {
					var documentosExt map[string]interface{}
					errdocumentoExt := request.GetJson("http://"+beego.AppConfig.String("DocumentosService")+"documento/"+fmt.Sprintf("%.f", documentoExtID), &documentosExt)

					if errdocumentoExt == nil {
						if documentosExt != nil {
							metadatoJSON := documentosExt["Metadatos"].(string)
							var metadato models.Metadatos
							json.Unmarshal([]byte(metadatoJSON), &metadato)

							resolucionExt = map[string]interface{}{
								"Id":         documentosExt["Id"],
								"Enlace":     documentosExt["Enlace"],
								"Resolucion": metadato.Resolucion,
								"Anno":       metadato.Anno,
								"Nombre":     documentosExt["Nombre"],
							}
						} else {
							return requestresponse.APIResponseDTO(true, 200, documentosExt), nil
						}

					} else {
						logs.Error(errdocumentoExt.Error())
					}
				}

				// recorrer el calendario para agrupar las actividades por proceso
				for _, calendario := range calendarios {
					proceso = nil
					proceso = map[string]interface{}{
						"NombreProceso": calendario["TipoEventoId"].(map[string]interface{})["Id"].(float64),
					}

					procesoResultado = append(procesoResultado, proceso)
				}

				for _, procesoList := range procesoResultado {

					procesoArr = append(procesoArr, fmt.Sprintf("%.f", procesoList["NombreProceso"].(float64)))

				}

				procesoResultado = nil

				m := make(map[string]bool)
				arr := make([]string, 0)

				// eliminar procesos duplicados
				for curIndex := 0; curIndex < len((*&procesoArr)); curIndex++ {
					curValue := (*&procesoArr)[curIndex]
					if has := m[curValue]; !has {
						m[curValue] = true
						arr = append(arr, curValue)
					}
				}
				*&procesoArr = arr

				for _, procesoList := range arr {

					var procesos []map[string]interface{}
					errproceso := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario_evento?query=TipoEventoId.Id:"+procesoList+"&TipoEventoId__Id.CalendarioID__Id:"+idCalendario, &procesos)

					if errproceso == nil {
						if procesos != nil {
							for _, proceso := range procesos {

								// consultar responsables
								// var responsableString = ""
								responsableTipoP = nil
								for _, responsable := range procesos {

									calendarioResponsableID := fmt.Sprintf("%.f", responsable["Id"].(float64))
									var responsables []map[string]interface{}
									errresponsable := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario_evento_tipo_publico?query=CalendarioEventoId__Id:"+calendarioResponsableID, &responsables)

									if errresponsable == nil {
										if responsables != nil {
											responsableList = nil
											for _, listRresponsable := range responsables {
												var responsablesID map[string]interface{} = listRresponsable["TipoPublicoId"].(map[string]interface{})
												// responsableID := fmt.Sprintf(responsablesID["Nombre"].(string))
												// responsableString = responsableID + ", " + responsableString

												responsableTipoP = map[string]interface{}{
													"responsableID": responsablesID["Id"].(float64),
													"Nombre":        fmt.Sprintf("%s", responsablesID["Nombre"].(string)),
												}
												responsableList = append(responsableList, responsableTipoP)
											}
										}
									} else {
										logs.Error(errresponsable.Error())
									}
								}

								actividad = nil
								actividad = map[string]interface{}{
									"actividadId":   proceso["Id"].(float64),
									"Nombre":        proceso["Nombre"].(string),
									"Descripcion":   proceso["Descripcion"].(string),
									"FechaInicio":   proceso["FechaInicio"].(string),
									"FechaFin":      proceso["FechaFin"].(string),
									"Activo":        proceso["Activo"].(bool),
									"TipoEventoId":  proceso["TipoEventoId"].(map[string]interface{}),
									"EventoPadreId": proceso["EventoPadreId"],
									"Responsable":   responsableList,
									"DependenciaId": proceso["DependenciaId"].(string),
								}
								actividadResultado = append(actividadResultado, actividad)

							}

							procesoAdd = nil
							procesoAdd = map[string]interface{}{
								"Proceso":     procesos[0]["TipoEventoId"].(map[string]interface{})["Nombre"].(string),
								"Actividades": actividadResultado,
							}

							procesoResultado = append(procesoResultado, procesoAdd)
							actividadResultado = nil

						} else {
							return requestresponse.APIResponseDTO(true, 200, procesos), nil
						}

					} else {
						logs.Error(errproceso.Error())
					}
				}
				calendarioAux := calendarios[0]["TipoEventoId"].(map[string]interface{})["CalendarioID"].(map[string]interface{})

				var calendariosExt []map[string]interface{}
				errcalendariosExt := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario?query=AplicaExtension:true,CalendarioPadreId.Id:"+idCalendario+"&limit=0", &calendariosExt)
				if errcalendariosExt == nil {
					fmt.Println("list: ", calendariosExt)
					if calendariosExt != nil && fmt.Sprintf("%v", calendariosExt) != "[map[]]" {
						calendariosExtlist = nil
						for _, calExt := range calendariosExt {
							Ext := map[string]interface{}{
								"Id":     calExt["Id"].(float64),
								"Nombre": calExt["Nombre"].(string),
								"Activo": calExt["Activo"].(bool),
							}
							calendariosExtlist = append(calendariosExtlist, Ext)
						}
					}
				} else {
					fmt.Println("error calen ext list", errcalendariosExt)
				}

				var ExisteExtension = false
				if calendariosExtlist != nil {
					ExisteExtension = true
				}

				resultado = map[string]interface{}{
					"Id":                      idCalendario,
					"Nombre":                  calendarioAux["Nombre"].(string),
					"PeriodoId":               calendarioAux["PeriodoId"].(float64),
					"MultiplePeriodoId":       calendarioAux["MultiplePeriodoId"].(string),
					"Activo":                  calendarioAux["Activo"].(bool),
					"Nivel":                   calendarioAux["Nivel"].(float64),
					"ListaCalendario":         versionCalendarioResultado,
					"resolucion":              resolucion,
					"DependenciaId":           calendarioAux["DependenciaId"].(string),
					"proceso":                 procesoResultado,
					"ExistenExtensiones":      ExisteExtension,
					"ListaExtension":          calendariosExtlist,
					"resolucionExt":           resolucionExt,
					"AplicaExtension":         calendarioAux["AplicaExtension"].(bool),
					"DependenciaParticularId": calendarioAux["DependenciaParticularId"].(string),
				}
				resultados = append(resultados, resultado)

				return requestresponse.APIResponseDTO(true, 200, resultados), nil

			} else {
				///////////////////////// sin eventos //////////////////////
				var calendario map[string]interface{}
				errcalendario := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario/"+idCalendario, &calendario)
				if errcalendario == nil {
					if calendario["Id"] != nil {

						if calendario["CalendarioPadreId"] != nil {
							padreID := fmt.Sprintf("%.f", calendario["CalendarioPadreId"].(map[string]interface{})["Id"].(float64))
							versionCalendario = map[string]interface{}{
								"Id":     padreID,
								"Nombre": calendario["CalendarioPadreId"].(map[string]interface{})["Nombre"],
							}
							versionCalendarioResultado = append(versionCalendarioResultado, versionCalendario)
						}

						documentoID := fmt.Sprintf("%.f", calendario["DocumentoId"].(float64))
						var documentos map[string]interface{}
						errdocumento := request.GetJson("http://"+beego.AppConfig.String("DocumentosService")+"documento/"+documentoID, &documentos)

						if errdocumento == nil {

							if documentos != nil {

								metadatoJSON := documentos["Metadatos"].(string)
								var metadato models.Metadatos
								json.Unmarshal([]byte(metadatoJSON), &metadato)

								resolucion = map[string]interface{}{
									"Id":         documentos["Id"],
									"Enlace":     documentos["Enlace"],
									"Resolucion": metadato.Resolucion,
									"Anno":       metadato.Anno,
									"Nombre":     documentos["Nombre"],
								}
							} else {
								return requestresponse.APIResponseDTO(true, 200, documentos), nil
							}

						} else {
							logs.Error(errdocumento.Error())
						}

						documentoExtID, ok := calendario["DocumentoExtensionId"].(float64)

						if documentoExtID != 0 && ok {
							var documentosExt map[string]interface{}
							errdocumentoExt := request.GetJson("http://"+beego.AppConfig.String("DocumentosService")+"documento/"+fmt.Sprintf("%.f", documentoExtID), &documentosExt)

							if errdocumentoExt == nil {
								if documentosExt != nil {
									metadatoJSON := documentosExt["Metadatos"].(string)
									var metadato models.Metadatos
									json.Unmarshal([]byte(metadatoJSON), &metadato)

									resolucionExt = map[string]interface{}{
										"Id":         documentosExt["Id"],
										"Enlace":     documentosExt["Enlace"],
										"Resolucion": metadato.Resolucion,
										"Anno":       metadato.Anno,
										"Nombre":     documentosExt["Nombre"],
									}
								} else {
									return requestresponse.APIResponseDTO(true, 200, documentosExt), nil
								}

							} else {
								logs.Error(errdocumentoExt.Error())
							}
						}

						var calendariosExt []map[string]interface{}
						errcalendariosExt := request.GetJson("http://"+beego.AppConfig.String("EventoService")+"calendario?query=AplicaExtension:true,CalendarioPadreId.Id:"+idCalendario+"&limit=0", &calendariosExt)
						if errcalendariosExt == nil {
							fmt.Println("list: ", calendariosExt)
							if calendariosExt != nil && fmt.Sprintf("%v", calendariosExt) != "[map[]]" {
								calendariosExtlist = nil
								for _, calExt := range calendariosExt {
									Ext := map[string]interface{}{
										"Id":     calExt["Id"].(float64),
										"Nombre": calExt["Nombre"].(string),
										"Activo": calExt["Activo"].(bool),
									}
									calendariosExtlist = append(calendariosExtlist, Ext)
								}
							}
						} else {
							fmt.Println("error calen ext list", errcalendariosExt)
						}

						var ExisteExtension = false
						if calendariosExtlist != nil {
							ExisteExtension = true
						}

						resultado = map[string]interface{}{
							"Id":                      idCalendario,
							"Nombre":                  calendario["Nombre"].(string),
							"PeriodoId":               calendario["PeriodoId"].(float64),
							"MultiplePeriodoId":       calendario["MultiplePeriodoId"].(string),
							"Activo":                  calendario["Activo"].(bool),
							"Nivel":                   calendario["Nivel"].(float64),
							"ListaCalendario":         versionCalendarioResultado,
							"resolucion":              resolucion,
							"DependenciaId":           calendario["DependenciaId"].(string),
							"proceso":                 procesoResultado,
							"ExistenExtensiones":      ExisteExtension,
							"ListaExtension":          calendariosExtlist,
							"resolucionExt":           resolucionExt,
							"AplicaExtension":         calendario["AplicaExtension"].(bool),
							"DependenciaParticularId": calendario["DependenciaParticularId"].(string),
						}
						resultados = append(resultados, resultado)

						return requestresponse.APIResponseDTO(true, 200, resultados), nil
					} else {
						return nil, errors.New("error del servicio GetCalendarInfo: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
					}

				} else {
					return requestresponse.APIResponseDTO(true, 200, calendarios), nil
				}

			}

		} else {
			return nil, errors.New("error del servicio GetCalendarInfo: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
		}

	} else {
		if resultado["Body"] == "<QuerySeter> no row found" {
			return nil, errors.New("error del servicio GetCalendarInfo: <QuerySeter> no row found")
		} else {
			return nil, errors.New("error del servicio GetCalendarInfo: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
		}
	}
}
