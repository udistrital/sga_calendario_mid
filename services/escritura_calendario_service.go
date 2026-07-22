package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/udistrital/sga_calendario_mid/helpers"
	"github.com/udistrital/sga_calendario_mid/models"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
)

type ImpactoDesasociacionCalendarioError struct {
	Impactos []map[string]interface{}
}

func (e *ImpactoDesasociacionCalendarioError) Error() string {
	return "No se pueden desasociar programas académicos con fechas particulares modificadas o extensiones vigentes."
}

func (e *ImpactoDesasociacionCalendarioError) Data() map[string]interface{} {
	return map[string]interface{}{
		"RequiereConfirmacion": false,
		"Bloqueante":           true,
		"Impactos":             e.Impactos,
	}
}

func PutCalendarioEstado(id string, data []byte, usuario string) (interface{}, error) {
	var recibido models.EstadoActivoRequest
	if err := json.Unmarshal(data, &recibido); err != nil {
		return nil, errors.New("error del servicio PutCalendarioEstado: solicitud inválida")
	}

	var calendario map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario/"+id, &calendario); err != nil || calendario == nil || calendario["Type"] == "error" {
		return nil, errors.New("error del servicio PutCalendarioEstado: no fue posible consultar el calendario")
	}

	anterior := deepCopyMap(calendario)

	activo := false
	if recibido.Activo == nil {
		activoActual, ok := calendario["Activo"].(bool)
		if !ok {
			return nil, errors.New("error del servicio PutCalendarioEstado: estado inválido")
		}
		activo = !activoActual
	} else {
		activo = *recibido.Activo
	}
	calendario["Activo"] = activo

	var resultado map[string]interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario/"+id, "PUT", &resultado, calendario); err != nil || resultado == nil || resultado["Type"] == "error" {
		return nil, errors.New("error del servicio PutCalendarioEstado: no fue posible actualizar el calendario")
	}

	entidadId, _ := strconv.Atoi(id)
	RegistrarAuditoria("calendario", entidadId, "PUT", anterior, resultado, usuario, "PutCalendarioEstado")
	if !activo {
		if err := inactivarProcesosCalendario(id, usuario, "PutCalendarioEstado"); err != nil {
			logs.Error(err)
		}
		if err := inactivarEventosCalendario(id, usuario, "PutCalendarioEstado"); err != nil {
			return nil, err
		}
	}

	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func PutCalendarioDependencias(id string, data []byte, usuario string, authHeader string) (interface{}, error) {
	var recibido models.CalendarioDependenciasRequest
	if err := json.Unmarshal(data, &recibido); err != nil {
		return nil, errors.New("error del servicio PutCalendarioDependencias: solicitud inválida")
	}
	var calendario map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario/"+id, &calendario); err != nil || calendario == nil || calendario["Type"] == "error" {
		return nil, errors.New("error del servicio PutCalendarioDependencias: no fue posible consultar el calendario")
	}

	anterior := deepCopyMap(calendario)

	dependenciaId := recibido.DependenciaId
	if dependenciaId == "" {
		return nil, errors.New("error del servicio PutCalendarioDependencias: dependencia inválida")
	}
	if err := helpers.ValidarFechasDependenciaEvento(dependenciaId); err != nil {
		return nil, err
	}
	impactos, eventos, errImpacto := impactosDesasociacionCalendario(id, calendario, dependenciaId)
	if errImpacto != nil {
		return nil, errImpacto
	}
	if len(impactos) > 0 {
		return nil, &ImpactoDesasociacionCalendarioError{Impactos: impactos}
	}
	calendario["DependenciaId"] = dependenciaId

	var resultado map[string]interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario/"+id, "PUT", &resultado, calendario); err != nil || resultado == nil || resultado["Type"] == "error" {
		return nil, errors.New("error del servicio PutCalendarioDependencias: no fue posible actualizar el calendario")
	}

	entidadId, _ := strconv.Atoi(id)
	RegistrarAuditoria("calendario", entidadId, "PUT", anterior, resultado, usuario, "PutCalendarioDependencias")
	if err := cascadaDependenciasCalendario(eventos, dependenciaId, usuario); err != nil {
		return nil, err
	}

	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func impactosDesasociacionCalendario(idCalendario string, calendario map[string]interface{}, dependenciaNueva string) ([]map[string]interface{}, []map[string]interface{}, error) {
	dependenciaAnterior, okAnterior := parseDependenciaEvento(calendario["DependenciaId"])
	dependenciaNuevaMap, okNueva := parseDependenciaEvento(dependenciaNueva)
	if !okAnterior {
		return []map[string]interface{}{}, []map[string]interface{}{}, nil
	}
	if !okNueva {
		dependenciaNuevaMap = map[string]interface{}{"proyectos": []interface{}{}, "fechas": []interface{}{}}
	}
	removidos := programasRemovidos(dependenciaAnterior, dependenciaNuevaMap)
	if len(removidos) == 0 {
		return []map[string]interface{}{}, []map[string]interface{}{}, nil
	}

	var eventos []map[string]interface{}
	urlEventos := beego.AppConfig.String("EventoService") + "calendario_evento?query=Activo:true,ProcesoId__CalendarioID__Id:" + idCalendario + "&limit=0"
	if err := request.GetJson(urlEventos, &eventos); err != nil {
		return nil, nil, errors.New("error del servicio PutCalendarioDependencias: no fue posible consultar actividades del calendario")
	}

	impactos := make([]map[string]interface{}, 0)
	for _, actividad := range eventos {
		dependenciaActividad, ok := parseDependenciaEvento(actividad["DependenciaId"])
		if !ok {
			continue
		}
		idActividad, idOk := idToString(actividad["Id"])
		if !idOk {
			continue
		}
		for _, programaID := range removidos {
			if !dependenciaIncluyeProyecto(dependenciaActividad, programaID) {
				continue
			}
			fechaParticular := tieneFechaParticularModificada(actividad, dependenciaActividad, programaID)
			extensionVigente, err := tieneExtensionVigenteDependencia(idActividad, programaID)
			if err != nil {
				return nil, nil, err
			}
			if fechaParticular || extensionVigente {
				impactos = append(impactos, map[string]interface{}{
					"ProgramaId":               programaID,
					"ActividadId":              idActividad,
					"Actividad":                nombreActividadImpacto(actividad),
					"FechaParticular":          fechaParticular,
					"ExtensionVigente":         extensionVigente,
					"ImplicaRetiroEnCascada":   true,
					"ImplicaAnularExtensiones": extensionVigente,
				})
			}
		}
	}
	return impactos, eventos, nil
}

