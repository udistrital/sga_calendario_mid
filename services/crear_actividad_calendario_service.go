package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
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
		Actividad := actividadCalendario["Actividad"]
		//Solicitid post a eventos service enviando el json recibido
		errActividad := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento", "POST", &actividadCalendarioPost, Actividad)
		if errActividad == nil && fmt.Sprintf("%v", actividadCalendarioPost["System"]) != "map[]" && actividadCalendarioPost["Id"] != nil {
			if actividadCalendarioPost["Status"] != 400 {
				IdActividad = actividadCalendarioPost["Id"]
			} else {
				logs.Error(errActividad)
			}
		} else {
			logs.Error(errActividad)
		}

		var totalPublico []interface{}
		//Guarda el JSON de la tabla tipo publico
		totalPublico = actividadCalendario["responsable"].([]interface{})

		for _, publicoTemp := range totalPublico {
			CalendarioEventoTipoPersona := map[string]interface{}{
				"Activo":             true,
				"TipoPublicoId":      map[string]interface{}{"Id": publicoTemp.(map[string]interface{})["responsableID"].(float64)},
				"CalendarioEventoId": map[string]interface{}{"Id": IdActividad.(float64)},
			}

			errActividadPersona := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento_tipo_publico", "POST", &actividadPersonaPost, CalendarioEventoTipoPersona)

			if errActividadPersona == nil && fmt.Sprintf("%v", actividadPersonaPost["System"]) != "map[]" && actividadPersonaPost["Id"] != nil {
				if actividadPersonaPost["Status"] != 400 {
					return requestresponse.APIResponseDTO(true, 200, actividadCalendarioPost), nil
				} else {
					var resultado2 map[string]interface{}
					request.SendJson(fmt.Sprintf(beego.AppConfig.String("EventoService")+"/calendario_evento/%.f", actividadCalendarioPost["Id"]), "DELETE", &resultado2, nil)
					logs.Error(errActividadPersona)
				}
			} else {
				logs.Error(errActividadPersona)
			}
		}
	}
	return nil, errors.New("error del servicio PostActividadCalendario: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
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
		datos := recibido["resp"].([]interface{})
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
					nuevoPublico := map[string]interface{}{
						"Activo":             true,
						"TipoPublicoId":      map[string]interface{}{"Id": tipoPublico.(map[string]interface{})["responsableID"]},
						"CalendarioEventoId": map[string]interface{}{"Id": actividadId},
					}
					errPost := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento_tipo_publico", "POST", &auxUpdate, nuevoPublico)
					if errPost == nil {
						actualizados = append(actualizados, auxUpdate)
					} else {
						logs.Error(errPost)
						return nil, errors.New("error del servicio UpdateActividadResponsables: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")

					}
				}
				return requestresponse.APIResponseDTO(false, 400, actualizados), nil
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
