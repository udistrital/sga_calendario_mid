package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/sga_calendario_mid/helpers"
	"github.com/udistrital/sga_calendario_mid/models"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
)

func PostActividadCalendario(data []byte) (interface{}, error) {
	//Almacena el json que se trae desde el cliente
	var actividadCalendario map[string]interface{}
	//Almacena el resultado del json en algunas operaciones
	var actividadCalendarioPost map[string]interface{}
	var IdActividad interface{}
	var actividadPersonaPost map[string]interface{}

	if err := json.Unmarshal(data, &actividadCalendario); err == nil {
		terceroID, err := helpers.TerceroIDFromPayload(actividadCalendario)
		if err != nil {
			return nil, errors.New("error del servicio PostActividadCalendario: " + err.Error())
		}
		Actividad := actividadCalendario["Actividad"]
		actividadMap, ok := Actividad.(map[string]interface{})
		if !ok {
			return nil, errors.New("error del servicio PostActividadCalendario: actividad inválida")
		}
		if err := validarFechasPayload(actividadMap, nil, "actividad"); err != nil {
			return nil, err
		}
		repetible, err := validarCatalogoActividadProceso(actividadMap)
		if err != nil {
			return nil, err
		}
		helpers.SetTerceroID(actividadMap, terceroID)
		helpers.NormalizeFechasTimeCalendarioEvento("calendario_evento", actividadMap)
		actividadCalendarioPost, err = crearCalendarioEventoConPolitica(actividadMap, repetible)
		if err != nil {
			return nil, errors.New("error del servicio PostActividadCalendario: " + err.Error())
		}
		IdActividad = actividadCalendarioPost["Id"]

		var totalPublico []interface{}
		//Guarda el JSON de la tabla tipo publico
		totalPublico = actividadCalendario["responsable"].([]interface{})

		for _, publicoTemp := range totalPublico {
			publicoMap := publicoTemp.(map[string]interface{})
			perfilID, ok := helpers.InterfaceToInt(publicoMap["responsableID"])
			if !ok || perfilID <= 0 {
				return nil, errors.New("error del servicio PostActividadCalendario: perfil de público dirigido inválido")
			}
			idActividadInt, _ := helpers.InterfaceToInt(IdActividad)
			CalendarioEventoTipoPersona := models.CalendarioEventoTipoPublicoPayload{
				Activo:             activoRelacionPublico(publicoMap),
				PerfilId:           perfilID,
				CalendarioEventoId: models.RelacionID{Id: idActividadInt},
			}

			errActividadPersona := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento_tipo_publico", "POST", &actividadPersonaPost, CalendarioEventoTipoPersona)

			if errActividadPersona == nil && fmt.Sprintf("%v", actividadPersonaPost["System"]) != "map[]" && actividadPersonaPost["Id"] != nil {
				if actividadPersonaPost["Status"] != 400 {
					return requestresponse.APIResponseDTO(true, 200, actividadCalendarioPost), nil
				} else {
					var resultado2 map[string]interface{}
					request.SendJson(fmt.Sprintf(beego.AppConfig.String("EventoService")+"/calendario_evento/%.f", actividadCalendarioPost["Id"]), "PUT", &resultado2, map[string]interface{}{"Activo": false, "TerceroId": terceroID})
					logs.Error(errActividadPersona)
				}
			} else {
				logs.Error(errActividadPersona)
			}
		}
	}
	return nil, errors.New("error del servicio PostActividadCalendario: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
}

func validarCatalogoActividadProceso(actividad map[string]interface{}) (bool, error) {
	procesoId, err := idRelacion(actividad["ProcesoId"])
	if err != nil {
		return false, errors.New("error del servicio PostActividadCalendario: proceso inválido")
	}
	eventoCatalogoId, err := idRelacion(actividad["EventoCatalogoId"])
	if err != nil {
		return false, errors.New("error del servicio PostActividadCalendario: evento de catálogo inválido")
	}

	var proceso map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"proceso/"+procesoId, &proceso); err != nil || proceso == nil || proceso["Type"] == "error" {
		return false, errors.New("error del servicio PostActividadCalendario: no fue posible consultar el proceso")
	}
	if activo, ok := proceso["Activo"].(bool); !ok || !activo {
		return false, errors.New("error del servicio PostActividadCalendario: el proceso no está activo")
	}

	procesoCatalogoId, err := idRelacion(proceso["ProcesoCatalogoId"])
	if err != nil {
		return false, errors.New("error del servicio PostActividadCalendario: proceso sin catálogo asociado")
	}

	var relaciones []map[string]interface{}
	url := beego.AppConfig.String("EventoService") + "evento_catalogo_proceso_catalogo?query=Activo:true,EventoCatalogoId__Id:" + eventoCatalogoId + ",ProcesoCatalogoId__Id:" + procesoCatalogoId + "&limit=1"
	if err := request.GetJson(url, &relaciones); err != nil || len(relaciones) == 0 || len(relaciones[0]) == 0 {
		return false, errors.New("error del servicio PostActividadCalendario: la actividad seleccionada no pertenece al proceso")
	}

	repetible, _ := relaciones[0]["Repetible"].(bool)
	return repetible, nil
}