func programasRemovidos(dependenciaAnterior map[string]interface{}, dependenciaNueva map[string]interface{}) []int {
	programasAnteriores := proyectosDependencia(dependenciaAnterior)
	programasNuevos := proyectosDependencia(dependenciaNueva)
	removidos := make([]int, 0)
	for programaID := range programasAnteriores {
		if !programasNuevos[programaID] {
			removidos = append(removidos, programaID)
		}
	}
	return removidos
}

func cascadaDependenciasCalendario(eventos []map[string]interface{}, dependenciaNueva string, usuario string) error {
	dependenciaNuevaMap, okNueva := parseDependenciaEvento(dependenciaNueva)
	if !okNueva {
		dependenciaNuevaMap = map[string]interface{}{"proyectos": []interface{}{}, "fechas": []interface{}{}}
	}
	programasCalendario := proyectosDependencia(dependenciaNuevaMap)
	for _, actividad := range eventos {
		dependenciaActividad, ok := parseDependenciaEvento(actividad["DependenciaId"])
		if !ok {
			continue
		}
		dependenciaActualizada, removidos, cambio := filtrarDependenciaActividadPorCalendario(dependenciaActividad, programasCalendario)
		if !cambio {
			continue
		}
		idActividad, idOk := idToString(actividad["Id"])
		if !idOk {
			continue
		}
		for _, programaID := range removidos {
			if err := inactivarExtensionesDependenciaRetirada(idActividad, programaID, usuario); err != nil {
				return err
			}
		}
		anterior := deepCopyMap(actividad)
		dependenciaBytes, _ := json.Marshal(dependenciaActualizada)
		actividad["DependenciaId"] = string(dependenciaBytes)
		var resultado map[string]interface{}
		if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento/"+idActividad, "PUT", &resultado, actividad); err != nil || resultado == nil || resultado["Type"] == "error" {
			return errors.New("error del servicio PutCalendarioDependencias: no fue posible actualizar actividades en cascada")
		}
		if entidadId, err := strconv.Atoi(idActividad); err == nil {
			RegistrarAuditoria("calendario_evento", entidadId, "PUT", anterior, resultado, usuario, "PutCalendarioDependencias/cascada")
		}
	}
	return nil
}

func filtrarDependenciaActividadPorCalendario(dependenciaActividad map[string]interface{}, programasCalendario map[int]bool) (map[string]interface{}, []int, bool) {
	resultado := deepCopyMap(dependenciaActividad)
	proyectosRaw, _ := dependenciaActividad["proyectos"].([]interface{})
	proyectosActualizados := make([]interface{}, 0)
	removidos := make([]int, 0)
	for _, proyecto := range proyectosRaw {
		programaID, ok := interfaceToInt(proyecto)
		if !ok {
			continue
		}
		if programasCalendario[programaID] {
			proyectosActualizados = append(proyectosActualizados, programaID)
		} else {
			removidos = append(removidos, programaID)
		}
	}
	if len(removidos) == 0 {
		return resultado, removidos, false
	}
	fechasActualizadas := make([]interface{}, 0)
	if fechasRaw, ok := dependenciaActividad["fechas"].([]interface{}); ok {
		for _, fecha := range fechasRaw {
			fechaMap, ok := fecha.(map[string]interface{})
			if !ok {
				continue
			}
			programaID, ok := interfaceToInt(fechaMap["Id"])
			if ok && programasCalendario[programaID] {
				fechasActualizadas = append(fechasActualizadas, fechaMap)
			}
		}
	}
	resultado["proyectos"] = proyectosActualizados
	resultado["fechas"] = fechasActualizadas
	return resultado, removidos, true
}

