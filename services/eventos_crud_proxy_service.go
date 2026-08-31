package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/udistrital/sga_calendario_mid/helpers"
	"github.com/udistrital/sga_calendario_mid/models"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
)

var recursosEventosPermitidos = map[string]bool{
	"calendario":                           true,
	"proceso":                              true,
	"calendario_evento":                    true,
	"evento_catalogo":                      true,
	"evento_catalogo_rol_gestion":          true,
	"evento_catalogo_proceso_catalogo":     true,
	"calendario_evento_extension":          true,
	"calendario_evento_extension_programa": true,
	"proceso_catalogo":                     true,
	"tipo_recurrencia":                     true,
}

var recursosConLecturaPut = map[string]bool{
	"calendario":                           true,
	"proceso":                              true,
	"calendario_evento":                    true,
	"evento_catalogo_rol_gestion":          true,
	"evento_catalogo_proceso_catalogo":     true,
	"calendario_evento_tipo_publico":       true,
	"calendario_evento_extension":          true,
	"calendario_evento_extension_programa": true,
}

func validarRecursoEventos(recurso string) error {
	if recurso == "" || strings.Contains(recurso, "/") || !recursosEventosPermitidos[recurso] {
		return errors.New("recurso de eventos no permitido")
	}
	return nil
}

func GetEventosCrud(recurso string, id string, query string) (interface{}, error) {
	if err := validarRecursoEventos(recurso); err != nil {
		return nil, err
	}
	url := beego.AppConfig.String("EventoService") + recurso
	if id != "" {
		url += "/" + id
	}
	if query != "" {
		url += "?" + query
	}

	var resultado interface{}
	if err := request.GetJson(url, &resultado); err != nil {
		return nil, errors.New("no fue posible consultar el recurso de eventos")
	}
	if resultado == nil && id == "" {
		return requestresponse.APIResponseDTO(true, 200, []interface{}{}), nil
	}
	if resultado == nil {
		return nil, errors.New("no fue posible consultar el recurso de eventos")
	}

	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func PerfilesConfiguracionSGA(authHeader string) ([]models.PerfilConfiguracion, error) {
	configuracionService := beego.AppConfig.String("ConfiguracionService")
	if configuracionService == "" {
		return nil, errors.New("ConfiguracionService no configurado")
	}

	aplicacionID, err := aplicacionConfiguracionID("SGA_MF", authHeader)
	if err != nil {
		return nil, err
	}

	var perfiles []map[string]interface{}
	if err := getJsonConfiguracion(configuracionService+"perfil/?query=Aplicacion.Id:"+strconv.Itoa(aplicacionID)+"&limit=0", authHeader, &perfiles); err != nil {
		return nil, err
	}

	resultado := make([]models.PerfilConfiguracion, 0, len(perfiles))
	for _, perfil := range perfiles {
		nombre := strings.TrimSpace(fmt.Sprintf("%v", perfil["Nombre"]))
		if nombre == "" || nombre == "<nil>" {
			nombre = strings.TrimSpace(fmt.Sprintf("%v", perfil["nombre"]))
		}
		if nombre == "" || nombre == "<nil>" {
			continue
		}
		codigo := strings.TrimSpace(fmt.Sprintf("%v", perfil["CodigoAbreviacion"]))
		if codigo == "" || codigo == "<nil>" {
			codigo = strings.TrimSpace(fmt.Sprintf("%v", perfil["codigo_abreviacion"]))
		}
		if codigo == "" || codigo == "<nil>" {
			codigo = nombre
		}
		perfilID, _ := helpers.InterfaceToInt(perfil["Id"])
		if perfilID == 0 {
			perfilID, _ = helpers.InterfaceToInt(perfil["id"])
		}
		resultado = append(resultado, models.PerfilConfiguracion{
			Id:                perfilID,
			Nombre:            nombre,
			CodigoAbreviacion: codigo,
			Activo:            true,
		})
	}
	return resultado, nil
}

func aplicacionConfiguracionID(alias string, authHeader string) (int, error) {
	var aplicaciones []map[string]interface{}
	if err := getJsonConfiguracion(beego.AppConfig.String("ConfiguracionService")+"aplicacion/?query=Alias:"+alias+"&limit=1", authHeader, &aplicaciones); err != nil {
		return 0, err
	}
	if len(aplicaciones) == 0 || len(aplicaciones[0]) == 0 {
		return 0, errors.New("No se encontró la aplicación de Configuración con alias " + alias)
	}
	if id, ok := helpers.InterfaceToInt(aplicaciones[0]["Id"]); ok {
		return id, nil
	}
	if id, ok := helpers.InterfaceToInt(aplicaciones[0]["id"]); ok {
		return id, nil
	}
	return 0, errors.New("La aplicación de Configuración no tiene identificador válido")
}

func getJsonConfiguracion(url string, authHeader string, target interface{}) error {
	if strings.TrimSpace(authHeader) != "" {
		request.SetHeader(authHeader)
		defer request.SetHeader("")
	}
	return request.GetJson(url, target)
}

func PostEventosCrud(recurso string, data []byte) (interface{}, error) {
	if err := validarRecursoEventos(recurso); err != nil {
		return nil, err
	}

	var payload interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, errors.New("solicitud inválida")
	}
	if err := validarFechasPayloadEventos(recurso, "", payload); err != nil {
		return nil, err
	}
	helpers.NormalizeFechasTimeCalendarioEvento(recurso, payload)
	if recurso == "calendario_evento" {
		terceroID, err := helpers.TerceroIDFromPayload(payload)
		if err != nil {
			return nil, err
		}
		payloadMap, ok := payload.(map[string]interface{})
		if !ok {
			return nil, errors.New("solicitud inválida")
		}
		helpers.SetTerceroID(payloadMap, terceroID)
		resultado, err := crearCalendarioEvento(payloadMap)
		if err != nil {
			return nil, err
		}
		return requestresponse.APIResponseDTO(true, 200, resultado), nil
	}

	var recibido interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+recurso, "POST", &recibido, payload); err != nil || recibido == nil {
		return nil, errors.New("no fue posible crear el recurso de eventos")
	}

	return requestresponse.APIResponseDTO(true, 200, recibido), nil
}

