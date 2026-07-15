package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/udistrital/sga_calendario_mid/helpers"
	"github.com/udistrital/sga_calendario_mid/models"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
)

func PostExtensionActividad(idActividad string, data []byte, usuario string, authHeader string) (interface{}, error) {
	var solicitud models.SolicitudExtensionActividad
	if err := json.Unmarshal(data, &solicitud); err != nil {
		return nil, errors.New("error del servicio PostExtensionActividad: solicitud inválida")
	}
	if solicitud.FechaFin.IsZero() {
		return nil, errors.New("error del servicio PostExtensionActividad: fecha fin requerida")
	}
	fechaFinSolicitud := solicitud.FechaFin.Format(time.RFC3339)
	if len(solicitud.Dependencias) == 0 {
		return nil, errors.New("error del servicio PostExtensionActividad: debe seleccionar al menos una dependencia")
	}
	idActividadInt, err := strconv.Atoi(idActividad)
	if err != nil {
		return nil, errors.New("error del servicio PostExtensionActividad: actividad inválida")
	}

	actividad, err := obtenerActividad(idActividad)
	if err != nil {
		return nil, err
	}
	fechaInicioActividad, _ := actividad["FechaInicio"].(string)
	if err := helpers.ValidarRangoFechas(fechaInicioActividad, fechaFinSolicitud); err != nil {
		return nil, errors.New("la fecha fin de la extensión no puede ser menor que la fecha inicio de la actividad")
	}
	fechaFinPayload, err := fechaParaModelo(solicitud.FechaFin)
	if err != nil {
		return nil, errors.New("error del servicio PostExtensionActividad: fecha fin inválida")
	}

	dependencias := make([]int, 0, len(solicitud.Dependencias))
	padres := make(map[int]interface{})
	for _, dependenciaRaw := range solicitud.Dependencias {
		dependenciaID, ok := interfaceToInt(dependenciaRaw)
		if !ok || dependenciaID <= 0 {
			return nil, errors.New("error del servicio PostExtensionActividad: dependencia inválida")
		}
		if !actividadIncluyeDependencia(actividad, dependenciaID) {
			return nil, errors.New("No se puede crear la extensión: la dependencia seleccionada no está asociada a la actividad.")
		}
		rango, err := rangoPermitidoActividadDependencia(idActividad, dependenciaID, actividad)
		if err != nil {
			return nil, err
		}
		if !fechaFinPayload.After(rango.RangoPermitido.FechaFin) {
			return nil, errors.New("la fecha fin de la extensión debe ser mayor a la fecha fin vigente de todos los programas seleccionados")
		}
		dependencias = append(dependencias, dependenciaID)
		padres[dependenciaID] = rango.ExtensionVigenteId
	}

	numeroExtension, err := siguienteNumeroExtension(idActividad)
	if err != nil {
		return nil, err
	}
	extensionPayload := models.CalendarioEventoExtensionPayload{
		CalendarioEventoId: models.RelacionID{Id: idActividadInt},
		FechaFin:           fechaFinPayload,
		DocumentoId:        documentoIdOrNil(solicitud.DocumentoId),
		Descripcion:        solicitud.Descripcion,
		NumeroExtension:    numeroExtension,
		Activo:             true,
	}

	var extension map[string]interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento_extension", "POST", &extension, extensionPayload); err != nil || extension == nil || extension["Type"] == "error" {
		return nil, errors.New("error del servicio PostExtensionActividad: no fue posible crear la extensión")
	}
	extensionID, ok := idToString(extension["Id"])
	if !ok {
		return nil, errors.New("error del servicio PostExtensionActividad: extensión creada sin identificador")
	}
	extensionIDInt, _ := strconv.Atoi(extensionID)
	if id, err := strconv.Atoi(extensionID); err == nil {
		RegistrarAuditoria("calendario_evento_extension", id, "POST", nil, extension, usuario, "PostExtensionActividad/calendario_evento_extension")
	}

	for _, dependenciaID := range dependencias {
		if err := inactivarVigenciaExtensionDependencia(idActividad, dependenciaID, usuario); err != nil {
			return nil, err
		}
		relacionPayload := models.CalendarioEventoExtensionProgramaPayload{
			CalendarioEventoId:          models.RelacionID{Id: idActividadInt},
			CalendarioEventoExtensionId: models.RelacionID{Id: extensionIDInt},
			DependenciaId:               dependenciaID,
			Vigente:                     true,
			Activo:                      true,
		}
		if padre := padres[dependenciaID]; padre != nil {
			if padreID, ok := interfaceToInt(padre); ok {
				relacionPayload.ExtensionPadreId = &models.RelacionID{Id: padreID}
			}
		}
		var relacion map[string]interface{}
		if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento_extension_programa", "POST", &relacion, relacionPayload); err != nil || relacion == nil || relacion["Type"] == "error" {
			return nil, errors.New("error del servicio PostExtensionActividad: no fue posible asociar dependencia a la extensión")
		}
		if id, ok := extractId(relacion); ok {
			RegistrarAuditoria("calendario_evento_extension_programa", id, "POST", nil, relacion, usuario, "PostExtensionActividad/calendario_evento_extension_programa")
		}
	}

	return requestresponse.APIResponseDTO(true, 200, extension), nil
}