func inactivarExtensionesDependenciaRetirada(idActividad string, dependenciaID int, usuario string) error {
	var relaciones []map[string]interface{}
	url := beego.AppConfig.String("EventoService") + "calendario_evento_extension_programa?query=Activo:true,Vigente:true,CalendarioEventoId__Id:" + idActividad + ",DependenciaId:" + strconv.Itoa(dependenciaID) + "&limit=0"
	if err := request.GetJson(url, &relaciones); err != nil {
		return errors.New("error del servicio PutCalendarioDependencias: no fue posible consultar extensiones vigentes")
	}
	for _, relacion := range relaciones {
		idRelacion, ok := idToString(relacion["Id"])
		if !ok {
			continue
		}
		anteriorRelacion := deepCopyMap(relacion)
		relacion["Activo"] = false
		relacion["Vigente"] = false
		var resultadoRelacion map[string]interface{}
		if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento_extension_programa/"+idRelacion, "PUT", &resultadoRelacion, relacion); err != nil || resultadoRelacion == nil || resultadoRelacion["Type"] == "error" {
			return errors.New("error del servicio PutCalendarioDependencias: no fue posible inactivar extensión por programa")
		}
		if id, err := strconv.Atoi(idRelacion); err == nil {
			RegistrarAuditoria("calendario_evento_extension_programa", id, "PUT", anteriorRelacion, resultadoRelacion, usuario, "PutCalendarioDependencias/cascada-extension-programa")
		}
		if extension, ok := relacion["CalendarioEventoExtensionId"].(map[string]interface{}); ok {
			if extensionID, ok := interfaceToInt(extension["Id"]); ok {
				if err := inactivarExtensionSinProgramasVigentes(idActividad, extensionID, usuario); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func inactivarExtensionSinProgramasVigentes(idActividad string, extensionID int, usuario string) error {
	var relacionesVigentes []map[string]interface{}
	url := beego.AppConfig.String("EventoService") + "calendario_evento_extension_programa?query=Activo:true,Vigente:true,CalendarioEventoId__Id:" + idActividad + ",CalendarioEventoExtensionId__Id:" + strconv.Itoa(extensionID) + "&limit=1"
	if err := request.GetJson(url, &relacionesVigentes); err != nil || (len(relacionesVigentes) > 0 && len(relacionesVigentes[0]) > 0) {
		return nil
	}
	var extension map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento_extension/"+strconv.Itoa(extensionID), &extension); err != nil || extension == nil || extension["Type"] == "error" {
		return errors.New("error del servicio PutCalendarioDependencias: no fue posible consultar extensión para cascada")
	}
	anterior := deepCopyMap(extension)
	extension["Activo"] = false
	var resultado map[string]interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento_extension/"+strconv.Itoa(extensionID), "PUT", &resultado, extension); err != nil || resultado == nil || resultado["Type"] == "error" {
		return errors.New("error del servicio PutCalendarioDependencias: no fue posible inactivar extensión sin programas vigentes")
	}
	RegistrarAuditoria("calendario_evento_extension", extensionID, "PUT", anterior, resultado, usuario, "PutCalendarioDependencias/cascada-extension")
	return nil
}

func nombreActividadImpacto(actividad map[string]interface{}) string {
	if nombre, ok := actividad["Nombre"].(string); ok && nombre != "" {
		return nombre
	}
	if catalogo, ok := actividad["EventoCatalogoId"].(map[string]interface{}); ok {
		if nombre, ok := catalogo["Nombre"].(string); ok && nombre != "" {
			return nombre
		}
		if descripcion, ok := catalogo["Descripcion"].(string); ok && descripcion != "" {
			return descripcion
		}
	}
	if id, ok := idToString(actividad["Id"]); ok {
		return "Actividad " + id
	}
	return "Actividad"
}

func ProgramasRemovidosCalendario(dependenciaAnterior map[string]interface{}, dependenciaNueva map[string]interface{}) []int {
	return programasRemovidos(dependenciaAnterior, dependenciaNueva)
}

func FiltrarDependenciaActividadPorCalendario(dependenciaActividad map[string]interface{}, programasCalendario map[int]bool) (map[string]interface{}, []int, bool) {
	return filtrarDependenciaActividadPorCalendario(dependenciaActividad, programasCalendario)
}

func TieneFechaParticularModificada(actividad map[string]interface{}, dependenciaMap map[string]interface{}, programaID int) bool {
	return tieneFechaParticularModificada(actividad, dependenciaMap, programaID)
}

func NombreActividadImpacto(actividad map[string]interface{}) string {
	return nombreActividadImpacto(actividad)
}

func PostProcesoCalendario(data []byte, usuario string) (interface{}, error) {
	var proceso map[string]interface{}
	var resultado map[string]interface{}
	if err := json.Unmarshal(data, &proceso); err != nil {
		return nil, errors.New("error del servicio PostProcesoCalendario: solicitud inválida")
	}
	if err := validarFechasPayload(proceso, nil, "proceso"); err != nil {
		return nil, err
	}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+"proceso", "POST", &resultado, proceso); err != nil || resultado == nil || resultado["Type"] == "error" {
		return nil, errors.New("error del servicio PostProcesoCalendario: no fue posible crear el proceso")
	}

	if id, ok := resultado["Id"].(float64); ok {
		RegistrarAuditoria("proceso", int(id), "POST", nil, resultado, usuario, "PostProcesoCalendario")
	}

	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func PutProcesoPeriodicidad(id string, data []byte, usuario string) (interface{}, error) {
	var recibido models.ProcesoPeriodicidadRequest
	if err := json.Unmarshal(data, &recibido); err != nil {
		return nil, errors.New("error del servicio PutProcesoPeriodicidad: solicitud inválida")
	}

	var proceso map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"proceso/"+id, &proceso); err != nil || proceso == nil || proceso["Type"] == "error" {
		return nil, errors.New("error del servicio PutProcesoPeriodicidad: no fue posible consultar el proceso")
	}

	anterior := deepCopyMap(proceso)

	if recibido.TipoRecurrenciaId.Id <= 0 {
		return nil, errors.New("error del servicio PutProcesoPeriodicidad: periodicidad inválida")
	}
	proceso["TipoRecurrenciaId"] = recibido.TipoRecurrenciaId

	var resultado map[string]interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+"proceso/"+id, "PUT", &resultado, proceso); err != nil || resultado == nil || resultado["Type"] == "error" {
		return nil, errors.New("error del servicio PutProcesoPeriodicidad: no fue posible actualizar el proceso")
	}

	entidadId, _ := strconv.Atoi(id)
	RegistrarAuditoria("proceso", entidadId, "PUT", anterior, resultado, usuario, "PutProcesoPeriodicidad")

	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func PutProcesoEstado(id string, data []byte, usuario string) (interface{}, error) {
	var recibido models.EstadoActivoRequest
	if err := json.Unmarshal(data, &recibido); err != nil {
		return nil, errors.New("error del servicio PutProcesoEstado: solicitud inválida")
	}

	var proceso map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"proceso/"+id, &proceso); err != nil || proceso == nil || proceso["Type"] == "error" {
		return nil, errors.New("error del servicio PutProcesoEstado: no fue posible consultar el proceso")
	}

	anterior := deepCopyMap(proceso)

	activo := false
	if recibido.Activo == nil {
		activoActual, ok := proceso["Activo"].(bool)
		if !ok {
			return nil, errors.New("error del servicio PutProcesoEstado: estado inválido")
		}
		activo = !activoActual
	} else {
		activo = *recibido.Activo
	}
	proceso["Activo"] = activo
	if activo && !calendarioActivoDeProceso(proceso) {
		return nil, errors.New("no se puede activar el proceso porque el calendario está inactivo")
	}

	var resultado map[string]interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+"proceso/"+id, "PUT", &resultado, proceso); err != nil || resultado == nil || resultado["Type"] == "error" {
		return nil, errors.New("error del servicio PutProcesoEstado: no fue posible actualizar el proceso")
	}

	entidadId, _ := strconv.Atoi(id)
	RegistrarAuditoria("proceso", entidadId, "PUT", anterior, resultado, usuario, "PutProcesoEstado")
	if !activo {
		if err := inactivarEventosProceso(id, usuario, "PutProcesoEstado"); err != nil {
			return nil, err
		}
	}

	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func inactivarProcesosCalendario(idCalendario string, usuario string, endpoint string) error {
	var procesos []map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"proceso?query=Activo:true,CalendarioID__Id:"+idCalendario+"&limit=0", &procesos); err != nil {
		return errors.New("error inactivando procesos del calendario")
	}
	for _, proceso := range procesos {
		idProceso, ok := idToString(proceso["Id"])
		if !ok {
			continue
		}
		anterior := deepCopyMap(proceso)
		proceso["Activo"] = false

		var resultado map[string]interface{}
		if err := request.SendJson(beego.AppConfig.String("EventoService")+"proceso/"+idProceso, "PUT", &resultado, proceso); err != nil || resultado == nil || resultado["Type"] == "error" {
			return errors.New("error inactivando proceso del calendario")
		}
		if id, err := strconv.Atoi(idProceso); err == nil {
			RegistrarAuditoria("proceso", id, "PUT", anterior, resultado, usuario, endpoint+"/proceso")
		}
		if err := inactivarEventosProceso(idProceso, usuario, endpoint); err != nil {
			return err
		}
	}
	return nil
}