func PutEventosCrud(recurso string, id string, data []byte) (interface{}, error) {
	if err := validarRecursoEventos(recurso); err != nil {
		return nil, err
	}
	if id == "" {
		return nil, errors.New("id requerido")
	}

	var payload interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, errors.New("solicitud inválida")
	}
	if err := validarFechasPayloadEventos(recurso, id, payload); err != nil {
		return nil, err
	}
	helpers.NormalizeFechasTimeCalendarioEvento(recurso, payload)
	if err := validarCambioPoliticaEventos(recurso, id, payload, false); err != nil {
		return nil, err
	}
	if recurso == "calendario_evento" {
		terceroID, err := helpers.TerceroIDFromPayload(payload)
		if err != nil {
			return nil, err
		}
		if payloadMap, ok := payload.(map[string]interface{}); ok {
			helpers.SetTerceroID(payloadMap, terceroID)
		}
	}
	if activo, ok := helpers.ActivoFromPayload(payload); ok && activo {
		switch recurso {
		case "proceso":
			if err := validarActivacionProceso(id); err != nil {
				return nil, err
			}
		case "calendario_evento":
			if err := validarActivacionEvento(id); err != nil {
				return nil, err
			}
		}
	}
	var actual interface{}
	if recurso == "calendario_evento" {
		if err := request.GetJson(beego.AppConfig.String("EventoService")+recurso+"/"+id, &actual); err != nil {
			return nil, errors.New("no fue posible consultar la actividad para validar duplicados")
		}
		payloadMap, ok := payload.(map[string]interface{})
		actualMap, actualOK := actual.(map[string]interface{})
		if !ok || !actualOK || actualMap == nil || actualMap["Type"] == "error" {
			return nil, errors.New("no fue posible consultar la actividad para validar duplicados")
		}
		delete(payloadMap, "NumeroOcurrencia")
		efectivo := calendarioEventoEfectivo(actualMap, payloadMap)
		if cambioIdentidadCalendarioEvento(actualMap, efectivo) {
			repetible, err := esRepetibleCalendarioEvento(efectivo)
			if err != nil {
				return nil, err
			}
			numero, err := siguienteNumeroOcurrencia(efectivo, repetible)
			if err != nil {
				return nil, err
			}
			payloadMap["NumeroOcurrencia"] = numero
			efectivo["NumeroOcurrencia"] = numero
		}
		if err := validarDuplicadoCalendarioEvento(efectivo, id); err != nil {
			return nil, err
		}
	}

	var resultado interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+recurso+"/"+id, "PUT", &resultado, payload); err != nil || resultado == nil {
		return nil, errors.New("no fue posible actualizar el recurso de eventos")
	}

	if recursosConLecturaPut[recurso] {
		resultado = entidadActualizadaEventos(recurso, id, resultado)
	}

	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func cambioIdentidadCalendarioEvento(actual map[string]interface{}, efectivo map[string]interface{}) bool {
	procesoAnterior, errProcesoAnterior := helpers.RelationIDToString(actual["ProcesoId"])
	procesoNuevo, errProcesoNuevo := helpers.RelationIDToString(efectivo["ProcesoId"])
	eventoAnterior, errEventoAnterior := helpers.RelationIDToString(actual["EventoCatalogoId"])
	eventoNuevo, errEventoNuevo := helpers.RelationIDToString(efectivo["EventoCatalogoId"])
	if errProcesoAnterior != nil || errProcesoNuevo != nil || errEventoAnterior != nil || errEventoNuevo != nil {
		return false
	}
	return procesoAnterior != procesoNuevo || eventoAnterior != eventoNuevo
}

func entidadActualizadaEventos(recurso string, id string, fallback interface{}) interface{} {
	var actualizado interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+recurso+"/"+id, &actualizado); err != nil || actualizado == nil {
		return fallback
	}
	return actualizado
}