func GetExtensionesActividad(idActividad string) (interface{}, error) {
	extensiones, err := extensionesActividad(idActividad)
	if err != nil {
		return nil, err
	}
	normalizarExtensionesSalida(extensiones)
	return requestresponse.APIResponseDTO(true, 200, extensiones), nil
}

func PutExtensionActividad(idActividad string, idExtension string, data []byte, usuario string, authHeader string) (interface{}, error) {
	var solicitud models.SolicitudExtensionActividad
	if err := json.Unmarshal(data, &solicitud); err != nil {
		return nil, errors.New("error del servicio PutExtensionActividad: solicitud inválida")
	}
	if solicitud.FechaFin.IsZero() {
		return nil, errors.New("error del servicio PutExtensionActividad: fecha fin requerida")
	}
	fechaFinSolicitud := solicitud.FechaFin.Format(time.RFC3339)
	idActividadInt, err := strconv.Atoi(idActividad)
	if err != nil || idActividadInt <= 0 {
		return nil, errors.New("error del servicio PutExtensionActividad: actividad inválida")
	}
	fechaFinPayload, err := fechaParaModelo(solicitud.FechaFin)
	if err != nil {
		return nil, errors.New("error del servicio PutExtensionActividad: fecha fin inválida")
	}

	extension, err := obtenerExtensionActividad(idActividad, idExtension)
	if err != nil {
		return nil, err
	}
	relaciones, err := relacionesExtensionActividad(idActividad, idExtension)
	if err != nil {
		return nil, err
	}
	if len(relaciones) == 0 {
		return nil, errors.New("error del servicio PutExtensionActividad: extensión sin dependencias asociadas")
	}
	actividad, err := obtenerActividad(idActividad)
	if err != nil {
		return nil, err
	}
	fechaInicioActividad, _ := actividad["FechaInicio"].(string)
	if err := helpers.ValidarRangoFechas(fechaInicioActividad, fechaFinSolicitud); err != nil {
		return nil, errors.New("la fecha fin de la extensión no puede ser menor que la fecha inicio de la actividad")
	}
	fechaFinBaseActividad, err := fechaParaModelo(actividad["FechaFin"])
	if err != nil {
		return nil, errors.New("error del servicio PutExtensionActividad: fecha fin original inválida")
	}

	for _, relacion := range relaciones {
		dependenciaID, _ := interfaceToInt(relacion["DependenciaId"])
		if !actividadIncluyeDependencia(actividad, dependenciaID) {
			return nil, errors.New("No se puede editar la extensión: una dependencia asociada a la extensión ya no pertenece a la actividad.")
		}
		limiteAnterior := fechaFinBaseActividad
		if padre, ok := relacion["ExtensionPadreId"].(map[string]interface{}); ok && padre != nil {
			if fechaPadre, ok := padre["FechaFin"]; ok && fmt.Sprintf("%v", fechaPadre) != "" {
				if parsedPadre, err := fechaParaModelo(fechaPadre); err == nil {
					limiteAnterior = parsedPadre
				}
			}
		}
		if !fechaFinPayload.After(limiteAnterior) {
			return nil, errors.New("la fecha fin de la extensión debe ser mayor a la fecha fin previa de todas las dependencias asociadas")
		}
	}

	anterior := deepCopyMap(extension)
	extension["FechaFin"] = fechaFinPayload
	extension["DocumentoId"] = documentoIdOrNil(solicitud.DocumentoId)
	extension["Descripcion"] = solicitud.Descripcion
	extension["CalendarioEventoId"] = map[string]interface{}{"Id": idActividadInt}
	var resultado map[string]interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento_extension/"+idExtension, "PUT", &resultado, extension); err != nil || resultado == nil || resultado["Type"] == "error" {
		return nil, errors.New("error del servicio PutExtensionActividad: no fue posible actualizar la extensión")
	}
	if id, err := strconv.Atoi(idExtension); err == nil {
		RegistrarAuditoria("calendario_evento_extension", id, "PUT", anterior, resultado, usuario, "PutExtensionActividad/calendario_evento_extension")
	}
	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func DeleteExtensionActividad(idActividad string, idExtension string, usuario string, authHeader string) (interface{}, error) {
	extension, err := obtenerExtensionActividad(idActividad, idExtension)
	if err != nil {
		return nil, err
	}
	idActividadInt, _ := strconv.Atoi(idActividad)
	idExtensionInt, _ := strconv.Atoi(idExtension)
	relaciones, err := relacionesExtensionActividad(idActividad, idExtension)
	if err != nil {
		return nil, err
	}
	for _, relacion := range relaciones {
		idRelacion, ok := idToString(relacion["Id"])
		if !ok {
			continue
		}
		idRelacionInt, _ := strconv.Atoi(idRelacion)
		anteriorRelacion := deepCopyMap(relacion)
		eraVigente := relacion["Vigente"] == true
		padre := relacion["ExtensionPadreId"]
		dependenciaID, _ := interfaceToInt(relacion["DependenciaId"])
		relacionPayload := models.CalendarioEventoExtensionProgramaPayload{
			Id:                          idRelacionInt,
			CalendarioEventoId:          models.RelacionID{Id: idActividadInt},
			CalendarioEventoExtensionId: models.RelacionID{Id: idExtensionInt},
			DependenciaId:               dependenciaID,
			Vigente:                     false,
			Activo:                      false,
			FechaCreacion:               fechaParaModeloSinError(relacion["FechaCreacion"]),
		}
		if padreID, ok := idRelacionExtension(padre); ok {
			relacionPayload.ExtensionPadreId = &models.RelacionID{Id: padreID}
		}
		var resultadoRelacion map[string]interface{}
		if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento_extension_programa/"+idRelacion, "PUT", &resultadoRelacion, relacionPayload); err != nil || resultadoRelacion == nil || resultadoRelacion["Type"] == "error" {
			return nil, errors.New("error del servicio DeleteExtensionActividad: no fue posible inactivar relación de extensión")
		}
		if id, err := strconv.Atoi(idRelacion); err == nil {
			RegistrarAuditoria("calendario_evento_extension_programa", id, "PUT", anteriorRelacion, resultadoRelacion, usuario, "DeleteExtensionActividad/calendario_evento_extension_programa")
		}
		if eraVigente {
			reactivarRelacionPadre(idActividad, relacion, padre, usuario)
		}
	}
	anterior := deepCopyMap(extension)
	numeroExtension, _ := interfaceToInt(extension["NumeroExtension"])
	extensionPayload := models.CalendarioEventoExtensionPayload{
		Id:                 idExtensionInt,
		CalendarioEventoId: models.RelacionID{Id: idActividadInt},
		FechaFin:           fechaParaModeloSinError(extension["FechaFin"]),
		DocumentoId:        documentoIdOrNil(extension["DocumentoId"]),
		Descripcion:        fmt.Sprintf("%v", extension["Descripcion"]),
		NumeroExtension:    numeroExtension,
		Activo:             false,
		FechaCreacion:      fechaParaModeloSinError(extension["FechaCreacion"]),
	}
	var resultado map[string]interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento_extension/"+idExtension, "PUT", &resultado, extensionPayload); err != nil || resultado == nil || resultado["Type"] == "error" {
		return nil, errors.New("error del servicio DeleteExtensionActividad: no fue posible inactivar la extensión")
	}
	if id, err := strconv.Atoi(idExtension); err == nil {
		RegistrarAuditoria("calendario_evento_extension", id, "PUT", anterior, resultado, usuario, "DeleteExtensionActividad/calendario_evento_extension")
	}
	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func GetRangoActividadDependencia(idActividad string, dependencia string) (interface{}, error) {
	dependenciaID, err := strconv.Atoi(dependencia)
	if err != nil {
		return nil, errors.New("error del servicio GetRangoActividadDependencia: dependencia inválida")
	}
	actividad, err := obtenerActividad(idActividad)
	if err != nil {
		return nil, err
	}
	rango, err := rangoPermitidoActividadDependencia(idActividad, dependenciaID, actividad)
	if err != nil {
		return nil, err
	}
	return requestresponse.APIResponseDTO(true, 200, rango), nil
}