func calendarioActivoDeProceso(proceso map[string]interface{}) bool {
	calendario, ok := proceso["CalendarioID"].(map[string]interface{})
	if !ok || calendario == nil {
		return false
	}
	idCalendario, ok := idToString(calendario["Id"])
	if !ok {
		return false
	}
	var calendarioActual map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario/"+idCalendario, &calendarioActual); err != nil || calendarioActual == nil || calendarioActual["Type"] == "error" {
		return false
	}
	activo, ok := calendarioActual["Activo"].(bool)
	return ok && activo
}

func procesoActivoYCalendarioActivoDeEvento(evento map[string]interface{}) bool {
	proceso, ok := evento["ProcesoId"].(map[string]interface{})
	if !ok || proceso == nil {
		return false
	}
	idProceso, ok := idToString(proceso["Id"])
	if !ok {
		return false
	}
	var procesoActual map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"proceso/"+idProceso, &procesoActual); err != nil || procesoActual == nil || procesoActual["Type"] == "error" {
		return false
	}
	activoProceso, ok := procesoActual["Activo"].(bool)
	if !ok || !activoProceso {
		return false
	}
	return calendarioActivoDeProceso(procesoActual)
}

func validarActivacionProceso(id string) error {
	var proceso map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"proceso/"+id, &proceso); err != nil || proceso == nil || proceso["Type"] == "error" {
		return errors.New("no fue posible validar el proceso")
	}
	if !calendarioActivoDeProceso(proceso) {
		return errors.New("no se puede activar el proceso porque el calendario está inactivo")
	}
	return nil
}

func validarActivacionEvento(id string) error {
	var evento map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento/"+id, &evento); err != nil || evento == nil || evento["Type"] == "error" {
		return errors.New("no fue posible validar el evento")
	}
	if !procesoActivoYCalendarioActivoDeEvento(evento) {
		return errors.New("no se puede activar el evento porque el proceso o calendario está inactivo")
	}
	return nil
}

func inactivarEventosProceso(idProceso string, usuario string, endpoint string) error {
	var eventos []map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento?query=Activo:true,ProcesoId__Id:"+idProceso+"&limit=0", &eventos); err != nil {
		return errors.New("error inactivando eventos del proceso")
	}
	for _, evento := range eventos {
		if err := inactivarEvento(evento, usuario, endpoint); err != nil {
			return err
		}
	}
	return nil
}

func inactivarEventosCalendario(idCalendario string, usuario string, endpoint string) error {
	var eventos []map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento?query=Activo:true,ProcesoId__CalendarioID__Id:"+idCalendario+"&limit=0", &eventos); err != nil {
		return errors.New("error inactivando eventos del calendario")
	}
	for _, evento := range eventos {
		if err := inactivarEvento(evento, usuario, endpoint); err != nil {
			return err
		}
	}
	return nil
}

func inactivarEvento(evento map[string]interface{}, usuario string, endpoint string) error {
	idEvento, ok := idToString(evento["Id"])
	if !ok {
		return nil
	}
	anterior := deepCopyMap(evento)
	evento["Activo"] = false

	var resultado map[string]interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento/"+idEvento, "PUT", &resultado, evento); err != nil || resultado == nil || resultado["Type"] == "error" {
		return errors.New("error inactivando evento del proceso")
	}
	if id, err := strconv.Atoi(idEvento); err == nil {
		RegistrarAuditoria("calendario_evento", id, "PUT", anterior, resultado, usuario, endpoint+"/calendario_evento")
	}
	if err := inactivarExtensionesActividad(idEvento, usuario, endpoint); err != nil {
		logs.Error(err)
	}
	return nil
}

