package helpers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/udistrital/sga_calendario_mid/models"
)

func ParseDependenciaEvento(value interface{}) (models.DependenciaEvento, bool) {
	dependenciaStr, ok := value.(string)
	if !ok {
		return models.DependenciaEvento{}, false
	}
	dependenciaStr = strings.TrimSpace(dependenciaStr)
	if dependenciaStr == "" || dependenciaStr == "{}" || dependenciaStr == "<nil>" {
		return models.DependenciaEvento{}, false
	}

	var dependencia models.DependenciaEvento
	if err := json.Unmarshal([]byte(dependenciaStr), &dependencia); err != nil {
		logs.Error("error parseando DependenciaId de evento: ", err)
		return models.DependenciaEvento{}, false
	}
	return dependencia, true
}

func DependenciaIncluyeProyecto(dependencia models.DependenciaEvento, proyectoID int) bool {
	for _, proyecto := range dependencia.Proyectos {
		if proyecto == proyectoID {
			return true
		}
	}
	return false
}

func FechaParticularProyecto(dependencia models.DependenciaEvento, proyectoID int) (models.FechaParticularPrograma, bool) {
	for _, fecha := range dependencia.Fechas {
		if fecha.Id != proyectoID {
			continue
		}
		if !fecha.ActivoPtr() {
			return models.FechaParticularPrograma{}, false
		}
		return fecha, true
	}
	return models.FechaParticularPrograma{}, false
}

func ValidarFechasDependenciaEvento(value interface{}) error {
	dependencia, ok := ParseDependenciaEvento(value)
	if !ok {
		return nil
	}
	for _, fecha := range dependencia.Fechas {
		if !fecha.ActivoPtr() {
			continue
		}
		inicio := strings.TrimSpace(fecha.Inicio)
		fin := strings.TrimSpace(fecha.Fin)
		if inicio == "" && fin == "" {
			continue
		}
		if inicio == "" || fin == "" {
			return fmt.Errorf("la fecha particular del programa %d debe tener fecha inicio y fecha fin", fecha.Id)
		}
		inicioParsed, err := ParseFecha(inicio)
		if err != nil {
			return fmt.Errorf("la fecha inicio particular del programa %d no tiene un formato válido", fecha.Id)
		}
		finParsed, err := ParseFecha(fin)
		if err != nil {
			return fmt.Errorf("la fecha fin particular del programa %d no tiene un formato válido", fecha.Id)
		}
		if finParsed.Before(inicioParsed) {
			return fmt.Errorf("la fecha fin particular del programa %d no puede ser menor que la fecha inicio particular", fecha.Id)
		}
	}
	return nil
}