func obtenerActividad(idActividad string) (map[string]interface{}, error) {
	var actividad map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento/"+idActividad, &actividad); err != nil || actividad == nil || actividad["Type"] == "error" {
		return nil, errors.New("error del servicio de extensión: no fue posible consultar la actividad")
	}
	return actividad, nil
}

func rangoPermitidoActividadDependencia(idActividad string, dependenciaID int, actividad map[string]interface{}) (models.RangoActividadDependencia, error) {
	if !actividadIncluyeDependencia(actividad, dependenciaID) {
		return models.RangoActividadDependencia{}, errors.New("La dependencia consultada no está asociada a la actividad; no aplica rango ni extensión.")
	}
	fechaInicio := fechaSalidaCalendario(actividad["FechaInicio"])
	fechaFinOriginal := fechaSalidaCalendario(actividad["FechaFin"])
	rango := models.RangoActividadDependencia{
		CalendarioEventoId: idActividad,
		DependenciaId:      dependenciaID,
		RangoOriginal: models.RangoFechas{
			FechaInicio: fechaInicio,
			FechaFin:    fechaFinOriginal,
		},
		RangoPermitido: models.RangoFechas{
			FechaInicio: fechaInicio,
			FechaFin:    fechaFinOriginal,
		},
		ExtensionVigenteId: nil,
	}

	var relaciones []map[string]interface{}
	url := beego.AppConfig.String("EventoService") + "calendario_evento_extension_programa?query=Activo:true,Vigente:true,CalendarioEventoId__Id:" + idActividad + ",DependenciaId:" + strconv.Itoa(dependenciaID) + "&limit=1"
	if err := request.GetJson(url, &relaciones); err != nil || len(relaciones) == 0 || len(relaciones[0]) == 0 {
		return rango, nil
	}
	extension, ok := relaciones[0]["CalendarioEventoExtensionId"].(map[string]interface{})
	if !ok || extension == nil {
		return rango, nil
	}
	fechaFinExtension := fechaSalidaCalendario(extension["FechaFin"])
	rango.RangoPermitido.FechaFin = fechaFinExtension
	rango.ExtensionVigenteId = extension["Id"]
	return rango, nil
}

