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

func PutCalendarioEstado(id string, data []byte) (interface{}, error) {
	var recibido models.EstadoActivoRequest
	if err := json.Unmarshal(data, &recibido); err != nil {
		return nil, errors.New("error del servicio PutCalendarioEstado: solicitud inválida")
	}
	if recibido.TerceroId <= 0 {
		return nil, errors.New("error del servicio PutCalendarioEstado: TerceroId inválido")
	}

	var calendario map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario/"+id, &calendario); err != nil || calendario == nil || calendario["Type"] == "error" {
		return nil, errors.New("error del servicio PutCalendarioEstado: no fue posible consultar el calendario")
	}

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

	if !activo {
		if err := inactivarProcesosCalendario(id, recibido.TerceroId); err != nil {
			logs.Error(err)
		}
		if err := inactivarEventosCalendario(id, recibido.TerceroId); err != nil {
			return nil, err
		}
	}

	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func PutCalendarioDependencias(id string, data []byte) (interface{}, error) {
	var recibido models.CalendarioDependenciasRequest
	if err := json.Unmarshal(data, &recibido); err != nil {
		return nil, errors.New("error del servicio PutCalendarioDependencias: solicitud inválida")
	}
	if recibido.TerceroId <= 0 {
		return nil, errors.New("error del servicio PutCalendarioDependencias: TerceroId inválido")
	}
	var calendario map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario/"+id, &calendario); err != nil || calendario == nil || calendario["Type"] == "error" {
		return nil, errors.New("error del servicio PutCalendarioDependencias: no fue posible consultar el calendario")
	}

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

	if err := cascadaDependenciasCalendario(eventos, dependenciaId, recibido.TerceroId); err != nil {
		return nil, err
	}

	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func impactosDesasociacionCalendario(idCalendario string, calendario map[string]interface{}, dependenciaNueva string) ([]map[string]interface{}, []map[string]interface{}, error) {
	dependenciaAnterior, okAnterior := helpers.ParseDependenciaEventoMap(calendario["DependenciaId"])
	dependenciaNuevaMap, okNueva := helpers.ParseDependenciaEventoMap(dependenciaNueva)
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
		dependenciaActividad, ok := helpers.ParseDependenciaEventoMap(actividad["DependenciaId"])
		if !ok {
			continue
		}
		idActividad, idOk := helpers.IDToString(actividad["Id"])
		if !idOk {
			continue
		}
		for _, programaID := range removidos {
			if !helpers.DependenciaMapIncluyeProyecto(dependenciaActividad, programaID) {
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
	programasAnteriores := helpers.ProyectosDependenciaMap(dependenciaAnterior)
	programasNuevos := helpers.ProyectosDependenciaMap(dependenciaNueva)
	removidos := make([]int, 0)
	for programaID := range programasAnteriores {
		if !programasNuevos[programaID] {
			removidos = append(removidos, programaID)
		}
	}
	return removidos
}

func cascadaDependenciasCalendario(eventos []map[string]interface{}, dependenciaNueva string, terceroID int) error {
	dependenciaNuevaMap, okNueva := helpers.ParseDependenciaEventoMap(dependenciaNueva)
	if !okNueva {
		dependenciaNuevaMap = map[string]interface{}{"proyectos": []interface{}{}, "fechas": []interface{}{}}
	}
	programasCalendario := helpers.ProyectosDependenciaMap(dependenciaNuevaMap)
	for _, actividad := range eventos {
		dependenciaActividad, ok := helpers.ParseDependenciaEventoMap(actividad["DependenciaId"])
		if !ok {
			continue
		}
		dependenciaActualizada, removidos, cambio := filtrarDependenciaActividadPorCalendario(dependenciaActividad, programasCalendario)
		if !cambio {
			continue
		}
		idActividad, idOk := helpers.IDToString(actividad["Id"])
		if !idOk {
			continue
		}
		for _, programaID := range removidos {
			if err := inactivarExtensionesDependenciaRetirada(idActividad, programaID); err != nil {
				return err
			}
		}
		dependenciaBytes, _ := json.Marshal(dependenciaActualizada)
		actividad["DependenciaId"] = string(dependenciaBytes)
		helpers.SetTerceroID(actividad, terceroID)
		var resultado map[string]interface{}
		if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento/"+idActividad, "PUT", &resultado, actividad); err != nil || resultado == nil || resultado["Type"] == "error" {
			return errors.New("error del servicio PutCalendarioDependencias: no fue posible actualizar actividades en cascada")
		}
	}
	return nil
}

func filtrarDependenciaActividadPorCalendario(dependenciaActividad map[string]interface{}, programasCalendario map[int]bool) (map[string]interface{}, []int, bool) {
	resultado := helpers.DeepCopyMap(dependenciaActividad)
	proyectosRaw, _ := dependenciaActividad["proyectos"].([]interface{})
	proyectosActualizados := make([]interface{}, 0)
	removidos := make([]int, 0)
	for _, proyecto := range proyectosRaw {
		programaID, ok := helpers.InterfaceToInt(proyecto)
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
			programaID, ok := helpers.InterfaceToInt(fechaMap["Id"])
			if ok && programasCalendario[programaID] {
				fechasActualizadas = append(fechasActualizadas, fechaMap)
			}
		}
	}
	resultado["proyectos"] = proyectosActualizados
	resultado["fechas"] = fechasActualizadas
	return resultado, removidos, true
}