func idToString(value interface{}) (string, bool) {
	return helpers.IDToString(value)
}

func PutActividadDependencias(id string, data []byte, usuario string) (interface{}, error) {
	var recibido models.ActividadDependenciasRequest
	if err := json.Unmarshal(data, &recibido); err != nil {
		return nil, errors.New("error del servicio PutActividadDependencias: solicitud inválida")
	}

	var actividad map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento/"+id, &actividad); err != nil || actividad == nil || actividad["Type"] == "error" {
		return nil, errors.New("error del servicio PutActividadDependencias: no fue posible consultar la actividad")
	}

	anterior := deepCopyMap(actividad)

	dependenciaId := recibido.DependenciaId
	if dependenciaId == "" {
		return nil, errors.New("error del servicio PutActividadDependencias: dependencia inválida")
	}
	if err := helpers.ValidarFechasDependenciaEvento(dependenciaId); err != nil {
		return nil, err
	}
	if err := validarDesasociacionActividad(id, actividad, dependenciaId); err != nil {
		return nil, err
	}
	actividad["DependenciaId"] = dependenciaId

	var resultado map[string]interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento/"+id, "PUT", &resultado, actividad); err != nil || resultado == nil || resultado["Type"] == "error" {
		return nil, errors.New("error del servicio PutActividadDependencias: no fue posible actualizar la actividad")
	}

	entidadId, _ := strconv.Atoi(id)
	RegistrarAuditoria("calendario_evento", entidadId, "PUT", anterior, resultado, usuario, "PutActividadDependencias")

	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func PostActividadesProgramasMasivo(idCalendario string, data []byte, usuario string) (interface{}, error) {
	var recibido models.ActividadesProgramasMasivoRequest
	if err := json.Unmarshal(data, &recibido); err != nil {
		return nil, errors.New("error del servicio PostActividadesProgramasMasivo: solicitud inválida")
	}

	operacion := strings.ToLower(strings.TrimSpace(recibido.Operacion))
	if operacion != "asociar" && operacion != "desasociar" {
		return nil, errors.New("error del servicio PostActividadesProgramasMasivo: operación inválida")
	}
	programas := intSliceUnicos(recibido.ProgramaIds)
	actividades := intSliceUnicos(recibido.ActividadIds)
	if len(programas) == 0 || len(actividades) == 0 {
		return nil, errors.New("error del servicio PostActividadesProgramasMasivo: debe seleccionar programas y actividades")
	}

	var calendario map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario/"+idCalendario, &calendario); err != nil || calendario == nil || calendario["Type"] == "error" {
		return nil, errors.New("error del servicio PostActividadesProgramasMasivo: no fue posible consultar el calendario")
	}
	dependenciaCalendario, ok := parseDependenciaEvento(calendario["DependenciaId"])
	if !ok {
		return nil, errors.New("error del servicio PostActividadesProgramasMasivo: calendario sin programas académicos válidos")
	}
	programasCalendario := proyectosDependencia(dependenciaCalendario)
	for _, programaID := range programas {
		if !programasCalendario[programaID] {
			return nil, errors.New("error del servicio PostActividadesProgramasMasivo: el programa académico " + strconv.Itoa(programaID) + " no está asociado al calendario")
		}
	}

	procesadas := 0
	actualizadas := 0
	sinCambios := 0
	impactos := make([]map[string]interface{}, 0)
	resultados := make([]map[string]interface{}, 0)
	pendientes := make([]map[string]interface{}, 0)
	for _, actividadID := range actividades {
		idActividad := strconv.Itoa(actividadID)
		var actividad map[string]interface{}
		if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento/"+idActividad, &actividad); err != nil || actividad == nil || actividad["Type"] == "error" {
			return nil, errors.New("error del servicio PostActividadesProgramasMasivo: no fue posible consultar la actividad " + idActividad)
		}
		pertenece, errPertenece := actividadPerteneceACalendario(actividad, idCalendario)
		if errPertenece != nil {
			return nil, errPertenece
		}
		if !pertenece {
			return nil, errors.New("error del servicio PostActividadesProgramasMasivo: la actividad " + idActividad + " no pertenece al calendario")
		}
		if activo, ok := actividad["Activo"].(bool); ok && !activo {
			impactos = append(impactos, map[string]interface{}{
				"ActividadId":       idActividad,
				"Actividad":         nombreActividadImpacto(actividad),
				"ActividadInactiva": true,
			})
			continue
		}

		dependenciaActividad, ok := parseDependenciaEvento(actividad["DependenciaId"])
		if !ok {
			dependenciaActividad = map[string]interface{}{"proyectos": []interface{}{}, "fechas": []interface{}{}}
		}
		dependenciaActualizada, cambio, impactosActividad, err := dependenciaActividadMasiva(actividad, dependenciaActividad, idActividad, programas, operacion)
		if err != nil {
			return nil, err
		}
		impactos = append(impactos, impactosActividad...)
		procesadas++
		if !cambio {
			sinCambios++
			continue
		}
		pendientes = append(pendientes, map[string]interface{}{
			"ActividadId":   actividadID,
			"IdActividad":   idActividad,
			"Actividad":     actividad,
			"DependenciaId": dependenciaActualizada,
		})
	}

	if len(impactos) > 0 {
		return nil, &ImpactoDesasociacionCalendarioError{Impactos: impactos}
	}

	for _, pendiente := range pendientes {
		actividadID := pendiente["ActividadId"].(int)
		idActividad := pendiente["IdActividad"].(string)
		actividad := pendiente["Actividad"].(map[string]interface{})
		dependenciaActualizada := pendiente["DependenciaId"].(map[string]interface{})
		anterior := deepCopyMap(actividad)
		dependenciaBytes, _ := json.Marshal(dependenciaActualizada)
		actividad["DependenciaId"] = string(dependenciaBytes)
		var resultado map[string]interface{}
		if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento/"+idActividad, "PUT", &resultado, actividad); err != nil || resultado == nil || resultado["Type"] == "error" {
			return nil, errors.New("error del servicio PostActividadesProgramasMasivo: no fue posible actualizar la actividad " + idActividad)
		}
		RegistrarAuditoria("calendario_evento", actividadID, "PUT", anterior, resultado, usuario, "PostActividadesProgramasMasivo/"+operacion)
		actualizadas++
		resultados = append(resultados, map[string]interface{}{"ActividadId": actividadID, "DependenciaId": dependenciaActualizada})
	}

	return requestresponse.APIResponseDTO(true, 200, map[string]interface{}{
		"Procesadas":   procesadas,
		"Actualizadas": actualizadas,
		"SinCambios":   sinCambios,
		"Operacion":    operacion,
		"Resultados":   resultados,
	}), nil
}