func actividadIncluyeDependencia(actividad map[string]interface{}, dependenciaID int) bool {
	if dependenciaID <= 0 {
		return false
	}
	dependenciaMap, ok := parseDependenciaEvento(actividad["DependenciaId"])
	if !ok {
		return false
	}
	return dependenciaIncluyeProyecto(dependenciaMap, dependenciaID)
}

func siguienteNumeroExtension(idActividad string) (int, error) {
	var extensiones []map[string]interface{}
	url := beego.AppConfig.String("EventoService") + "calendario_evento_extension?query=CalendarioEventoId__Id:" + idActividad + "&sortby=NumeroExtension&order=desc&limit=1"
	if err := request.GetJson(url, &extensiones); err != nil {
		return 0, errors.New("error del servicio PostExtensionActividad: no fue posible calcular el consecutivo")
	}
	if len(extensiones) == 0 || len(extensiones[0]) == 0 {
		return 1, nil
	}
	numero, ok := interfaceToInt(extensiones[0]["NumeroExtension"])
	if !ok {
		return 1, nil
	}
	return numero + 1, nil
}

func inactivarVigenciaExtensionDependencia(idActividad string, dependenciaID int, usuario string) error {
	var relaciones []map[string]interface{}
	url := beego.AppConfig.String("EventoService") + "calendario_evento_extension_programa?query=Activo:true,Vigente:true,CalendarioEventoId__Id:" + idActividad + ",DependenciaId:" + strconv.Itoa(dependenciaID) + "&limit=0"
	if err := request.GetJson(url, &relaciones); err != nil {
		return errors.New("error del servicio PostExtensionActividad: no fue posible consultar vigencias previas")
	}
	for _, relacion := range relaciones {
		idRelacion, ok := idToString(relacion["Id"])
		if !ok {
			continue
		}
		anterior := deepCopyMap(relacion)
		relacion["Vigente"] = false
		var resultado map[string]interface{}
		if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento_extension_programa/"+idRelacion, "PUT", &resultado, relacion); err != nil || resultado == nil || resultado["Type"] == "error" {
			return errors.New("error del servicio PostExtensionActividad: no fue posible actualizar vigencia previa")
		}
		if id, err := strconv.Atoi(idRelacion); err == nil {
			RegistrarAuditoria("calendario_evento_extension_programa", id, "PUT", anterior, resultado, usuario, "PostExtensionActividad/vigencia-previa")
		}
	}
	return nil
}