func inactivarExtensionesDependenciaRetirada(idActividad string, dependenciaID int) error {
	var relaciones []map[string]interface{}
	url := beego.AppConfig.String("EventoService") + "calendario_evento_extension_programa?query=Activo:true,Vigente:true,CalendarioEventoExtensionId__CalendarioEventoId__Id:" + idActividad + ",DependenciaId:" + strconv.Itoa(dependenciaID) + "&limit=0"
	if err := request.GetJson(url, &relaciones); err != nil {
		return errors.New("error del servicio PutCalendarioDependencias: no fue posible consultar extensiones vigentes")
	}
	for _, relacion := range relaciones {
		idRelacion, ok := helpers.IDToString(relacion["Id"])
		if !ok {
			continue
		}
		relacion["Activo"] = false
		relacion["Vigente"] = false
		var resultadoRelacion map[string]interface{}
		if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento_extension_programa/"+idRelacion, "PUT", &resultadoRelacion, relacion); err != nil || resultadoRelacion == nil || resultadoRelacion["Type"] == "error" {
			return errors.New("error del servicio PutCalendarioDependencias: no fue posible inactivar extensión por programa")
		}
		if extension, ok := relacion["CalendarioEventoExtensionId"].(map[string]interface{}); ok {
			if extensionID, ok := helpers.InterfaceToInt(extension["Id"]); ok {
				if err := inactivarExtensionSinProgramasVigentes(idActividad, extensionID); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func inactivarExtensionSinProgramasVigentes(idActividad string, extensionID int) error {
	var relacionesVigentes []map[string]interface{}
	url := beego.AppConfig.String("EventoService") + "calendario_evento_extension_programa?query=Activo:true,Vigente:true,CalendarioEventoExtensionId__CalendarioEventoId__Id:" + idActividad + ",CalendarioEventoExtensionId__Id:" + strconv.Itoa(extensionID) + "&limit=1"
	if err := request.GetJson(url, &relacionesVigentes); err != nil || (len(relacionesVigentes) > 0 && len(relacionesVigentes[0]) > 0) {
		return nil
	}
	var extension map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento_extension/"+strconv.Itoa(extensionID), &extension); err != nil || extension == nil || extension["Type"] == "error" {
		return errors.New("error del servicio PutCalendarioDependencias: no fue posible consultar extensión para cascada")
	}
	extension["Activo"] = false
	var resultado map[string]interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento_extension/"+strconv.Itoa(extensionID), "PUT", &resultado, extension); err != nil || resultado == nil || resultado["Type"] == "error" {
		return errors.New("error del servicio PutCalendarioDependencias: no fue posible inactivar extensión sin programas vigentes")
	}
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
	if id, ok := helpers.IDToString(actividad["Id"]); ok {
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

func PostProcesoCalendario(data []byte) (interface{}, error) {
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

	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func PutProcesoPeriodicidad(id string, data []byte) (interface{}, error) {
	var recibido models.ProcesoPeriodicidadRequest
	if err := json.Unmarshal(data, &recibido); err != nil {
		return nil, errors.New("error del servicio PutProcesoPeriodicidad: solicitud inválida")
	}

	var proceso map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"proceso/"+id, &proceso); err != nil || proceso == nil || proceso["Type"] == "error" {
		return nil, errors.New("error del servicio PutProcesoPeriodicidad: no fue posible consultar el proceso")
	}

	if recibido.TipoRecurrenciaId.Id <= 0 {
		return nil, errors.New("error del servicio PutProcesoPeriodicidad: periodicidad inválida")
	}
	proceso["TipoRecurrenciaId"] = recibido.TipoRecurrenciaId

	var resultado map[string]interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+"proceso/"+id, "PUT", &resultado, proceso); err != nil || resultado == nil || resultado["Type"] == "error" {
		return nil, errors.New("error del servicio PutProcesoPeriodicidad: no fue posible actualizar el proceso")
	}

	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func PutProcesoEstado(id string, data []byte) (interface{}, error) {
	var recibido models.EstadoActivoRequest
	if err := json.Unmarshal(data, &recibido); err != nil {
		return nil, errors.New("error del servicio PutProcesoEstado: solicitud inválida")
	}
	if recibido.TerceroId <= 0 {
		return nil, errors.New("error del servicio PutProcesoEstado: TerceroId inválido")
	}

	var proceso map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"proceso/"+id, &proceso); err != nil || proceso == nil || proceso["Type"] == "error" {
		return nil, errors.New("error del servicio PutProcesoEstado: no fue posible consultar el proceso")
	}

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

	if !activo {
		if err := inactivarEventosProceso(id, recibido.TerceroId); err != nil {
			return nil, err
		}
	}

	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func inactivarProcesosCalendario(idCalendario string, terceroID int) error {
	var procesos []map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"proceso?query=Activo:true,CalendarioID__Id:"+idCalendario+"&limit=0", &procesos); err != nil {
		return errors.New("error inactivando procesos del calendario")
	}
	for _, proceso := range procesos {
		idProceso, ok := helpers.IDToString(proceso["Id"])
		if !ok {
			continue
		}
		proceso["Activo"] = false

		var resultado map[string]interface{}
		if err := request.SendJson(beego.AppConfig.String("EventoService")+"proceso/"+idProceso, "PUT", &resultado, proceso); err != nil || resultado == nil || resultado["Type"] == "error" {
			return errors.New("error inactivando proceso del calendario")
		}
		if err := inactivarEventosProceso(idProceso, terceroID); err != nil {
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
	idCalendario, ok := helpers.IDToString(calendario["Id"])
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
	idProceso, ok := helpers.IDToString(proceso["Id"])
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

func inactivarEventosProceso(idProceso string, terceroID int) error {
	var eventos []map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento?query=Activo:true,ProcesoId__Id:"+idProceso+"&limit=0", &eventos); err != nil {
		return errors.New("error inactivando eventos del proceso")
	}
	for _, evento := range eventos {
		if err := inactivarEvento(evento, terceroID); err != nil {
			return err
		}
	}
	return nil
}

func inactivarEventosCalendario(idCalendario string, terceroID int) error {
	var eventos []map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento?query=Activo:true,ProcesoId__CalendarioID__Id:"+idCalendario+"&limit=0", &eventos); err != nil {
		return errors.New("error inactivando eventos del calendario")
	}
	for _, evento := range eventos {
		if err := inactivarEvento(evento, terceroID); err != nil {
			return err
		}
	}
	return nil
}

func inactivarEvento(evento map[string]interface{}, terceroID int) error {
	idEvento, ok := helpers.IDToString(evento["Id"])
	if !ok {
		return nil
	}
	evento["Activo"] = false
	helpers.SetTerceroID(evento, terceroID)

	var resultado map[string]interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento/"+idEvento, "PUT", &resultado, evento); err != nil || resultado == nil || resultado["Type"] == "error" {
		return errors.New("error inactivando evento del proceso")
	}
	if err := inactivarExtensionesActividad(idEvento); err != nil {
		logs.Error(err)
	}
	return nil
}

func PutActividadDependencias(id string, data []byte) (interface{}, error) {
	var recibido models.ActividadDependenciasRequest
	if err := json.Unmarshal(data, &recibido); err != nil {
		return nil, errors.New("error del servicio PutActividadDependencias: solicitud inválida")
	}
	if recibido.TerceroId <= 0 {
		return nil, errors.New("error del servicio PutActividadDependencias: TerceroId inválido")
	}

	var actividad map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario_evento/"+id, &actividad); err != nil || actividad == nil || actividad["Type"] == "error" {
		return nil, errors.New("error del servicio PutActividadDependencias: no fue posible consultar la actividad")
	}

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
	helpers.SetTerceroID(actividad, recibido.TerceroId)

	var resultado map[string]interface{}
	if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento/"+id, "PUT", &resultado, actividad); err != nil || resultado == nil || resultado["Type"] == "error" {
		return nil, errors.New("error del servicio PutActividadDependencias: no fue posible actualizar la actividad")
	}

	return requestresponse.APIResponseDTO(true, 200, resultado), nil
}

