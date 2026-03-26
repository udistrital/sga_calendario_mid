package services

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
)

func PostCalendario(data []byte) (interface{}, error) {
	var calendario map[string]interface{}
	var calendarioParam []map[string]interface{}
	var tipoEvento []map[string]interface{}
	var calendarioEvento []map[string]interface{}
	var calendarioEventoTipoPublico []map[string]interface{}
	var tipoPublico map[string]interface{}
	var resultadoPost map[string]interface{}
	var resultadoPostResponsable map[string]interface{}
	var errCalendarioParam = errors.New("")
	var errorGetAll bool
	var message string

	var dataPost map[string]interface{}
	if err := json.Unmarshal(data, &dataPost); err == nil {

		idCalendario := fmt.Sprintf("%.f", dataPost["Id"].(float64))
		idPeriodo := fmt.Sprintf("%.f", dataPost["PeriodoIdClone"].(float64))
		idNivel := fmt.Sprintf("%.f", dataPost["NivelClone"].(float64))

		errCalendario := request.GetJson(beego.AppConfig.String("EventoService")+"calendario/"+idCalendario, &calendario)
		if errCalendario == nil {
			if calendario != nil {

				// if dataPost["NivelClone"].(float64) == calendario["Nivel"].(float64) {
				// 	errCalendarioParam = request.GetJson(beego.AppConfig.String("EventoService")+"calendario?query=Activo:true,PeriodoId:"+idPeriodo+",Nivel:"+idNivel+"&sortby=Id&order=desc&offset=1&limit=0", &calendarioParam)
				// } else {
				// 	errCalendarioParam = request.GetJson(beego.AppConfig.String("EventoService")+"calendario?query=Activo:true,PeriodoId:"+idPeriodo+",Nivel:"+idNivel+"&sortby=Id&order=desc", &calendarioParam)
				// }

				errCalendarioParam = request.GetJson(beego.AppConfig.String("EventoService")+"calendario?query=Activo:true,PeriodoId:"+idPeriodo+",Nivel:"+idNivel+"&sortby=Id&order=desc", &calendarioParam)

				if errCalendarioParam == nil {
					if calendarioParam != nil && calendarioParam[0]["Id"] != nil {

						idCalendarioParam := fmt.Sprintf("%.f", calendarioParam[0]["Id"].(float64))

						// persistir tipo_evento si el calendario que se esta clonando los tiene
						errTipoEvento := request.GetJson(beego.AppConfig.String("EventoService")+"tipo_evento?query=CalendarioID__Id:"+idCalendarioParam, &tipoEvento)
						if errTipoEvento == nil {
							if tipoEvento != nil && tipoEvento[0]["Id"] != nil {
								for _, tEvento := range tipoEvento {

									idOld := fmt.Sprintf("%.f", tEvento["Id"].(float64))
									tEvento["Id"] = 0
									tEvento["CalendarioID"] = calendario

									errTipoEventoPost := request.SendJson(beego.AppConfig.String("EventoService")+"/tipo_evento", "POST", &resultadoPost, tEvento)
									if errTipoEventoPost == nil && fmt.Sprintf("%v", resultadoPost["System"]) != "map[]" && resultadoPost["Id"] != nil {
										if resultadoPost["Status"] != 400 {
											tEvento["Id"] = resultadoPost["Id"]

											// presistir calendario_evento si el tipo_evento que se esta clonando esta asociado en el campo tipo_evento_id del calendario_evento
											errCalendarioEvento := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento?query=TipoEventoId__Id:"+idOld, &calendarioEvento)
											if errCalendarioEvento == nil {
												if calendarioEvento != nil && calendarioEvento[0]["Id"] != nil {
													idCalendarioEventoOld := fmt.Sprintf("%.f", calendarioEvento[0]["Id"].(float64))
													for _, cEvento := range calendarioEvento {

														cEvento["Id"] = 0
														cEvento["TipoEventoId"] = tEvento
														cEvento["FechaInicio"] = "2020-01-01T00:00:00-05:00"
														cEvento["FechaFin"] = "2020-01-01T00:00:00-05:00"

														errCalendarioEventoPost := request.SendJson(beego.AppConfig.String("EventoService")+"/calendario_evento", "POST", &resultadoPost, cEvento)
														if errCalendarioEventoPost == nil && fmt.Sprintf("%v", resultadoPost["System"]) != "map[]" && resultadoPost["Id"] != nil {
															if resultadoPost["Status"] != 400 {

																//validar si existe relcion de responsables, tabla rompimiento calendario_evento_tipo_publico y tipo_publico
																errCalendarioEventoTipoPublico := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento_tipo_publico?query=CalendarioEventoId__Id:"+idCalendarioEventoOld, &calendarioEventoTipoPublico)
																if errCalendarioEventoTipoPublico == nil && fmt.Sprintf("%v", resultadoPost["System"]) != "map[]" && resultadoPost["Id"] != nil {
																	if resultadoPost["Status"] == nil {
																		for _, cEventoTipoPublico := range calendarioEventoTipoPublico {
																			tipoPublicoOld := fmt.Sprintf("%.f", cEventoTipoPublico["TipoPublicoId"].(map[string]interface{})["Id"].(float64))
																			errTipoPublico := request.GetJson(beego.AppConfig.String("EventoService")+"tipo_publico/"+tipoPublicoOld, &tipoPublico)
																			if errTipoPublico == nil && fmt.Sprintf("%v", resultadoPost["System"]) != "map[]" && resultadoPost["Id"] != nil {
																				if resultadoPost["Status"] == nil {

																					cEventoTipoPublico["Id"] = 0
																					cEventoTipoPublico["CalendarioEventoId"] = resultadoPost
																					cEventoTipoPublico["TipoPublicoId"] = tipoPublico

																					errCalendarioEventoPost := request.SendJson(beego.AppConfig.String("EventoService")+"/calendario_evento_tipo_publico", "POST", &resultadoPostResponsable, cEventoTipoPublico)
																					if errCalendarioEventoPost == nil && fmt.Sprintf("%v", resultadoPostResponsable["System"]) != "map[]" && resultadoPostResponsable["Id"] != nil {
																						if resultadoPost["Status"] != 400 {
																							// fmt.Println("calendario_evento nuevo: ", resultadoPostResponsable["Id"])
																						}
																					}

																				}
																			} else {
																				errorGetAll = true
																				message += errTipoPublico.Error()
																			}
																		}
																	}
																} else {
																	errorGetAll = true
																	message += errCalendarioEventoTipoPublico.Error()
																}

															}
														} else {
															errorGetAll = true
															message += errCalendarioEventoPost.Error()
														}

													}
												}
											} else {
												errorGetAll = true
												message += errCalendarioEvento.Error()
											}
										}
									} else {
										errorGetAll = true
										message += errTipoEventoPost.Error()
									}

								}

							} else {
								errorGetAll = true
								message += "No data found tipoEvento[0]"
							}
						} else {
							errorGetAll = true
							message += errTipoEvento.Error()
						}
					} else {
						errorGetAll = true
						message += "No data found calendarioParam[0]"
					}
				} else {
					errorGetAll = true
					message += errCalendarioParam.Error()
				}
			} else {
				errorGetAll = true
				message += "No data found"
			}
		} else {
			errorGetAll = true
			message += errCalendarioParam.Error()
		}

	}
	if !errorGetAll {
		return calendario, nil
	} else {
		return nil, errors.New("error del servicio PostCalendario: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
	}
}

func PostCalendarioPadre(data []byte) (interface{}, error) {
	var calendario map[string]interface{}
	var calendarioParam []map[string]interface{}
	var tipoEvento []map[string]interface{}
	var calendarioEvento []map[string]interface{}
	var resultadoPost map[string]interface{}
	var resultado map[string]interface{}
	var errCalendarioParam = errors.New("")
	var errorGetAll bool

	var dataPost map[string]interface{}
	if err := json.Unmarshal(data, &dataPost); err == nil {
		idCalendario := fmt.Sprintf("%.f", dataPost["Id"].(float64))
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
					if calendarioParam != nil && calendarioParam[0]["Id"] != nil {
						idCalendarioParam := fmt.Sprintf("%.f", calendarioParam[0]["Id"].(float64))

						// persistir tipo_evento si el calendario que se esta clonando los tiene
						errTipoEvento := request.GetJson(beego.AppConfig.String("EventoService")+"tipo_evento?query=CalendarioID__Id:"+idCalendarioParam, &tipoEvento)
						if errTipoEvento == nil {
							if tipoEvento != nil && tipoEvento[0]["Id"] != nil {
								for _, tEvento := range tipoEvento {
									idOld := fmt.Sprintf("%.f", tEvento["Id"].(float64))
									tEvento["Id"] = 0
									tEvento["CalendarioID"] = calendario

									errTipoEventoPost := request.SendJson(beego.AppConfig.String("EventoService")+"/tipo_evento", "POST", &resultadoPost, tEvento)
									if errTipoEventoPost == nil && fmt.Sprintf("%v", resultadoPost["System"]) != "map[]" && resultadoPost["Id"] != nil {
										if resultadoPost["Status"] != 400 {
											tEvento["Id"] = resultadoPost["Id"]

											// presistir calendario_evento si el tipo_evento que se esta clonando esta asociado en el campo tipo_evento_id del calendario_evento
											errCalendarioEvento := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento?query=TipoEventoId__Id:"+idOld, &calendarioEvento)
											if errCalendarioEvento == nil {
												if calendarioEvento != nil && calendarioEvento[0]["Id"] != nil {
													for _, cEvento := range calendarioEvento {
														cEvento["Id"] = 0
														cEvento["TipoEventoId"] = tEvento
														cEvento["FechaInicio"] = "2000-01-01T00:00:00-05:00"
														cEvento["FechaFin"] = "2000-01-01T00:00:00-05:00"

														errCalendarioEventoPost := request.SendJson(beego.AppConfig.String("EventoService")+"/calendario_evento", "POST", &resultadoPost, cEvento)
														if errCalendarioEventoPost == nil && fmt.Sprintf("%v", resultadoPost["System"]) != "map[]" && resultadoPost["Id"] != nil {
															if resultadoPost["Status"] != 400 {
																fmt.Println("calendario_evento nuevo: ", resultadoPost["Id"])
															} else {
																errorGetAll = true
																logs.Error(errCalendarioEventoPost.Error())
															}
														} else {
															errorGetAll = true
														}
													}
												} else {
													errorGetAll = true
													logs.Error(errCalendarioEvento.Error())
												}
											} else {
												errorGetAll = true
											}
										} else {
											errorGetAll = true
											logs.Error(errTipoEventoPost.Error())
										}
									} else {
										errorGetAll = true
									}

								}
								resultado = map[string]interface{}{
									"Id": idCalendario,
								}
							} else {
								errorGetAll = true
							}
						} else {
							errorGetAll = true
							logs.Error(errTipoEvento.Error())
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
		return resultado, nil
	} else {
		return nil, errors.New("error del servicio PostCalendarioPadre: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
	}
}

func PostCalendarioExtension(data []byte) (interface{}, error) {
	var CalendarioExtension map[string]interface{}
	var CalendarioPadre map[string]interface{}
	var CalendarioExtCrear map[string]interface{}

	var tipoEvento []map[string]interface{}
	var calendarioEvento []map[string]interface{}
	var calendarioEventoTipoPublico []map[string]interface{}
	var tipoPublico map[string]interface{}
	var resultadoPost map[string]interface{}
	var resultadoPostResponsable map[string]interface{}

	if err := json.Unmarshal(data, &CalendarioExtension); err == nil {
		IdPadre := fmt.Sprintf("%.f", CalendarioExtension["CalendarioPadre"])

		errCalendarioPadre := request.GetJson(beego.AppConfig.String("EventoService")+"calendario/"+IdPadre, &CalendarioPadre)
		if errCalendarioPadre == nil {
			if _, ok := CalendarioPadre["Id"]; ok {
				CalendarioPadre["AplicaExtension"] = true
				CalendarioPadre["DependenciaParticularId"] = CalendarioExtension["Dependencias"]
				CalendarioPadre["DocumentoExtensionId"] = CalendarioExtension["DocumentoExtensionId"]
				CalendarioPadre["CalendarioPadreId"] = map[string]interface{}{"Id": CalendarioExtension["CalendarioPadre"]}
				delete(CalendarioPadre, "Id")

				errCalendarioExtCrear := request.SendJson(beego.AppConfig.String("EventoService")+"calendario", "POST", &CalendarioExtCrear, CalendarioPadre)
				if errCalendarioExtCrear == nil {

					// persistir tipo_evento si el calendario que se esta clonando los tiene
					errTipoEvento := request.GetJson(beego.AppConfig.String("EventoService")+"tipo_evento?query=CalendarioID__Id:"+IdPadre+",Activo:true&limit=0", &tipoEvento)
					if errTipoEvento == nil {
						if tipoEvento != nil && tipoEvento[0]["Id"] != nil {
							for _, tEvento := range tipoEvento {

								idOld := fmt.Sprintf("%.f", tEvento["Id"].(float64))
								tEvento["Id"] = 0
								tEvento["CalendarioID"] = CalendarioExtCrear

								errTipoEventoPost := request.SendJson(beego.AppConfig.String("EventoService")+"/tipo_evento", "POST", &resultadoPost, tEvento)
								if errTipoEventoPost == nil && fmt.Sprintf("%v", resultadoPost["System"]) != "map[]" && resultadoPost["Id"] != nil {
									if resultadoPost["Status"] != 400 {
										tEvento["Id"] = resultadoPost["Id"]

										// presistir calendario_evento si el tipo_evento que se esta clonando esta asociado en el campo tipo_evento_id del calendario_evento
										errCalendarioEvento := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento?query=TipoEventoId__Id:"+idOld+",Activo:true&limit=0", &calendarioEvento)
										if errCalendarioEvento == nil {
											if calendarioEvento != nil && calendarioEvento[0]["Id"] != nil {
												idCalendarioEventoOld := fmt.Sprintf("%v", calendarioEvento[0]["Id"])
												for _, cEvento := range calendarioEvento {

													cEvento["Id"] = 0
													cEvento["TipoEventoId"] = tEvento

													errCalendarioEventoPost := request.SendJson(beego.AppConfig.String("EventoService")+"/calendario_evento", "POST", &resultadoPost, cEvento)
													if errCalendarioEventoPost == nil && fmt.Sprintf("%v", resultadoPost["System"]) != "map[]" && resultadoPost["Id"] != nil {
														if resultadoPost["Status"] != 400 {

															//validar si existe relcion de responsables, tabla rompimiento calendario_evento_tipo_publico y tipo_publico
															errCalendarioEventoTipoPublico := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento_tipo_publico?query=CalendarioEventoId__Id:"+idCalendarioEventoOld+",Activo:true&limit=0", &calendarioEventoTipoPublico)
															if errCalendarioEventoTipoPublico == nil && fmt.Sprintf("%v", resultadoPost["System"]) != "map[]" && resultadoPost["Id"] != nil {
																if resultadoPost["Status"] == nil {
																	for _, cEventoTipoPublico := range calendarioEventoTipoPublico {
																		tipoPublicoOld := fmt.Sprintf("%.f", cEventoTipoPublico["TipoPublicoId"].(map[string]interface{})["Id"].(float64))
																		errTipoPublico := request.GetJson(beego.AppConfig.String("EventoService")+"tipo_publico/"+tipoPublicoOld, &tipoPublico)
																		if errTipoPublico == nil && fmt.Sprintf("%v", resultadoPost["System"]) != "map[]" && resultadoPost["Id"] != nil {
																			if resultadoPost["Status"] == nil {

																				cEventoTipoPublico["Id"] = 0
																				cEventoTipoPublico["CalendarioEventoId"] = resultadoPost
																				cEventoTipoPublico["TipoPublicoId"] = tipoPublico

																				errCalendarioEventoPost := request.SendJson(beego.AppConfig.String("EventoService")+"/calendario_evento_tipo_publico", "POST", &resultadoPostResponsable, cEventoTipoPublico)
																				if errCalendarioEventoPost == nil && fmt.Sprintf("%v", resultadoPostResponsable["System"]) != "map[]" && resultadoPostResponsable["Id"] != nil {
																					if resultadoPost["Status"] != 400 {
																						// fmt.Println("calendario_evento nuevo: ", resultadoPostResponsable["Id"])
																					}
																				}

																			}
																		} else {
																			fmt.Println("error get tipo_publico", errTipoPublico)
																			logs.Error(errTipoPublico.Error())
																			return nil, errors.New("error del servicio PostCalendarioExtension: error get tipo_publico")
																		}
																	}
																	return requestresponse.APIResponseDTO(true, 200, nil), nil

																}
															} else {
																//errorGetAll = true
																fmt.Println("error get calend_evento_tipo_pub", errCalendarioEventoTipoPublico)
																logs.Error(errCalendarioEventoTipoPublico.Error())
																return nil, errors.New("error del servicio PostCalendarioExtension: error get calend_evento_tipo_pub")
															}
														}
													} else {
														fmt.Println("error post calend event", errCalendarioEventoPost)
														logs.Error(errCalendarioEventoPost.Error())
														return nil, errors.New("error del servicio PostCalendarioExtension: error post calend event")
													}

												}
											}
										} else {
											//errorGetAll = true
											fmt.Println("error get calend_evento", errCalendarioEvento)
											logs.Error(errCalendarioEvento.Error())
											return nil, errors.New("error del servicio PostCalendarioExtension: error get calend_evento")
										}
									}
								} else {
									//errorGetAll = true
									fmt.Println("error put evento", errTipoEventoPost)
									logs.Error(errTipoEventoPost.Error())
									return nil, errors.New("error del servicio PostCalendarioExtension: error put evento")
								}

							}

						} else {
							return requestresponse.APIResponseDTO(true, 200, nil, "ok pero no eventos"), nil
						}
					} else {
						fmt.Println("error get eventos", errTipoEvento)
						logs.Error(errTipoEvento.Error())
						return nil, errors.New("error del servicio PostCalendarioExtension: error get eventos")
					}

				} else {
					fmt.Println("error put calend padre extension", errCalendarioExtCrear)
					logs.Error(errCalendarioExtCrear.Error())
					return nil, errors.New("error del servicio PostCalendarioExtension: error put calend padre extension")
				}
			} else {
				fmt.Println("error get calend padre")
				return nil, errors.New("error del servicio PostCalendarioExtension: error get calend padre")
			}
		} else {
			fmt.Println("error get calend padre", errCalendarioPadre)
			return nil, errors.New("error del servicio PostCalendarioExtension: error get calend padre")
		}
	} else {
		fmt.Println("error body")
		return nil, errors.New("error del servicio PostCalendarioExtension: error body")
	}
	return nil, errors.New("error del servicio PostCalendarioExtension: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
}