func PostValidarActividadesProgramasMasivo(idCalendario string, data []byte) (interface{}, error) {
	var recibido models.ActividadesProgramasMasivoRequest
	if err := json.Unmarshal(data, &recibido); err != nil {
		return nil, errors.New("error del servicio PostValidarActividadesProgramasMasivo: solicitud inválida")
	}
	programas := intSliceUnicos(recibido.ProgramaIds)
	actividades := intSliceUnicos(recibido.ActividadIds)
	if len(programas) == 0 {
		return nil, errors.New("error del servicio PostValidarActividadesProgramasMasivo: debe seleccionar programas")
	}

	var calendario map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario/"+idCalendario, &calendario); err != nil || calendario == nil || calendario["Type"] == "error" {
		return nil, errors.New("error del servicio PostValidarActividadesProgramasMasivo: no fue posible consultar el calendario")
	}
	dependenciaCalendario, ok := parseDependenciaEvento(calendario["DependenciaId"])
	if !ok {
		return nil, errors.New("error del servicio PostValidarActividadesProgramasMasivo: calendario sin programas académicos válidos")
	}
	programasCalendario := proyectosDependencia(dependenciaCalendario)
	programasInvalidos := make([]int, 0)
	for _, programaID := range programas {
		if !programasCalendario[programaID] {
			programasInvalidos = append(programasInvalidos, programaID)
		}
	}
	if len(programasInvalidos) > 0 {
		return nil, errors.New("error del servicio PostValidarActividadesProgramasMasivo: hay programas no asociados al calendario")
	}

	actividadesConsultadas, err := consultarActividadesMasivas(idCalendario, actividades)
	if err != nil {
		return nil, err
	}
	validaciones := make([]map[string]interface{}, 0)
	pendientesAsociar := 0
	disponiblesDesasociar := 0
	bloqueadas := 0
	for _, actividad := range actividadesConsultadas {
		validacion, err := validarActividadProgramasMasivo(idCalendario, actividad, programas)
		if err != nil {
			return nil, err
		}
		if puede, _ := validacion["PuedeAsociar"].(bool); puede {
			pendientesAsociar++
		}
		if puede, _ := validacion["PuedeDesasociar"].(bool); puede {
			disponiblesDesasociar++
		}
		if bloqueos, ok := validacion["Bloqueos"].([]map[string]interface{}); ok && len(bloqueos) > 0 {
			bloqueadas++
		}
		validaciones = append(validaciones, validacion)
	}

	return requestresponse.APIResponseDTO(true, 200, map[string]interface{}{
		"ProgramaIds":             programas,
		"Actividades":             validaciones,
		"PendientesAsociar":       pendientesAsociar,
		"DisponiblesDesasociar":   disponiblesDesasociar,
		"BloqueadasDesasociacion": bloqueadas,
	}), nil
}

func consultarActividadesMasivas(idCalendario string, actividades []int) ([]map[string]interface{}, error) {
	if len(actividades) == 0 {
		var eventos []map[string]interface{}
		urlEventos := beego.AppConfig.String("EventoService") + "calendario_evento?query=Activo:true,ProcesoId__CalendarioID__Id:" + idCalendario + "&limit=0"
		if err := request.GetJson(urlEventos, &eventos); err != nil {
			return nil, errors.New("error del servicio PostValidarActividadesProgramasMasivo: no fue posible consultar actividades del calendario")
		}
		return eventos, nil
	}
	resultado := make([]map[string]interface{}, 0, len(actividades))
	for _, actividadID := range actividades {
		idActividad := strconv.Itoa(actividadID)
		var actividad map[string]interface{}
		if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento/"+idActividad, &actividad); err != nil || actividad == nil || actividad["Type"] == "error" {
			return nil, errors.New("error del servicio PostValidarActividadesProgramasMasivo: no fue posible consultar la actividad " + idActividad)
		}
		pertenece, errPertenece := actividadPerteneceACalendario(actividad, idCalendario)
		if errPertenece != nil {
			return nil, errPertenece
		}
		if !pertenece {
			return nil, errors.New("error del servicio PostValidarActividadesProgramasMasivo: la actividad " + idActividad + " no pertenece al calendario")
		}
		resultado = append(resultado, actividad)
	}
	return resultado, nil
}