func idRelacion(valor interface{}) (string, error) {
	return helpers.RelationIDToString(valor)
}

func UpdateActividadResponsables(idStr string, data []byte) (interface{}, error) {
	var recibido map[string]interface{}
	var guardados []map[string]interface{}
	var actualizados []map[string]interface{}
	var auxDelete string
	var auxUpdate map[string]interface{}
	var errBorrado error

	actividadId, _ := strconv.Atoi(idStr)
	if err := json.Unmarshal(data, &recibido); err == nil {
		terceroID, err := helpers.TerceroIDFromPayload(recibido)
		if err != nil {
			return nil, errors.New("error del servicio UpdateActividadResponsables: " + err.Error())
		}
		if actividad, ok := recibido["actividad"].(map[string]interface{}); ok {
			helpers.SetTerceroID(actividad, terceroID)
			if err := actualizarFechasActividad(idStr, actividad, terceroID); err != nil {
				return nil, err
			}
		}
		datos, tieneResponsables := recibido["resp"].([]interface{})
		if !tieneResponsables {
			return requestresponse.APIResponseDTO(true, 200, map[string]interface{}{"actividad": actividadId}), nil
		}
		errConsulta := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento_tipo_publico?query=CalendarioEventoId__Id:"+idStr, &guardados)
		if errConsulta == nil {
			if len(guardados) > 0 {
				for _, registro := range guardados {
					idRegistro := fmt.Sprintf("%.f", registro["Id"].(float64))
					errBorrado = request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento_tipo_publico/"+idRegistro, "DELETE", &auxDelete, nil)
					fmt.Println(errBorrado)
				}
			}
			if errBorrado == nil {
				for _, tipoPublico := range datos {
					publicoMap := tipoPublico.(map[string]interface{})
					perfilID, ok := helpers.InterfaceToInt(publicoMap["responsableID"])
					if !ok || perfilID <= 0 {
						return nil, errors.New("error del servicio UpdateActividadResponsables: perfil de público dirigido inválido")
					}
					nuevoPublico := models.CalendarioEventoTipoPublicoPayload{
						Activo:             activoRelacionPublico(publicoMap),
						PerfilId:           perfilID,
						CalendarioEventoId: models.RelacionID{Id: actividadId},
					}
					errPost := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento_tipo_publico", "POST", &auxUpdate, nuevoPublico)
					if errPost == nil {
						actualizados = append(actualizados, auxUpdate)
					} else {
						logs.Error(errPost)
						return nil, errors.New("error del servicio UpdateActividadResponsables: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")

					}
				}
				return requestresponse.APIResponseDTO(true, 200, actualizados), nil
			} else {
				logs.Error(errBorrado)
				return nil, errors.New("error del servicio UpdateActividadResponsables: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
			}
		} else {
			logs.Error(errConsulta)
			return nil, errors.New("error del servicio UpdateActividadResponsables: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
		}
	} else {
		logs.Error(err)
		return nil, errors.New("error del servicio UpdateActividadResponsables: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
	}
}

func activoRelacionPublico(publico map[string]interface{}) bool {
	activo, ok := publico["Activo"].(bool)
	if !ok {
		activo, ok = publico["activo"].(bool)
	}
	if !ok {
		return true
	}
	return activo
}

func actualizarFechasActividad(idStr string, actividad map[string]interface{}, terceroID int) error {
	fechaInicio, okInicio := actividad["FechaInicio"].(string)
	fechaFin, okFin := actividad["FechaFin"].(string)
	activo, okActivo := actividad["Activo"].(bool)
	if !okInicio && !okFin && !okActivo {
		return nil
	}

	var calendarioEvento map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento/"+idStr, &calendarioEvento); err != nil || calendarioEvento == nil || calendarioEvento["Type"] == "error" {
		return errors.New("error del servicio UpdateActividadResponsables: no fue posible consultar la actividad")
	}
	if okInicio {
		calendarioEvento["FechaInicio"] = fechaInicio
	}
	if okFin {
		calendarioEvento["FechaFin"] = fechaFin
	}
	if okActivo {
		if activo {
			if err := validarActivacionEvento(idStr); err != nil {
				return err
			}
		}
		calendarioEvento["Activo"] = activo
	}
	if okInicio || okFin {
		if err := validarActualizacionFechasGlobalesActividad(idStr, calendarioEvento); err != nil {
			return err
		}
	}
	if err := validarDuplicadoCalendarioEvento(calendarioEvento, idStr); err != nil {
		return errors.New("error del servicio UpdateActividadResponsables: " + err.Error())
	}
	if okInicio {
		calendarioEvento["FechaInicio"] = helpers.FechaTimeParaCRUD(fechaInicio)
	}
	if okFin {
		calendarioEvento["FechaFin"] = helpers.FechaTimeParaCRUD(fechaFin)
	}
	if calendarioEvento["DependenciaId"] == nil || calendarioEvento["DependenciaId"] == "" {
		calendarioEvento["DependenciaId"] = `{"proyectos":[],"fechas":[]}`
	}
	helpers.SetTerceroID(calendarioEvento, terceroID)

	var resultado map[string]interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento/"+idStr, "PUT", &resultado, calendarioEvento); err != nil || resultado == nil || resultado["Type"] == "error" {
		return errors.New("error del servicio UpdateActividadResponsables: no fue posible actualizar las fechas de la actividad")
	}
	if okActivo && !activo {
		if err := inactivarExtensionesActividad(idStr); err != nil {
			logs.Error(err)
		}
	}

	return nil
}

func validarActualizacionFechasGlobalesActividad(idStr string, actividad map[string]interface{}) error {
	fechaInicioStr, _ := actividad["FechaInicio"].(string)
	fechaFinStr, _ := actividad["FechaFin"].(string)
	fechaInicio, err := parseFechaExtension(fechaInicioStr)
	if err != nil {
		return errors.New("La fecha de inicio global no tiene un formato válido.")
	}
	fechaFin, err := parseFechaExtension(fechaFinStr)
	if err != nil {
		return errors.New("La fecha fin global no tiene un formato válido.")
	}
	extensiones, err := extensionesActividad(idStr)
	if err != nil {
		return nil
	}
	for _, extension := range extensiones {
		fechaFinExtensionStr, _ := extension["FechaFin"].(string)
		fechaFinExtension, err := parseFechaExtension(fechaFinExtensionStr)
		if err != nil {
			continue
		}
		if fechaInicio.After(fechaFinExtension) {
			return errors.New("No se puede guardar la fecha inicio global porque supera una extensión activa con fecha fin " + helpers.FormatFechaGMTMinus5(fechaFinExtension) + ". Ajuste la fecha inicio o anule/modifique la extensión primero.")
		}
		if !fechaFinExtension.After(fechaFin) {
			return errors.New("No se puede guardar la fecha fin global porque debe ser menor que la fecha fin de las extensiones activas. Hay una extensión activa hasta " + helpers.FormatFechaGMTMinus5(fechaFinExtension) + ". Anule/modifique la extensión o use una fecha fin global anterior.")
		}
	}
	return nil
}
