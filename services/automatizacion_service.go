package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/astaxie/beego"
	"github.com/go-co-op/gocron"
	"github.com/udistrital/sga_calendario_mid/helpers"
)

func VerificacionActividadesDiarias() {
	// fmt.Println("Checking daily activities.....................................")

	var calendarios []map[string]interface{}

	resp, err := http.Get("http://" + beego.AppConfig.String("EventoService") + "automatizacion?limit=0&query=ejecucion_unica:false")
	if err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			err = json.Unmarshal(body, &calendarios)
			if err == nil {
				for _, calendario := range calendarios {

					fechaFinStr, ok := calendario["CalendarioEventoId"].(map[string]interface{})["FechaFin"].(string)
					if !ok {
						fmt.Println("FechaFin no es una cadena de texto válida")
						continue
					}

					fechaFin, err := helpers.ParseFecha(fechaFinStr)
					if err != nil {
						fmt.Println("Error al parsear la FechaFin:", err)
						continue
					}

					now := time.Now().In(helpers.GMTMinus5Location)

					// Comparar solo la fecha (ignorando la hora)
					if isSameDate(fechaFin, now) {

						// Se deserializa el endpoint JSON para obtener el valor de data
						endpointJSON := calendario["Endpoint"].(string)
						var endpointData map[string]string

						err := json.Unmarshal([]byte(endpointJSON), &endpointData)
						if err != nil {
							fmt.Println("Error al deserializar el endpoint JSON:", err)
							continue
						}
						dataValue := endpointData["data"]

						EjecucionEndpoint(dataValue)
					} else {
						fmt.Println("La fecha de fin no es hoy. No se ejecuta ninguna acción.")
					}

				}
			}
		}
	} else {
		// Handle the error
		fmt.Println("Error calling endpoint:", err)
	}

}

func VerificacionActividadesEjecucionUnica(scheduler *gocron.Scheduler) {
	// fmt.Println("Scheduling one-time activities.....................................")
	var calendarios []map[string]interface{}

	resp, err := http.Get("http://" + beego.AppConfig.String("EventoService") + "automatizacion?limit=0&query=ejecucion_unica:false")
	if err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			err = json.Unmarshal(body, &calendarios)
			if err == nil {
				for _, calendario := range calendarios {

					fechaFinStr, _ := calendario["CalendarioEventoId"].(map[string]interface{})["FechaFin"].(string)
					// fechaFinStr := "2024-07-18T21:49:00Z"
					fmt.Println("FechaFin:", fechaFinStr)

					fechaFin, err := helpers.ParseFecha(fechaFinStr)
					if err != nil {
						fmt.Println("Error parsing fechaFinStr:", err)
						// continue
					}

					// Función anónima que se ejecutará cuando se dispare la tarea
					task := func() {
						fmt.Println("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
						// Aquí puedes poner la lógica adicional que desees ejecutar
					}

					// Programar la tarea utilizando ScheduleOneTimeTask
					ScheduleOneTimeTask(scheduler, fechaFin, task)

				}
			}
		}
	} else {
		// Handle the error
		fmt.Println("Error calling endpoint:", err)
	}

}

func ScheduleOneTimeTask(scheduler *gocron.Scheduler, t time.Time, task func()) {
	if t.In(helpers.GMTMinus5Location).After(time.Now().In(helpers.GMTMinus5Location)) {
		scheduler.StartAt(t).Do(task)
		fmt.Printf("Task scheduled to run at %v\n", t)
	} else {
		fmt.Println("The specified time is in the past.")
	}
}

func isSameDate(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.Month() == t2.Month() && t1.Day() == t2.Day()
}

func EjecucionEndpoint(url string) {
	//llamando al endpoint
	resp, err := http.Get("http://" + url)
	if err != nil {
		fmt.Println("Error al llamar al endpoint:", err)
	}

	_ = resp // response
}