func PostActividadesProgramasMasivo(idCalendario string, data []byte) (interface{}, error) {
	var recibido models.ActividadesProgramasMasivoRequest
	if err := json.Unmarshal(data, &recibido); err != nil {
		return nil, errors.New("error del servicio PostActividadesProgramasMasivo: solicitud inválida")
	}
	if recibido.TerceroId <= 0 {
		return nil, errors.New("error del servicio PostActividadesProgramasMasivo: TerceroId inválido")
	}

	operacion := strings.ToLower(strings.TrimSpace(recibido.Operacion))
	if operacion != "asociar" && operacion != "desasociar" {
		return nil, errors.New("error del servicio PostActividadesProgramasMasivo: operación inválida")
	}
	programas := helpers.IntSliceUnicos(recibido.ProgramaIds)
	actividades := helpers.IntSliceUnicos(recibido.ActividadIds)
	if len(programas) == 0 || len(actividades) == 0 {
		return nil, errors.New("error del servicio PostActividadesProgramasMasivo: debe seleccionar programas y actividades")
	}

	var calendario map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario/"+idCalendario, &calendario); err != nil || calendario == nil || calendario["Type"] == "error" {
		return nil, errors.New("error del servicio PostActividadesProgramasMasivo: no fue posible consultar el calendario")
	}
	dependenciaCalendario, ok := helpers.ParseDependenciaEventoMap(calendario["DependenciaId"])
	if !ok {
		return nil, errors.New("error del servicio PostActividadesProgramasMasivo: calendario sin programas académicos válidos")
	}
	programasCalendario := helpers.ProyectosDependenciaMap(dependenciaCalendario)
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

		dependenciaActividad, ok := helpers.ParseDependenciaEventoMap(actividad["DependenciaId"])
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
		dependenciaBytes, _ := json.Marshal(dependenciaActualizada)
		actividad["DependenciaId"] = string(dependenciaBytes)
		helpers.SetTerceroID(actividad, recibido.TerceroId)
		var resultado map[string]interface{}
		if err := request.SendJson(beego.AppConfig.String("EventoService")+"calendario_evento/"+idActividad, "PUT", &resultado, actividad); err != nil || resultado == nil || resultado["Type"] == "error" {
			return nil, errors.New("error del servicio PostActividadesProgramasMasivo: no fue posible actualizar la actividad " + idActividad)
		}
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
	programas := helpers.IntSliceUnicos(recibido.ProgramaIds)
	actividades := helpers.IntSliceUnicos(recibido.ActividadIds)
	if len(programas) == 0 {
		return nil, errors.New("error del servicio PostValidarActividadesProgramasMasivo: debe seleccionar programas")
	}

	var calendario map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"calendario/"+idCalendario, &calendario); err != nil || calendario == nil || calendario["Type"] == "error" {
		return nil, errors.New("error del servicio PostValidarActividadesProgramasMasivo: no fue posible consultar el calendario")
	}
	dependenciaCalendario, ok := helpers.ParseDependenciaEventoMap(calendario["DependenciaId"])
	if !ok {
		return nil, errors.New("error del servicio PostValidarActividadesProgramasMasivo: calendario sin programas académicos válidos")
	}
	programasCalendario := helpers.ProyectosDependenciaMap(dependenciaCalendario)
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
		validacion, err := validarActividadProgramasMasivo(actividad, programas)
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

