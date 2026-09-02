package helpers

import (
	"encoding/json"

	"github.com/astaxie/beego/logs"
)

func ParseDependenciaEventoMap(dependencia interface{}) (map[string]interface{}, bool) {
	dependenciaModel, ok := ParseDependenciaEvento(dependencia)
	if !ok {
		return nil, false
	}
	var dependenciaMap map[string]interface{}
	data, _ := json.Marshal(dependenciaModel)
	if err := json.Unmarshal(data, &dependenciaMap); err != nil {
		logs.Error("error parseando DependenciaId de evento: ", err)
		return nil, false
	}
	return dependenciaMap, true
}

func DependenciaMapIncluyeProyecto(dependenciaMap map[string]interface{}, proyectoID int) bool {
	return ProyectosDependenciaMap(dependenciaMap)[proyectoID]
}

func FechaParticularProyectoMap(dependenciaMap map[string]interface{}, proyectoID int) (map[string]interface{}, bool) {
	fechas, ok := dependenciaMap["fechas"].([]interface{})
	if !ok {
		return nil, false
	}
	for _, fecha := range fechas {
		fechaMap, ok := fecha.(map[string]interface{})
		if !ok {
			continue
		}
		id, ok := InterfaceToInt(fechaMap["Id"])
		if !ok || id != proyectoID {
			continue
		}
		activo, ok := fechaMap["Activo"].(bool)
		if ok && !activo {
			return nil, false
		}
		return fechaMap, true
	}
	return nil, false
}

func ProyectosDependenciaMap(dependenciaMap map[string]interface{}) map[int]bool {
	resultado := make(map[int]bool)
	proyectos, ok := dependenciaMap["proyectos"].([]interface{})
	if !ok {
		return resultado
	}
	for _, proyecto := range proyectos {
		if proyectoID, ok := InterfaceToInt(proyecto); ok && proyectoID > 0 {
			resultado[proyectoID] = true
		}
	}
	return resultado
}

func CalendarioPertenece(calendarioValue interface{}, idCalendario string) bool {
	calendario, ok := calendarioValue.(map[string]interface{})
	if !ok || calendario == nil {
		return false
	}
	id, ok := IDToString(calendario["Id"])
	return ok && id == idCalendario
}