func tieneExtensionVigenteDependencia(idActividad string, dependenciaID int) (bool, error) {
	var relaciones []map[string]interface{}
	url := beego.AppConfig.String("EventoService") + "calendario_evento_extension_programa?query=Activo:true,Vigente:true,CalendarioEventoId__Id:" + idActividad + ",DependenciaId:" + strconv.Itoa(dependenciaID) + "&limit=1"
	if err := request.GetJson(url, &relaciones); err != nil {
		return false, errors.New("error del servicio PutActividadDependencias: no fue posible consultar extensiones vigentes")
	}
	return len(relaciones) > 0 && len(relaciones[0]) > 0, nil
}

func extensionesActividad(idActividad string) ([]map[string]interface{}, error) {
	var extensiones []map[string]interface{}
	url := beego.AppConfig.String("EventoService") + "calendario_evento_extension?query=Activo:true,CalendarioEventoId__Id:" + idActividad + "&sortby=NumeroExtension&order=asc&limit=0"
	if err := request.GetJson(url, &extensiones); err != nil {
		return nil, errors.New("error consultando extensiones de actividad")
	}
	for i := range extensiones {
		idExtension, ok := idToString(extensiones[i]["Id"])
		if !ok {
			continue
		}
		var programas []map[string]interface{}
		urlProgramas := beego.AppConfig.String("EventoService") + "calendario_evento_extension_programa?query=Activo:true,CalendarioEventoExtensionId__Id:" + idExtension + "&limit=0"
		if err := request.GetJson(urlProgramas, &programas); err == nil {
			extensiones[i]["Programas"] = programas
		}
	}
	return extensiones, nil
}

func obtenerExtensionActividad(idActividad string, idExtension string) (map[string]interface{}, error) {
	var extension map[string]interface{}
	url := beego.AppConfig.String("EventoService") + "calendario_evento_extension/" + idExtension
	if err := request.GetJson(url, &extension); err != nil || extension == nil || extension["Type"] == "error" {
		return nil, errors.New("error consultando extensión de actividad")
	}
	actividad, ok := extension["CalendarioEventoId"].(map[string]interface{})
	actividadID, idOk := interfaceToInt(actividad["Id"])
	idActividadInt, err := strconv.Atoi(idActividad)
	if !ok || !idOk || err != nil || actividadID != idActividadInt {
		return nil, errors.New("error consultando extensión de actividad: la extensión no pertenece a la actividad")
	}
	if extension["Activo"] == false {
		return nil, errors.New("error consultando extensión de actividad: la extensión está inactiva")
	}
	return extension, nil
}