func validarActividadProgramasMasivo(actividad map[string]interface{}, programas []int) (map[string]interface{}, error) {
	idActividad, _ := helpers.IDToString(actividad["Id"])
	dependenciaActividad, ok := helpers.ParseDependenciaEventoMap(actividad["DependenciaId"])
	if !ok {
		dependenciaActividad = map[string]interface{}{"proyectos": []interface{}{}, "fechas": []interface{}{}}
	}
	programasActuales := helpers.ProyectosDependenciaMap(dependenciaActividad)
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
	resultado := helpers.DeepCopyMap(dependenciaActividad)
	programasActuales := helpers.ProyectosDependenciaMap(dependenciaActividad)
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
			programaID, ok := helpers.InterfaceToInt(fechaMap["Id"])
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
			programaID, ok := helpers.InterfaceToInt(fechaMap["Id"])
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
			"Inicio":       helpers.FechaDependenciaActividad(actividad["FechaInicio"]),
			"Fin":          helpers.FechaDependenciaActividad(actividad["FechaFin"]),
			"Modificacion": time.Now().In(helpers.GMTMinus5Location).Format("2006-01-02T15:04:05"),
			"Activo":       true,
		})
		fechasPorPrograma[programaID] = true
		cambio = true
	}
	return fechasActualizadas, cambio
}