func validarActividadProgramasMasivo(idCalendario string, actividad map[string]interface{}, programas []int) (map[string]interface{}, error) {
	_ = idCalendario
	idActividad, _ := idToString(actividad["Id"])
	dependenciaActividad, ok := parseDependenciaEvento(actividad["DependenciaId"])
	if !ok {
		dependenciaActividad = map[string]interface{}{"proyectos": []interface{}{}, "fechas": []interface{}{}}
	}
	programasActuales := proyectosDependencia(dependenciaActividad)
	asociados := make([]int, 0)
	faltantes := make([]int, 0)
	bloqueos := make([]map[string]interface{}, 0)
	activo, okActivo := actividad["Activo"].(bool)
	if !okActivo {
		activo = true
	}
	if !activo {
		bloqueos = append(bloqueos, map[string]interface{}{
			"ActividadId":       idActividad,
			"ActividadInactiva": true,
			"Motivo":            "Actividad inactiva",
		})
	}
	for _, programaID := range programas {
		if programasActuales[programaID] {
			asociados = append(asociados, programaID)
			fechaParticular := tieneFechaParticularModificada(actividad, dependenciaActividad, programaID)
			extensionVigente, err := tieneExtensionVigenteDependencia(idActividad, programaID)
			if err != nil {
				return nil, err
			}
			if fechaParticular || extensionVigente {
				bloqueos = append(bloqueos, map[string]interface{}{
					"ProgramaId":       programaID,
					"ActividadId":      idActividad,
					"FechaParticular":  fechaParticular,
					"ExtensionVigente": extensionVigente,
				})
			}
		} else {
			faltantes = append(faltantes, programaID)
		}
	}
	return map[string]interface{}{
		"ActividadId":        idActividad,
		"Nombre":             nombreActividadImpacto(actividad),
		"Descripcion":        descripcionActividadImpacto(actividad),
		"Activo":             activo,
		"Asociados":          asociados,
		"Faltantes":          faltantes,
		"Bloqueos":           bloqueos,
		"PuedeAsociar":       activo && len(faltantes) > 0,
		"PuedeDesasociar":    activo && len(asociados) > 0 && len(bloqueos) == 0,
		"TotalSeleccionados": len(programas),
	}, nil
}

func descripcionActividadImpacto(actividad map[string]interface{}) string {
	if descripcion, ok := actividad["Descripcion"].(string); ok && descripcion != "" {
		return descripcion
	}
	if catalogo, ok := actividad["EventoCatalogoId"].(map[string]interface{}); ok {
		if descripcion, ok := catalogo["Descripcion"].(string); ok && descripcion != "" {
			return descripcion
		}
	}
	return ""
}

func dependenciaActividadMasiva(actividad map[string]interface{}, dependenciaActividad map[string]interface{}, idActividad string, programas []int, operacion string) (map[string]interface{}, bool, []map[string]interface{}, error) {
	resultado := deepCopyMap(dependenciaActividad)
	programasActuales := proyectosDependencia(dependenciaActividad)
	impactos := make([]map[string]interface{}, 0)
	cambio := false

	if operacion == "asociar" {
		for _, programaID := range programas {
			if !programasActuales[programaID] {
				programasActuales[programaID] = true
				cambio = true
			}
		}
		resultado["proyectos"] = proyectosOrdenadosInterface(programasActuales)
		fechasActualizadas, cambioFechas := fechasConDefaultProgramas(actividad, dependenciaActividad, programas)
		if cambioFechas {
			cambio = true
		}
		resultado["fechas"] = fechasActualizadas
		return resultado, cambio, impactos, nil
	}

	for _, programaID := range programas {
		if !programasActuales[programaID] {
			continue
		}
		fechaParticular := tieneFechaParticularModificada(actividad, dependenciaActividad, programaID)
		extensionVigente, err := tieneExtensionVigenteDependencia(idActividad, programaID)
		if err != nil {
			return nil, false, nil, err
		}
		if fechaParticular || extensionVigente {
			impactos = append(impactos, map[string]interface{}{
				"ProgramaId":       programaID,
				"ActividadId":      idActividad,
				"Actividad":        nombreActividadImpacto(actividad),
				"FechaParticular":  fechaParticular,
				"ExtensionVigente": extensionVigente,
			})
			continue
		}
		delete(programasActuales, programaID)
		cambio = true
	}
	if len(impactos) > 0 {
		return resultado, false, impactos, nil
	}

	fechasActualizadas := make([]interface{}, 0)
	if fechasRaw, ok := dependenciaActividad["fechas"].([]interface{}); ok {
		for _, fecha := range fechasRaw {
			fechaMap, ok := fecha.(map[string]interface{})
			if !ok {
				continue
			}
			programaID, ok := interfaceToInt(fechaMap["Id"])
			if ok && programasActuales[programaID] {
				fechasActualizadas = append(fechasActualizadas, fechaMap)
			}
		}
	}
	resultado["proyectos"] = proyectosOrdenadosInterface(programasActuales)
	resultado["fechas"] = fechasActualizadas
	return resultado, cambio, impactos, nil
}

func fechasConDefaultProgramas(actividad map[string]interface{}, dependenciaActividad map[string]interface{}, programas []int) ([]interface{}, bool) {
	fechasActualizadas := make([]interface{}, 0)
	fechasPorPrograma := make(map[int]bool)
	if fechasRaw, ok := dependenciaActividad["fechas"].([]interface{}); ok {
		for _, fecha := range fechasRaw {
			fechaMap, ok := fecha.(map[string]interface{})
			if !ok {
				continue
			}
			programaID, ok := interfaceToInt(fechaMap["Id"])
			if ok {
				fechasPorPrograma[programaID] = true
			}
			fechasActualizadas = append(fechasActualizadas, fechaMap)
		}
	}

	cambio := false
	for _, programaID := range programas {
		if fechasPorPrograma[programaID] {
			continue
		}
		fechasActualizadas = append(fechasActualizadas, map[string]interface{}{
			"Id":           programaID,
			"Inicio":       fechaDependenciaActividad(actividad["FechaInicio"]),
			"Fin":          fechaDependenciaActividad(actividad["FechaFin"]),
			"Modificacion": time.Now().In(helpers.GMTMinus5Location).Format("2006-01-02T15:04:05"),
			"Activo":       true,
		})
		fechasPorPrograma[programaID] = true
		cambio = true
	}
	return fechasActualizadas, cambio
}