func relacionesExtensionActividad(idActividad string, idExtension string) ([]map[string]interface{}, error) {
	var relaciones []map[string]interface{}
	url := beego.AppConfig.String("EventoService") + "calendario_evento_extension_programa?query=Activo:true,CalendarioEventoId__Id:" + idActividad + ",CalendarioEventoExtensionId__Id:" + idExtension + "&limit=0"
	if err := request.GetJson(url, &relaciones); err != nil {
		return nil, errors.New("error consultando dependencias de la extensión")
	}
	return relaciones, nil
}

func reactivarRelacionPadre(idActividad string, relacion map[string]interface{}, padre interface{}, usuario string) {
	padreMap, ok := padre.(map[string]interface{})
	if !ok || padreMap == nil {
		return
	}
	padreID, ok := idToString(padreMap["Id"])
	if !ok {
		return
	}
	dependenciaID, ok := interfaceToInt(relacion["DependenciaId"])
	if !ok {
		return
	}
	var relacionesPadre []map[string]interface{}
	url := beego.AppConfig.String("EventoService") + "calendario_evento_extension_programa?query=Activo:true,Vigente:false,CalendarioEventoId__Id:" + idActividad + ",CalendarioEventoExtensionId__Id:" + padreID + ",DependenciaId:" + strconv.Itoa(dependenciaID) + "&limit=1"
	if err := request.GetJson(url, &relacionesPadre); err != nil || len(relacionesPadre) == 0 || len(relacionesPadre[0]) == 0 {
		return
	}
	idRelacionPadre, ok := idToString(relacionesPadre[0]["Id"])
	if !ok {
		return
	}
	anterior := deepCopyMap(relacionesPadre[0])
	relacionesPadre[0]["Vigente"] = true
	var resultado map[string]interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento_extension_programa/"+idRelacionPadre, "PUT", &resultado, relacionesPadre[0]); err == nil && resultado != nil && resultado["Type"] != "error" {
		if id, err := strconv.Atoi(idRelacionPadre); err == nil {
			RegistrarAuditoria("calendario_evento_extension_programa", id, "PUT", anterior, resultado, usuario, "DeleteExtensionActividad/reactivar-padre")
		}
	}
}

func idRelacionExtension(value interface{}) (int, bool) {
	return helpers.IDFromRelation(value)
}

func parseFechaExtension(fecha string) (time.Time, error) {
	return helpers.ParseFecha(fecha)
}

func fechaSalidaCalendario(value interface{}) time.Time {
	fecha := strings.TrimSpace(fmt.Sprintf("%v", value))
	if fecha == "" || fecha == "<nil>" {
		return time.Time{}
	}
	return fechaParaModeloSinError(fecha)
}

func fechaParaModelo(value interface{}) (time.Time, error) {
	return helpers.FechaTimeParaModelo(value)
}

func fechaParaModeloSinError(value interface{}) time.Time {
	fecha, err := fechaParaModelo(value)
	if err != nil {
		return time.Time{}
	}
	return fecha
}

func normalizarExtensionesSalida(extensiones []map[string]interface{}) {
	for i := range extensiones {
		normalizarExtensionSalida(extensiones[i])
		programas, ok := extensiones[i]["Programas"].([]map[string]interface{})
		if ok {
			for j := range programas {
				normalizarRelacionExtensionSalida(programas[j])
			}
			continue
		}
		programasInterface, ok := extensiones[i]["Programas"].([]interface{})
		if ok {
			for _, programa := range programasInterface {
				if relacion, ok := programa.(map[string]interface{}); ok {
					normalizarRelacionExtensionSalida(relacion)
				}
			}
		}
	}
}

func normalizarExtensionSalida(extension map[string]interface{}) {
	if extension == nil {
		return
	}
	for _, campo := range []string{"FechaFin", "FechaCreacion", "FechaModificacion"} {
		if value, ok := extension[campo]; ok {
			extension[campo] = fechaSalidaCalendario(value)
		}
	}
	if calendario, ok := extension["CalendarioEventoId"].(map[string]interface{}); ok {
		normalizarCalendarioEventoSalida(calendario)
	}
}

