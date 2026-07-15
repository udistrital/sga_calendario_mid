package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
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
	var mutex sync.Mutex
	wge := new(errgroup.Group)

	errCalendario := request.GetJson(beego.AppConfig.String("EventoService")+"calendario?limit=0&sortby=Id&order=desc", &calendarios)
	if errCalendario == nil {
		if len(calendarios) > 0 && len(calendarios[0]) > 0 && fmt.Sprintf("%v", calendarios[0]["Nombre"]) != "map[]" {
			fmt.Println(len(calendarios))
			//Limitación de la cantidad de hilos a utilizar, valores negativas representan sin limite
			wge.SetLimit(10)
			for _, calendario := range calendarios {

				calendario := calendario
				//Declaración función anonima
				wge.Go(func() error {
					var periodo map[string]interface{}
					var errPeriodo error

					periodoID := fmt.Sprintf("%.f", calendario["PeriodoId"].(float64))
					errPeriodo = request.GetJson(beego.AppConfig.String("ParametroService")+"periodo/"+periodoID, &periodo)
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
						mutex.Lock()
						resultados = append(resultados, resultado)
						mutex.Unlock()
					} else {
						return errPeriodo
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
		sort.SliceStable(resultados, func(i, j int) bool {
			periodoI, _ := resultados[i]["Periodo"].(string)
			periodoJ, _ := resultados[j]["Periodo"].(string)
			if periodoI == periodoJ {
				nombreI, _ := resultados[i]["Nombre"].(string)
				nombreJ, _ := resultados[j]["Nombre"].(string)
				return nombreI < nombreJ
			}
			return periodoI > periodoJ
		})
		return requestresponse.APIResponseDTO(true, 200, resultados), nil
	} else {
		return nil, errors.New("error del servicio GetAll: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
	}
}

func GetOnePorId(idCalendario string) (interface{}, error) {
	var resultado map[string]interface{}
	var resultados []map[string]interface{}
	var versionCalendarioResultado []map[string]interface{}
	var documento map[string]interface{}
	var resolucion map[string]interface{}
	var procesoArr []string
	var proceso map[string]interface{}
	var procesoResultado []map[string]interface{}
	var actividad map[string]interface{}
	var procesoAdd map[string]interface{}

	if resultado["Type"] != "error" {
		// consultar calendario evento por tipo evento
		var calendarios []map[string]interface{}
		errcalendario := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento?query=ProcesoId__Id.CalendarioID__Id:"+idCalendario, &calendarios)

		if errcalendario == nil {
			if calendarios[0]["Id"] != nil {

				documento = calendarios[0]["ProcesoId"].(map[string]interface{})["CalendarioID"].(map[string]interface{})
				documentoID := fmt.Sprintf("%.f", documento["DocumentoId"].(float64))
				var documentos map[string]interface{}
				errdocumento := request.GetJson(beego.AppConfig.String("DocumentosService")+"documento/"+documentoID, &documentos)

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
						"NombreProceso": calendario["ProcesoId"].(map[string]interface{})["Id"].(float64),
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
						errproceso := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento?query=ProcesoId.Id:"+procesoList+"&ProcesoId__Id.CalendarioID__Id:"+idCalendario, &procesos)

						if errproceso == nil {
							if procesos != nil {
								for _, proceso := range procesos {

									responsableList := responsablesActividad(proceso)
									idActividad := fmt.Sprintf("%.f", proceso["Id"].(float64))

									actividad = nil
									nombreActividad, descripcionActividad := datosEventoCatalogo(proceso["EventoCatalogoId"])
									actividad = map[string]interface{}{
										"actividadId":      proceso["Id"].(float64),
										"Nombre":           nombreActividad,
										"Descripcion":      descripcionActividad,
										"FechaInicio":      proceso["FechaInicio"].(string),
										"FechaFin":         proceso["FechaFin"].(string),
										"Activo":           proceso["Activo"].(bool),
										"ProcesoId":        proceso["ProcesoId"].(map[string]interface{}),
										"EventoCatalogoId": proceso["EventoCatalogoId"],
										"Responsable":      responsableList,
										"Extensiones":      extensionResumenActividad(idActividad),
									}

									actividadResultado = append(actividadResultado, actividad)

								}

								nombreProceso, _, _ := datosProcesoCatalogo(procesos[0]["ProcesoId"])
								procesoAdd = nil
								procesoAdd = map[string]interface{}{
									"Proceso":     nombreProceso,
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

				calendarioAux := calendarios[0]["ProcesoId"].(map[string]interface{})["CalendarioID"].(map[string]interface{})
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
				errcalendario := request.GetJson(beego.AppConfig.String("EventoService")+"calendario/"+idCalendario, &calendario)
				if errcalendario == nil {
					if calendario["Id"] != nil {

						documentoID := fmt.Sprintf("%.f", calendario["DocumentoId"].(float64))
						var documentos map[string]interface{}

						errdocumento := request.GetJson(beego.AppConfig.String("DocumentosService")+"documento/"+documentoID, &documentos)

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

func PutInhabilitarCalendario(idCalendario string, data []byte, usuario string) (interface{}, error) {
	var calendario map[string]interface{}
	var procesos []map[string]interface{}
	var calendarioEvento []map[string]interface{}
	var resultado map[string]interface{}
	var dataPut map[string]interface{}
	var message string
	var success bool = true
	alertas := []interface{}{"Response:"}
	if err := json.Unmarshal(data, &dataPut); err == nil {

		errCalendario := request.GetJson(beego.AppConfig.String("EventoService")+"calendario/"+idCalendario, &calendario)
		if errCalendario == nil {
			if calendario != nil {

				calendarioAnterior := deepCopyMap(calendario)
				calendario["Activo"] = false

				errCalendario := request.SendJson(beego.AppConfig.String("EventoService")+"calendario/"+idCalendario, "PUT", &resultado, calendario)
				if resultado["Type"] == "error" || errCalendario != nil || resultado["Status"] == "404" || resultado["Message"] != nil {
					success = false
				} else {
					if id, err := strconv.Atoi(idCalendario); err == nil {
						RegistrarAuditoria("calendario", id, "PUT", calendarioAnterior, resultado, usuario, "PutInhabilitarCalendario/calendario")
					}

					errCalendario := request.GetJson(beego.AppConfig.String("EventoService")+"proceso?query=Activo:true,CalendarioID__Id:"+idCalendario, &procesos)
					if errCalendario == nil {
						if len(procesos) > 0 && procesos[0] != nil && len(procesos[0]) > 0 {

							for _, proceso := range procesos {

								idProceso := fmt.Sprintf("%.f", proceso["Id"].(float64))

								procesoAnterior := deepCopyMap(proceso)
								proceso["Activo"] = false

								errCalendario := request.SendJson(beego.AppConfig.String("EventoService")+"proceso/"+idProceso, "PUT", &resultado, proceso)
								if resultado["Type"] == "error" || errCalendario != nil || resultado["Status"] == "404" || resultado["Message"] != nil {
									success = false
								} else {
									if id, err := strconv.Atoi(idProceso); err == nil {
										RegistrarAuditoria("proceso", id, "PUT", procesoAnterior, resultado, usuario, "PutInhabilitarCalendario/proceso")
									}

									errCalendario := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento?query=Activo:true,ProcesoId__Id:"+idProceso, &calendarioEvento)
									if errCalendario == nil {
										if len(calendarioEvento) > 0 && calendarioEvento[0] != nil && len(calendarioEvento[0]) > 0 {

											for _, cEvento := range calendarioEvento {

												idCalendarioEvento := fmt.Sprintf("%.f", cEvento["Id"].(float64))

												cEventoAnterior := deepCopyMap(cEvento)
												cEvento["Activo"] = false

												errCalendario := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento/"+idCalendarioEvento, "PUT", &resultado, cEvento)
												if resultado["Type"] == "error" || errCalendario != nil || resultado["Status"] == "404" || resultado["Message"] != nil {
													success = false
												} else {
													if id, err := strconv.Atoi(idCalendarioEvento); err == nil {
														RegistrarAuditoria("calendario_evento", id, "PUT", cEventoAnterior, resultado, usuario, "PutInhabilitarCalendario/calendario_evento")
													}
													if err := inactivarExtensionesActividad(idCalendarioEvento, usuario, "PutInhabilitarCalendario"); err != nil {
														logs.Error(err)
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

func PostCalendarioHijo(data []byte, usuario string) (interface{}, error) {
	var AuxCalendarioHijo map[string]interface{}
	var calendarioHijoPost map[string]interface{}

	if err := json.Unmarshal(data, &AuxCalendarioHijo); err == nil {
		periodoID := fmt.Sprintf("%.f", AuxCalendarioHijo["PeriodoId"].(float64))
		nivelID := fmt.Sprintf("%.f", AuxCalendarioHijo["Nivel"].(float64))
		calendarioExistente, err := calendarioActivoPorPeriodoNivel(periodoID, nivelID)
		if err != nil {
			logs.Error(err)
			return nil, errors.New("error del servicio PostCalendarioHijo: no fue posible validar calendario existente")
		}
		if calendarioExistente {
			return nil, errors.New("Ya existe un calendario activo para el periodo y nivel seleccionados")
		}

		CalendarioHijo := map[string]interface{}{
			"Nombre":            AuxCalendarioHijo["Nombre"],
			"DependenciaId":     `{"proyectos":[]}`,
			"DocumentoId":       AuxCalendarioHijo["DocumentoId"],
			"PeriodoId":         AuxCalendarioHijo["PeriodoId"],
			"AplicacionId":      0,
			"Nivel":             AuxCalendarioHijo["Nivel"],
			"Activo":            AuxCalendarioHijo["Activo"],
			"FechaCreacion":     time_bogota.TiempoBogotaFormato(),
			"FechaModificacion": time_bogota.TiempoBogotaFormato(),
		}
		errCalendarioHijo := request.SendJson(beego.AppConfig.String("EventoService")+"calendario", "POST", &calendarioHijoPost, CalendarioHijo)

		if errCalendarioHijo == nil && fmt.Sprintf("%v", calendarioHijoPost["System"]) != "map[]" && calendarioHijoPost["Id"] != nil {
			if calendarioHijoPost["Status"] != 400 {
				if hijoId, ok := calendarioHijoPost["Id"].(float64); ok {
					RegistrarAuditoria("calendario", int(hijoId), "POST", nil, calendarioHijoPost, usuario, "PostCalendarioIndependiente/calendario")
				}
				return requestresponse.APIResponseDTO(true, 200, calendarioHijoPost), nil
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

func calendarioActivoPorPeriodoNivel(periodoID string, nivelID string) (bool, error) {
	var calendarios []map[string]interface{}
	url := beego.AppConfig.String("EventoService") + "calendario?query=Activo:true,PeriodoId:" + periodoID + ",Nivel:" + nivelID + "&limit=1"
	if err := request.GetJson(url, &calendarios); err != nil {
		return false, err
	}
	return len(calendarios) > 0 && len(calendarios[0]) > 0 && calendarios[0]["Id"] != nil, nil
}

func datosEventoCatalogo(eventoCatalogo interface{}) (string, string) {
	catalogo, ok := eventoCatalogo.(map[string]interface{})
	if !ok || catalogo == nil {
		return "", ""
	}
	nombre, _ := catalogo["Nombre"].(string)
	descripcion, _ := catalogo["Descripcion"].(string)
	return nombre, descripcion
}

func datosEventoCatalogoCompleto(eventoCatalogo interface{}) (string, string, string) {
	catalogo, ok := eventoCatalogo.(map[string]interface{})
	if !ok || catalogo == nil {
		return "", "", ""
	}
	nombre, _ := catalogo["Nombre"].(string)
	descripcion, _ := catalogo["Descripcion"].(string)
	codigo, _ := catalogo["CodigoAbreviacion"].(string)
	return nombre, descripcion, codigo
}

func datosProcesoCatalogo(proceso interface{}) (string, string, string) {
	procesoMap, ok := proceso.(map[string]interface{})
	if !ok || procesoMap == nil {
		return "", "", ""
	}
	catalogo, ok := procesoMap["ProcesoCatalogoId"].(map[string]interface{})
	if !ok || catalogo == nil {
		return "", "", ""
	}
	nombre, _ := catalogo["Nombre"].(string)
	descripcion, _ := catalogo["Descripcion"].(string)
	codigo, _ := catalogo["CodigoAbreviacion"].(string)
	return nombre, descripcion, codigo
}

func responsablesActividad(actividad map[string]interface{}) []map[string]interface{} {
	idActividad, ok := idToString(actividad["Id"])
	if !ok {
		return []map[string]interface{}{}
	}

	var relaciones []map[string]interface{}
	err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento_tipo_publico?query=CalendarioEventoId__Id:"+idActividad+"&limit=0", &relaciones)
	if err != nil {
		logs.Error(err.Error())
		return []map[string]interface{}{}
	}

	responsables := make([]map[string]interface{}, 0, len(relaciones))
	for _, relacion := range relaciones {
		perfilID, ok := interfaceToInt(relacion["PerfilId"])
		if !ok || perfilID <= 0 {
			continue
		}
		nombre := nombrePerfilConfiguracion(perfilID)
		responsables = append(responsables, map[string]interface{}{
			"responsableID": perfilID,
			"Nombre":        nombre,
			"Activo":        relacion["Activo"] != false,
		})
	}
	return responsables
}

func nombrePerfilConfiguracion(perfilID int) string {
	var perfil map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("ConfiguracionService")+"perfil/"+strconv.Itoa(perfilID), &perfil); err != nil || perfil == nil {
		return fmt.Sprintf("Perfil %d", perfilID)
	}
	if nombre, ok := perfil["Nombre"].(string); ok && nombre != "" {
		return nombre
	}
	if nombre, ok := perfil["nombre"].(string); ok && nombre != "" {
		return nombre
	}
	return fmt.Sprintf("Perfil %d", perfilID)
}

func GetCalendarInfo(idCalendario string) (interface{}, error) {
	var resultado map[string]interface{}
	var resultados []map[string]interface{}
	var actividadResultado []map[string]interface{}
	var versionCalendarioResultado []map[string]interface{}
	var documento map[string]interface{}
	var resolucion map[string]interface{}
	var procesoArr []string
	var proceso map[string]interface{}
	var procesoResultado []map[string]interface{}
	var actividad map[string]interface{}
	var procesoAdd map[string]interface{}

	//var resolucion_ext map[string]interface{}

	if resultado["Type"] != "error" {
		// consultar calendario evento por tipo evento
		var calendarios []map[string]interface{}
		errcalendario := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento?query=ProcesoId__Id.CalendarioID__Id:"+idCalendario, &calendarios)
		if errcalendario == nil {
			if len(calendarios) > 0 && calendarios[0]["Id"] != nil {

				documento = calendarios[0]["ProcesoId"].(map[string]interface{})["CalendarioID"].(map[string]interface{})
				documentoID := fmt.Sprintf("%.f", documento["DocumentoId"].(float64))

				var documentos map[string]interface{}
				errdocumento := request.GetJson(beego.AppConfig.String("DocumentosService")+"documento/"+documentoID, &documentos)

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
						"NombreProceso": calendario["ProcesoId"].(map[string]interface{})["Id"].(float64),
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
					errproceso := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento?query=ProcesoId.Id:"+procesoList+"&ProcesoId__Id.CalendarioID__Id:"+idCalendario, &procesos)

					if errproceso == nil {
						if procesos != nil {
							for _, proceso := range procesos {

								responsableList := responsablesActividad(proceso)
								idActividad := fmt.Sprintf("%.f", proceso["Id"].(float64))

								actividad = nil
								nombreActividad, descripcionActividad := datosEventoCatalogo(proceso["EventoCatalogoId"])
								actividad = map[string]interface{}{
									"actividadId":      proceso["Id"].(float64),
									"Nombre":           nombreActividad,
									"Descripcion":      descripcionActividad,
									"FechaInicio":      proceso["FechaInicio"].(string),
									"FechaFin":         proceso["FechaFin"].(string),
									"Activo":           proceso["Activo"].(bool),
									"ProcesoId":        proceso["ProcesoId"].(map[string]interface{}),
									"EventoCatalogoId": proceso["EventoCatalogoId"],
									"Responsable":      responsableList,
									"Extensiones":      extensionResumenActividad(idActividad),
									"DependenciaId":    proceso["DependenciaId"].(string),
								}
								actividadResultado = append(actividadResultado, actividad)

							}

							nombreProceso, _, _ := datosProcesoCatalogo(procesos[0]["ProcesoId"])
							procesoAdd = nil
							procesoAdd = map[string]interface{}{
								"Proceso":     nombreProceso,
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
				calendarioAux := calendarios[0]["ProcesoId"].(map[string]interface{})["CalendarioID"].(map[string]interface{})

				resultado = map[string]interface{}{
					"Id":                 idCalendario,
					"Nombre":             calendarioAux["Nombre"].(string),
					"PeriodoId":          calendarioAux["PeriodoId"].(float64),
					"Activo":             calendarioAux["Activo"].(bool),
					"Nivel":              calendarioAux["Nivel"].(float64),
					"ListaCalendario":    versionCalendarioResultado,
					"resolucion":         resolucion,
					"DependenciaId":      calendarioAux["DependenciaId"].(string),
					"proceso":            procesoResultado,
					"ExistenExtensiones": false,
					"ListaExtension":     []map[string]interface{}{},
				}
				resultados = append(resultados, resultado)

				return requestresponse.APIResponseDTO(true, 200, resultados), nil

			} else {
				///////////////////////// sin eventos //////////////////////
				var calendario map[string]interface{}
				errcalendario := request.GetJson(beego.AppConfig.String("EventoService")+"calendario/"+idCalendario, &calendario)
				if errcalendario == nil {
					if calendario["Id"] != nil {

						documentoID := fmt.Sprintf("%.f", calendario["DocumentoId"].(float64))
						var documentos map[string]interface{}
						errdocumento := request.GetJson(beego.AppConfig.String("DocumentosService")+"documento/"+documentoID, &documentos)

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

						dependenciaId, _ := calendario["DependenciaId"].(string)

						resultado = map[string]interface{}{
							"Id":                 idCalendario,
							"Nombre":             calendario["Nombre"].(string),
							"PeriodoId":          calendario["PeriodoId"].(float64),
							"Activo":             calendario["Activo"].(bool),
							"Nivel":              calendario["Nivel"].(float64),
							"ListaCalendario":    versionCalendarioResultado,
							"resolucion":         resolucion,
							"DependenciaId":      dependenciaId,
							"proceso":            procesoResultado,
							"ExistenExtensiones": false,
							"ListaExtension":     []map[string]interface{}{},
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