func fechaDependenciaActividad(value interface{}) string {
	fecha := strings.TrimSpace(fmt.Sprintf("%v", value))
	if fecha == "" || fecha == "<nil>" {
		return ""
	}
	parsed, err := helpers.ParseFecha(fecha)
	if err == nil {
		return parsed.In(helpers.GMTMinus5Location).Format("2006-01-02T15:04:05")
	}
	fecha = strings.Replace(fecha, " ", "T", 1)
	if len(fecha) >= len("2006-01-02T15:04:05") {
		return fecha[:len("2006-01-02T15:04:05")]
	}
	return fecha
}

func actividadPerteneceACalendario(actividad map[string]interface{}, idCalendario string) (bool, error) {
	proceso, ok := actividad["ProcesoId"].(map[string]interface{})
	if !ok || proceso == nil {
		return false, nil
	}
	if calendarioPertenece(proceso["CalendarioID"], idCalendario) {
		return true, nil
	}
	idProceso, ok := idToString(proceso["Id"])
	if !ok {
		return false, nil
	}
	var procesoCompleto map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"proceso/"+idProceso, &procesoCompleto); err != nil || procesoCompleto == nil || procesoCompleto["Type"] == "error" {
		return false, errors.New("error del servicio PostActividadesProgramasMasivo: no fue posible validar el proceso de la actividad")
	}
	return calendarioPertenece(procesoCompleto["CalendarioID"], idCalendario), nil
}

func calendarioPertenece(calendarioValue interface{}, idCalendario string) bool {
	calendario, ok := calendarioValue.(map[string]interface{})
	if !ok || calendario == nil {
		return false
	}
	id, ok := idToString(calendario["Id"])
	return ok && id == idCalendario
}

func intSliceUnicos(valores []int) []int {
	vistos := make(map[int]bool)
	resultado := make([]int, 0)
	for _, valor := range valores {
		if valor <= 0 || vistos[valor] {
			continue
		}
		vistos[valor] = true
		resultado = append(resultado, valor)
	}
	return resultado
}

func proyectosOrdenadosInterface(proyectos map[int]bool) []interface{} {
	resultado := make([]interface{}, 0, len(proyectos))
	for proyectoID := range proyectos {
		resultado = append(resultado, proyectoID)
	}
	return resultado
}

func validarDesasociacionActividad(idActividad string, actividad map[string]interface{}, dependenciaNueva string) error {
	dependenciaAnterior, okAnterior := parseDependenciaEvento(actividad["DependenciaId"])
	dependenciaNuevaMap, okNueva := parseDependenciaEvento(dependenciaNueva)
	if !okAnterior {
		return nil
	}
	if !okNueva {
		dependenciaNuevaMap = map[string]interface{}{"proyectos": []interface{}{}, "fechas": []interface{}{}}
	}

	programasAnteriores := proyectosDependencia(dependenciaAnterior)
	programasNuevos := proyectosDependencia(dependenciaNuevaMap)
	removidos := make([]int, 0)
	for programaID := range programasAnteriores {
		if !programasNuevos[programaID] {
			removidos = append(removidos, programaID)
		}
	}
	if len(removidos) == 0 {
		return nil
	}

	programasBloqueados := make([]string, 0)
	for _, programaID := range removidos {
		if tieneFechaParticularModificada(actividad, dependenciaAnterior, programaID) {
			programasBloqueados = append(programasBloqueados, strconv.Itoa(programaID))
			continue
		}
		tieneExtension, err := tieneExtensionVigenteDependencia(idActividad, programaID)
		if err != nil {
			return err
		}
		if tieneExtension {
			programasBloqueados = append(programasBloqueados, strconv.Itoa(programaID))
		}
	}

	if len(programasBloqueados) > 0 {
		return errors.New("No se pueden desasociar los programas académicos " + strings.Join(programasBloqueados, ", ") + " porque tienen fechas particulares modificadas o extensiones vigentes. Anule o normalice esas modificaciones antes de desasociarlos.")
	}
	return nil
}

func proyectosDependencia(dependenciaMap map[string]interface{}) map[int]bool {
	resultado := make(map[int]bool)
	proyectos, ok := dependenciaMap["proyectos"].([]interface{})
	if !ok {
		return resultado
	}
	for _, proyecto := range proyectos {
		if proyectoID, ok := interfaceToInt(proyecto); ok && proyectoID > 0 {
			resultado[proyectoID] = true
		}
	}
	return resultado
}

func tieneFechaParticularModificada(actividad map[string]interface{}, dependenciaMap map[string]interface{}, programaID int) bool {
	fechaParticular, ok := fechaParticularProyecto(dependenciaMap, programaID)
	if !ok {
		return false
	}
	fechaInicioGlobal, errInicioGlobal := parseFechaExtension(fmt.Sprintf("%v", actividad["FechaInicio"]))
	fechaFinGlobal, errFinGlobal := parseFechaExtension(fmt.Sprintf("%v", actividad["FechaFin"]))
	fechaInicioParticular, errInicio := parseFechaExtension(fmt.Sprintf("%v", fechaParticular["Inicio"]))
	fechaFinParticular, errFin := parseFechaExtension(fmt.Sprintf("%v", fechaParticular["Fin"]))
	if errInicioGlobal != nil || errFinGlobal != nil || errInicio != nil || errFin != nil {
		return fmt.Sprintf("%v", fechaParticular["Inicio"]) != fmt.Sprintf("%v", actividad["FechaInicio"]) || fmt.Sprintf("%v", fechaParticular["Fin"]) != fmt.Sprintf("%v", actividad["FechaFin"])
	}
	return !fechaInicioParticular.Equal(fechaInicioGlobal) || !fechaFinParticular.Equal(fechaFinGlobal)
}

func deepCopyMap(src map[string]interface{}) map[string]interface{} {
	return helpers.DeepCopyMap(src)
}
