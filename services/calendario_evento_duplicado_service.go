package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/udistrital/sga_calendario_mid/helpers"
	"github.com/udistrital/utils_oas/request"
)

func validarDuplicadoCalendarioEvento(evento map[string]interface{}, idExcluir string) error {
	activo, ok := evento["Activo"].(bool)
	if !ok || !activo {
		return nil
	}
	repetible, err := esRepetibleCalendarioEvento(evento)
	if err != nil {
		return err
	}
	return validarDuplicadoCalendarioEventoConPolitica(evento, idExcluir, repetible)
}

func validarDuplicadoCalendarioEventoConPolitica(evento map[string]interface{}, idExcluir string, repetible bool) error {
	procesoID, err := helpers.RelationIDToString(evento["ProcesoId"])
	if err != nil {
		return errors.New("no fue posible validar duplicados: proceso inválido")
	}
	eventoCatalogoID, err := helpers.RelationIDToString(evento["EventoCatalogoId"])
	if err != nil {
		return errors.New("no fue posible validar duplicados: evento de catálogo inválido")
	}
	var existentes []map[string]interface{}
	url := fmt.Sprintf(
		"%scalendario_evento?query=Activo:true,ProcesoId__Id:%s,EventoCatalogoId__Id:%s&limit=0",
		beego.AppConfig.String("EventoService"),
		procesoID,
		eventoCatalogoID,
	)
	if err := request.GetJson(url, &existentes); err != nil {
		return errors.New("no fue posible validar duplicados de actividad")
	}
	return validarActividadActivaExistente(evento, existentes, idExcluir, repetible)
}

func esRepetibleCalendarioEvento(evento map[string]interface{}) (bool, error) {
	procesoID, err := helpers.RelationIDToString(evento["ProcesoId"])
	if err != nil {
		return false, errors.New("no fue posible validar duplicados: proceso inválido")
	}
	eventoCatalogoID, err := helpers.RelationIDToString(evento["EventoCatalogoId"])
	if err != nil {
		return false, errors.New("no fue posible validar duplicados: evento de catálogo inválido")
	}

	var proceso map[string]interface{}
	if err := request.GetJson(beego.AppConfig.String("EventoService")+"proceso/"+procesoID, &proceso); err != nil || proceso == nil || proceso["Type"] == "error" {
		return false, errors.New("no fue posible validar la política de la actividad")
	}
	procesoCatalogoID, err := helpers.RelationIDToString(proceso["ProcesoCatalogoId"])
	if err != nil {
		return false, errors.New("no fue posible validar la política de la actividad")
	}

	var relaciones []map[string]interface{}
	url := fmt.Sprintf(
		"%sevento_catalogo_proceso_catalogo?query=Activo:true,EventoCatalogoId__Id:%s,ProcesoCatalogoId__Id:%s&limit=1",
		beego.AppConfig.String("EventoService"),
		eventoCatalogoID,
		procesoCatalogoID,
	)
	if err := request.GetJson(url, &relaciones); err != nil || len(relaciones) == 0 || len(relaciones[0]) == 0 {
		return false, errors.New("la actividad seleccionada no pertenece al proceso")
	}

	repetible, _ := relaciones[0]["Repetible"].(bool)
	return repetible, nil
}

func validarActividadActivaExistente(evento map[string]interface{}, existentes []map[string]interface{}, idExcluir string, repetible bool) error {
	activo, ok := evento["Activo"].(bool)
	if !ok || !activo {
		return nil
	}
	idExcluirNumerico, errIDExcluir := strconv.Atoi(idExcluir)

	for _, existente := range existentes {
		if len(existente) == 0 {
			continue
		}
		if id, ok := helpers.InterfaceToInt(existente["Id"]); ok && errIDExcluir == nil && id == idExcluirNumerico {
			continue
		}
		if !repetible {
			return errors.New("la actividad ya tiene una ocurrencia activa en el proceso; debe inactivarla antes de crear otra")
		}
		rangoIgual, err := rangosCalendarioEventoIguales(evento, existente)
		if err != nil {
			return errors.New("no fue posible comparar el rango de fechas de la actividad")
		}
		if rangoIgual {
			return errors.New("la actividad ya tiene una ocurrencia activa con el mismo rango de fechas")
		}
	}

	return nil
}

func rangosCalendarioEventoIguales(primero, segundo map[string]interface{}) (bool, error) {
	inicioPrimero, existe, err := fechaCalendarioEvento(primero["FechaInicio"])
	if err != nil || !existe {
		return false, errors.New("fecha de inicio inválida")
	}
	inicioSegundo, existe, err := fechaCalendarioEvento(segundo["FechaInicio"])
	if err != nil || !existe {
		return false, errors.New("fecha de inicio inválida")
	}
	if !inicioPrimero.Equal(inicioSegundo) {
		return false, nil
	}

	finPrimero, tieneFinPrimero, err := fechaCalendarioEvento(primero["FechaFin"])
	if err != nil {
		return false, err
	}
	finSegundo, tieneFinSegundo, err := fechaCalendarioEvento(segundo["FechaFin"])
	if err != nil {
		return false, err
	}
	if tieneFinPrimero != tieneFinSegundo {
		return false, nil
	}
	return !tieneFinPrimero || finPrimero.Equal(finSegundo), nil
}

func fechaCalendarioEvento(valor interface{}) (time.Time, bool, error) {
	texto := strings.TrimSpace(fmt.Sprintf("%v", valor))
	if valor == nil || texto == "" || texto == "<nil>" {
		return time.Time{}, false, nil
	}
	fecha, err := helpers.ParseFecha(texto)
	return fecha, true, err
}

func calendarioEventoEfectivo(actual map[string]interface{}, cambios map[string]interface{}) map[string]interface{} {
	efectivo := make(map[string]interface{}, len(actual)+len(cambios))
	for campo, valor := range actual {
		efectivo[campo] = valor
	}
	for campo, valor := range cambios {
		if campo == "NumeroOcurrencia" {
			continue
		}
		efectivo[campo] = valor
	}
	return efectivo
}