func normalizarRelacionExtensionSalida(relacion map[string]interface{}) {
	if relacion == nil {
		return
	}
	for _, campo := range []string{"FechaCreacion", "FechaModificacion"} {
		if value, ok := relacion[campo]; ok {
			relacion[campo] = fechaSalidaCalendario(value)
		}
	}
	if calendario, ok := relacion["CalendarioEventoId"].(map[string]interface{}); ok {
		normalizarCalendarioEventoSalida(calendario)
	}
	if extension, ok := relacion["CalendarioEventoExtensionId"].(map[string]interface{}); ok {
		normalizarExtensionSalida(extension)
	}
	if extensionPadre, ok := relacion["ExtensionPadreId"].(map[string]interface{}); ok {
		normalizarExtensionSalida(extensionPadre)
	}
}

func normalizarCalendarioEventoSalida(calendario map[string]interface{}) {
	if calendario == nil {
		return
	}
	for _, campo := range []string{"FechaInicio", "FechaFin", "FechaCreacion", "FechaModificacion"} {
		if value, ok := calendario[campo]; ok {
			calendario[campo] = fechaSalidaCalendario(value)
		}
	}
}

func documentoIdOrNil(value interface{}) interface{} {
	return helpers.DocumentoIDOrNil(value)
}

func validarRangoFechasDependencia(idActividad string, dependenciaID int, fechaInicio string, fechaFin string) error {
	actividad, err := obtenerActividad(idActividad)
	if err != nil {
		return err
	}
	rango, err := rangoPermitidoActividadDependencia(idActividad, dependenciaID, actividad)
	if err != nil {
		return err
	}
	inicio, err := parseFechaExtension(fechaInicio)
	if err != nil {
		return errors.New("fecha inicio inválida")
	}
	fin, err := parseFechaExtension(fechaFin)
	if err != nil {
		return errors.New("fecha fin inválida")
	}
	if fin.Before(inicio) {
		return errors.New("la fecha fin no puede ser menor que la fecha inicio")
	}
	inicioPermitido := rango.RangoPermitido.FechaInicio
	finPermitido := rango.RangoPermitido.FechaFin
	if inicio.Before(inicioPermitido) || fin.After(finPermitido) || fin.Before(inicio) {
		return errors.New("las fechas superan el rango autorizado para el programa seleccionado")
	}
	return nil
}

func inactivarExtensionesActividad(idActividad string, usuario string, endpoint string) error {
	extensiones, err := extensionesActividad(idActividad)
	if err != nil {
		return nil
	}
	for _, extension := range extensiones {
		idExtension, ok := idToString(extension["Id"])
		if !ok {
			continue
		}
		anterior := deepCopyMap(extension)
		extension["Activo"] = false
		var resultado map[string]interface{}
		if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento_extension/"+idExtension, "PUT", &resultado, extension); err == nil && resultado != nil && resultado["Type"] != "error" {
			if id, err := strconv.Atoi(idExtension); err == nil {
				RegistrarAuditoria("calendario_evento_extension", id, "PUT", anterior, resultado, usuario, endpoint+"/calendario_evento_extension")
			}
		}
	}
	var relaciones []map[string]interface{}
	url := beego.AppConfig.String("EventoService") + "calendario_evento_extension_programa?query=Activo:true,CalendarioEventoId__Id:" + idActividad + "&limit=0"
	if err := request.GetJson(url, &relaciones); err != nil {
		return nil
	}
	for _, relacion := range relaciones {
		idRelacion, ok := idToString(relacion["Id"])
		if !ok {
			continue
		}
		anterior := deepCopyMap(relacion)
		relacion["Activo"] = false
		relacion["Vigente"] = false
		var resultado map[string]interface{}
		if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento_extension_programa/"+idRelacion, "PUT", &resultado, relacion); err == nil && resultado != nil && resultado["Type"] != "error" {
			if id, err := strconv.Atoi(idRelacion); err == nil {
				RegistrarAuditoria("calendario_evento_extension_programa", id, "PUT", anterior, resultado, usuario, endpoint+"/calendario_evento_extension_programa")
			}
		}
	}
	return nil
}

func extensionResumenActividad(idActividad string) []map[string]interface{} {
	extensiones, err := extensionesActividad(idActividad)
	if err != nil {
		return []map[string]interface{}{}
	}
	normalizarExtensionesSalida(extensiones)
	return extensiones
}

var _ = fmt.Sprintf
