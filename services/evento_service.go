package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
)

func PostEvento(data []byte) (interface{}, error) {
	var Evento map[string]interface{}
	var response interface{} = nil
	var success bool = false
	var message string = ""
	if err := json.Unmarshal(data, &Evento); err == nil {

		EventoPost := make(map[string]interface{})
		/* EventoPost["Evento"] = map[string]interface{}{
			// "CalendarioEvento": Evento["Evento"],
			// "EncargadosEvento": Evento["EncargadosEvento"],
			// "TiposPublico": Evento["TiposPublico"]
		}*/

		Evento["Evento"].(map[string]interface{})["Activo"] = true
		EventoPost["CalendarioEvento"] = Evento["Evento"]

		encargadosEvento := make([]map[string]interface{}, 0)
		for _, encargadoTemp := range Evento["EncargadosEvento"].([]interface{}) {
			encargadoEvento := encargadoTemp.(map[string]interface{})
			encargadosEvento = append(encargadosEvento, map[string]interface{}{
				"RolEncargadoEventoId": encargadoEvento["RolEncargadoEventoId"],
				"EncargadoId":          encargadoEvento["EncargadoId"],
				"CalendarioEventoId":   map[string]interface{}{"Id": 0},
				"Activo":               true,
			})
		}
		EventoPost["EncargadosEvento"] = encargadosEvento

		tiposPublico := make([]map[string]interface{}, 0)
		for _, tipoPublicoTemp := range Evento["TiposPublico"].([]interface{}) {
			tipoPublico := tipoPublicoTemp.(map[string]interface{})
			tiposPublico = append(tiposPublico, map[string]interface{}{
				"Nombre":             tipoPublico["Nombre"],
				"CalendarioEventoId": map[string]interface{}{"Id": 0},
				"Activo":             true,
			})
		}
		EventoPost["TiposPublico"] = tiposPublico

		var resultadoEvento map[string]interface{}
		errProduccion := request.SendJson(beego.AppConfig.String("EventoService")+"tr_evento", "POST", &resultadoEvento, EventoPost)
		if resultadoEvento["Type"] == "error" || errProduccion != nil || resultadoEvento["Status"] == "404" || resultadoEvento["Message"] != nil {
			response = resultadoEvento
			success = false
		} else {
			response = Evento
			success = true
		}

	} else {
		message += err.Error()
		success = false
	}

	if success {
		return requestresponse.APIResponseDTO(true, 200, response), nil
	} else {
		if message != "" {
			return nil, errors.New("error del servicio PostEvento: " + message)
		} else {
			return nil, errors.New("error del servicio PostEvento: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
		}
	}
}

func PutEvento(idStr string, data []byte) (interface{}, error) {
	var response interface{} = nil
	var success bool = false
	var message string = ""
	var statusCode int = 400
	var Evento map[string]interface{}

	if err := json.Unmarshal(data, &Evento); err == nil {

		EventoPut := make(map[string]interface{})

		EventoPut["CalendarioEvento"] = Evento["Evento"]
		// EventoPut["EncargadosEvento"] = Evento["EncargadosEvento"];
		// EventoPut["TiposPublico"] = Evento["TiposPublico"];
		EventoPut["EncargadosEventoBorrados"] = Evento["EncargadosEventoBorrados"]
		EventoPut["TiposPublicoBorrados"] = Evento["TiposPublicoBorrados"]
		// Nuevos encargados de evento
		encargadosEvento := make([]map[string]interface{}, 0)
		for _, encargadoTemp := range Evento["EncargadosEvento"].([]interface{}) {
			encargadoEvento := encargadoTemp.(map[string]interface{})
			// solo se agregan los nuevos encargados
			fmt.Println("Encargado", encargadoEvento["Id"], encargadoEvento["EncargadoId"])
			if encargadoEvento["Id"].(float64) == 0 {
				fmt.Println("Agrega Encargado", encargadoEvento["Id"], encargadoEvento["EncargadoId"])
				encargadosEvento = append(encargadosEvento, map[string]interface{}{
					"RolEncargadoEventoId": encargadoEvento["RolEncargadoEventoId"],
					"EncargadoId":          encargadoEvento["EncargadoId"],
					"CalendarioEventoId":   map[string]interface{}{"Id": Evento["Evento"].(map[string]interface{})["Id"]},
					"Activo":               true,
				})
			}
		}
		EventoPut["EncargadosEvento"] = encargadosEvento

		tiposPublico := make([]map[string]interface{}, 0)
		for _, tipoPublicoTemp := range Evento["TiposPublico"].([]interface{}) {
			tipoPublico := tipoPublicoTemp.(map[string]interface{})
			if tipoPublico["Id"] != nil {
				tiposPublico = append(tiposPublico, map[string]interface{}{
					"Nombre":             tipoPublico["Nombre"],
					"CalendarioEventoId": map[string]interface{}{"Id": Evento["Evento"].(map[string]interface{})["Id"]},
					"Id":                 tipoPublico["Id"],
					"Activo":             true,
				})
			} else {
				tiposPublico = append(tiposPublico, map[string]interface{}{
					"Nombre":             tipoPublico["Nombre"],
					"CalendarioEventoId": map[string]interface{}{"Id": Evento["Evento"].(map[string]interface{})["Id"]},
					"Id":                 0,
					"Activo":             true,
				})
			}
		}
		EventoPut["TiposPublico"] = tiposPublico

		var resultadoEvento map[string]interface{}
		errProduccion := request.SendJson(beego.AppConfig.String("EventoService")+"/tr_evento/"+idStr, "PUT", &resultadoEvento, EventoPut)
		if resultadoEvento["Type"] == "error" || errProduccion != nil || resultadoEvento["Status"] == "404" || resultadoEvento["Message"] != nil {
			response = resultadoEvento
			statusCode = 400
			success = false
		} else {
			response = Evento
			statusCode = 200
			success = true
		}

	} else {
		statusCode = 400
		message += err.Error()
		success = false
	}

	if success {
		return requestresponse.APIResponseDTO(success, statusCode, response), nil
	} else {
		if message != "" {
			return nil, errors.New("error del servicio PostEvento: " + message)
		} else {
			return nil, errors.New("error del servicio PostEvento: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
		}
	}
}

func GetEvento(persona string) (interface{}, error) {
	var response interface{} = nil
	var success bool = false
	var message string = ""
	var statusCode int = 400
	var eventos []map[string]interface{}

	fmt.Println("Get Evento")
	personaId, _ := strconv.ParseFloat(persona, 64)
	errEventos := request.GetJson(beego.AppConfig.String("EventoService")+"/tr_evento/"+persona, &eventos)
	if errEventos != nil || eventos[0]["CalendarioEvento"] == nil {
		success = false
		statusCode = 400
		message += "Error: errEventos es nil"

	} else {
		fmt.Println("paso")
		for _, evento := range eventos {

			if evento["CalendarioEvento"] != nil {

				encargados := evento["EncargadosEvento"].([]interface{})
				for _, encargadoTemp := range encargados {
					// seleccionar el rol de la persona
					encargado := encargadoTemp.(map[string]interface{})
					if encargado["EncargadoId"] == personaId {
						evento["RolPersona"] = encargado
					}
					// //cargar nombre del autor
					var encargadoEvento map[string]interface{}
					errEncargado := request.GetJson(beego.AppConfig.String("TercerosService")+"/tercero/"+fmt.Sprintf("%.f", encargado["EncargadoId"].(float64)), &encargadoEvento)
					if encargadoEvento["Type"] == "error" || errEncargado != nil {
						success = false
						statusCode = 400
						message += "Error: errEncargado es nil"
					} else {
						encargado["Nombre"] = encargadoEvento["PrimerNombre"].(string) + " " + encargadoEvento["SegundoNombre"].(string) + " " + encargadoEvento["PrimerApellido"].(string) + " " + encargadoEvento["SegundoApellido"].(string)
					}
				}
				// cargar nombre de la dependencia
				calendarioEvento := evento["CalendarioEvento"].(map[string]interface{})
				tipoEvento := calendarioEvento["TipoEventoId"].(map[string]interface{})
				var dependencia []map[string]interface{}
				errDependencia := request.GetJson(beego.AppConfig.String("OikosService")+"dependencia_tipo_dependencia/?query=DependenciaId__Id:"+fmt.Sprintf("%.f", tipoEvento["DependenciaId"].(float64)), &dependencia)
				if dependencia == nil || errDependencia != nil {
					success = false
					message += "Error: errDependencia es nil"
					statusCode = 400
				} else {
					calendarioEvento["TipoDependenciaId"] = dependencia[0]["TipoDependenciaId"]
					calendarioEvento["DependenciaId"] = dependencia[0]["DependenciaId"]

				}

				// cargar periodo
				var periodo map[string]interface{}
				errPeriodo := request.GetJson(beego.AppConfig.String("ParametroService")+"periodo/"+fmt.Sprintf("%.f", calendarioEvento["PeriodoId"].(float64)), &periodo)
				if periodo == nil || errPeriodo != nil {
					success = false
					message += "Error: errPeriodo es nil"
					statusCode = 400
				} else {
					evento["Periodo"] = periodo["Data"]
				}
				evento["FechaInicio"] = calendarioEvento["FechaInicio"]
				evento["FechaFin"] = calendarioEvento["FechaFin"]
				evento["Descripcion"] = calendarioEvento["Descripcion"]
				evento["TipoEvento"] = tipoEvento["Nombre"]
				evento["Dependencia"] = calendarioEvento["DependenciaId"].(map[string]interface{})["Nombre"]

			}
		}
		response = eventos
		success = true
		statusCode = 200
	}

	if success {
		return requestresponse.APIResponseDTO(success, statusCode, response), nil
	} else {
		if message != "" {
			return nil, errors.New("error del servicio GetEvento: " + message)
		} else {
			return nil, errors.New("error del servicio GetEvento: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
		}
	}
}

func DeleteEvento(id string) (interface{}, error) {
	var eventoDeleted map[string]interface{}

	errEvento := request.SendJson(fmt.Sprintf("%s", beego.AppConfig.String("EventoService")+"/tr_evento/"+id), "DELETE", &eventoDeleted, nil)
	if errEvento != nil || eventoDeleted["Message"] != nil {
		return nil, errors.New("error del servicio DeleteEvento: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
	} else {
		cadena, ok := eventoDeleted["Type"].(string)
		if ok {
			if cadena == "error" {
				return nil, errors.New("error del servicio DeleteEvento: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
			} else {
				return requestresponse.APIResponseDTO(true, 200, nil), nil
			}
		} else {
			return nil, errors.New("error del servicio DeleteEvento: La solicitud contiene un tipo de dato incorrecto o un parámetro inválido")
		}

	}
}