func actividadPerteneceACalendario(actividad map[string]interface{}, idCalendario string) (bool, error) {
	proceso, ok := actividad["ProcesoId"].(map[string]interface{})
	if !ok || proceso == nil {
		return false, nil
	}
	if helpers.CalendarioPertenece(proceso["CalendarioID"], idCalendario) {
		return true, nil
	}
	idProceso, ok := helpers.IDToString(proceso["Id"])
	if !ok {
		return false, nil
	}
	var procesoCompleto map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"proceso/"+idProceso, &procesoCompleto); err != nil || procesoCompleto == nil || procesoCompleto["Type"] == "error" {
		return false, errors.New("error del servicio PostActividadesProgramasMasivo: no fue posible validar el proceso de la actividad")
	}
	return helpers.CalendarioPertenece(procesoCompleto["CalendarioID"], idCalendario), nil
}

func proyectosOrdenadosInterface(proyectos map[int]bool) []interface{} {
	resultado := make([]interface{}, 0, len(proyectos))
	for proyectoID := range proyectos {
		resultado = append(resultado, proyectoID)
	}
	return resultado
}

func validarDesasociacionActividad(idActividad string, actividad map[string]interface{}, dependenciaNueva string) error {
	dependenciaAnterior, okAnterior := helpers.ParseDependenciaEventoMap(actividad["DependenciaId"])
	dependenciaNuevaMap, okNueva := helpers.ParseDependenciaEventoMap(dependenciaNueva)
	if !okAnterior {
		return nil
	}
	if !okNueva {
		dependenciaNuevaMap = map[string]interface{}{"proyectos": []interface{}{}, "fechas": []interface{}{}}
	}

	programasAnteriores := helpers.ProyectosDependenciaMap(dependenciaAnterior)
	programasNuevos := helpers.ProyectosDependenciaMap(dependenciaNuevaMap)
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

func tieneFechaParticularModificada(actividad map[string]interface{}, dependenciaMap map[string]interface{}, programaID int) bool {
	fechaParticular, ok := helpers.FechaParticularProyectoMap(dependenciaMap, programaID)
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