func DeleteEventosCrud(recurso string, id string, data []byte) (interface{}, error) {
	if err := validarRecursoEventos(recurso); err != nil {
		return nil, err
	}
	if id == "" {
		return nil, errors.New("id requerido")
	}
	var payload interface{}
	if len(strings.TrimSpace(string(data))) > 0 {
		if err := json.Unmarshal(data, &payload); err != nil {
			return nil, errors.New("solicitud inválida")
		}
	}
	if recurso == "calendario_evento" {
		terceroID, err := helpers.TerceroIDFromPayload(payload)
		if err != nil {
			return nil, err
		}
		if payloadMap, ok := payload.(map[string]interface{}); ok {
			helpers.SetTerceroID(payloadMap, terceroID)
		}
	}
	if err := validarCambioPoliticaEventos(recurso, id, payload, true); err != nil {
		return nil, err
	}
	var resultado interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+recurso+"/"+id, "DELETE", &resultado, payload); err != nil || resultado == nil {
		return nil, errors.New("no fue posible eliminar el recurso de eventos")
	}

	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func validarCambioPoliticaEventos(recurso string, id string, payload interface{}, eliminar bool) error {
	if recurso != "proceso" && recurso != "evento_catalogo_proceso_catalogo" {
		return nil
	}

	var actual map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+recurso+"/"+id, &actual); err != nil || actual == nil || actual["Type"] == "error" {
		return errors.New("no fue posible consultar el recurso de eventos")
	}
	payloadMap, _ := payload.(map[string]interface{})

	if recurso == "proceso" {
		if nuevo, existe := payloadMap["ProcesoCatalogoId"]; existe {
			anteriorID, errAnterior := helpers.RelationIDToString(actual["ProcesoCatalogoId"])
			nuevoID, errNuevo := helpers.RelationIDToString(nuevo)
			if errAnterior != nil || errNuevo != nil || anteriorID != nuevoID {
				return errors.New("el catálogo de un proceso existente no se puede modificar")
			}
		}
		return nil
	}

	procesoCatalogoID, err := helpers.RelationIDToString(actual["ProcesoCatalogoId"])
	if err != nil {
		return errors.New("la relación de catálogos es inválida")
	}
	eventoCatalogoID, err := helpers.RelationIDToString(actual["EventoCatalogoId"])
	if err != nil {
		return errors.New("la relación de catálogos es inválida")
	}
	if nuevo, existe := payloadMap["ProcesoCatalogoId"]; existe {
		nuevoID, err := helpers.RelationIDToString(nuevo)
		if err != nil || nuevoID != procesoCatalogoID {
			return errors.New("los catálogos de una relación existente no se pueden modificar")
		}
	}
	if nuevo, existe := payloadMap["EventoCatalogoId"]; existe {
		nuevoID, err := helpers.RelationIDToString(nuevo)
		if err != nil || nuevoID != eventoCatalogoID {
			return errors.New("los catálogos de una relación existente no se pueden modificar")
		}
	}

	desactivar := false
	if activo, existe := helpers.ActivoFromPayload(payload); existe {
		desactivar = !activo
	}
	if eliminar || desactivar {
		tieneActivas, err := relacionTieneOcurrenciasActivas(procesoCatalogoID, eventoCatalogoID, false)
		if err != nil {
			return err
		}
		if tieneActivas {
			return errors.New("no se puede retirar una relación utilizada por actividades activas")
		}
	}

	repetibleAnterior, _ := actual["Repetible"].(bool)
	repetibleNuevo, cambiaRepetible := payloadMap["Repetible"].(bool)
	if cambiaRepetible && repetibleAnterior && !repetibleNuevo {
		tieneMultiples, err := relacionTieneOcurrenciasActivas(procesoCatalogoID, eventoCatalogoID, true)
		if err != nil {
			return err
		}
		if tieneMultiples {
			return errors.New("no se puede marcar la relación como no repetible mientras existan varias ocurrencias activas")
		}
	}
	return nil
}

func relacionTieneOcurrenciasActivas(procesoCatalogoID string, eventoCatalogoID string, multiples bool) (bool, error) {
	var procesos []map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"proceso?query=ProcesoCatalogoId__Id:"+procesoCatalogoID+"&limit=0", &procesos); err != nil {
		return false, errors.New("no fue posible validar las actividades de la relación")
	}
	for _, proceso := range procesos {
		procesoID, err := helpers.RelationIDToString(proceso["Id"])
		if err != nil {
			continue
		}
		var actividades []map[string]interface{}
		url := beego.AppConfig.String("EventoService") + "calendario_evento?query=ProcesoId__Id:" + procesoID + ",EventoCatalogoId__Id:" + eventoCatalogoID + ",Activo:true&limit=0"
		if err := request.GetJson(url, &actividades); err != nil {
			return false, errors.New("no fue posible validar las actividades de la relación")
		}
		if (!multiples && len(actividades) > 0) || (multiples && len(actividades) > 1) {
			return true, nil
		}
	}
	return false, nil
}
