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

var recursosAuditables = map[string]bool{
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
		perfilID, _ := interfaceToInt(perfil["Id"])
		if perfilID == 0 {
			perfilID, _ = interfaceToInt(perfil["id"])
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
	if id, ok := interfaceToInt(aplicaciones[0]["Id"]); ok {
		return id, nil
	}
	if id, ok := interfaceToInt(aplicaciones[0]["id"]); ok {
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

func PostEventosCrud(recurso string, data []byte, usuario string) (interface{}, error) {
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
	normalizarFechasTimeCalendarioEvento(recurso, payload)

	var recibido interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+recurso, "POST", &recibido, payload); err != nil || recibido == nil {
		return nil, errors.New("no fue posible crear el recurso de eventos")
	}

	if recursosAuditables[recurso] {
		if id, ok := extractId(recibido); ok {
			RegistrarAuditoria(recurso, id, "POST", nil, recibido, usuario, "PostEventosCrud/"+recurso)
		}
	}

	return requestresponse.APIResponseDTO(true, 200, recibido), nil
}

func PutEventosCrud(recurso string, id string, data []byte, usuario string) (interface{}, error) {
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
	normalizarFechasTimeCalendarioEvento(recurso, payload)
	if activo, ok := activoFromPayload(payload); ok && activo {
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
	var anterior interface{}
	if recursosAuditables[recurso] {
		request.GetJson(beego.AppConfig.String("EventoService")+recurso+"/"+id, &anterior)
	}

	var resultado interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+recurso+"/"+id, "PUT", &resultado, payload); err != nil || resultado == nil {
		return nil, errors.New("no fue posible actualizar el recurso de eventos")
	}

	if recursosAuditables[recurso] {
		entidadId, _ := strconv.Atoi(id)
		RegistrarAuditoria(recurso, entidadId, "PUT", anterior, resultado, usuario, "PutEventosCrud/"+recurso)
	}

	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func normalizarFechasTimeCalendarioEvento(recurso string, payload interface{}) {
	if recurso != "calendario_evento" {
		return
	}
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return
	}
	for _, campo := range []string{"FechaInicio", "FechaFin"} {
		if value, ok := payloadMap[campo]; ok && value != nil {
			payloadMap[campo] = helpers.FechaTimeParaCRUD(value)
		}
	}
}

func activoFromPayload(payload interface{}) (bool, bool) {
	m, ok := payload.(map[string]interface{})
	if !ok {
		return false, false
	}
	activo, ok := m["Activo"].(bool)
	return activo, ok
}

func DeleteEventosCrud(recurso string, id string, usuario string) (interface{}, error) {
	if err := validarRecursoEventos(recurso); err != nil {
		return nil, err
	}
	if id == "" {
		return nil, errors.New("id requerido")
	}
	var anterior interface{}
	if recursosAuditables[recurso] {
		request.GetJson(beego.AppConfig.String("EventoService")+recurso+"/"+id, &anterior)
	}

	var resultado interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+recurso+"/"+id, "DELETE", &resultado, nil); err != nil || resultado == nil {
		return nil, errors.New("no fue posible eliminar el recurso de eventos")
	}

	if recursosAuditables[recurso] {
		entidadId, _ := strconv.Atoi(id)
		RegistrarAuditoria(recurso, entidadId, "DELETE", anterior, nil, usuario, "DeleteEventosCrud/"+recurso)
	}

	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func extractId(v interface{}) (int, bool) {
	return helpers.ExtractID(v)
}
