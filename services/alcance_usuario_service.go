package services

import (
	"errors"

	"github.com/astaxie/beego"
	"github.com/udistrital/sga_calendario_mid/helpers"
	"github.com/udistrital/utils_oas/request"
	"github.com/udistrital/utils_oas/requestresponse"
)

func FacultadesOikosSecretario(documento string) (interface{}, error) {
	if documento == "" {
		return nil, errors.New("documento del secretario requerido")
	}
	wso2Service := beego.AppConfig.String("Wso2Service")
	academicaService := beego.AppConfig.String("Wso2AcademicaService")
	homologacionService := beego.AppConfig.String("Wso2HomologacionService")
	if wso2Service == "" || academicaService == "" || homologacionService == "" {
		return nil, errors.New("servicios WSO2 no configurados")
	}

	var resSecretario map[string]interface{}
	if err := request.GetJsonWSO2(wso2Service+academicaService+"/facultad_secretaria/"+documento, &resSecretario); err != nil {
		return nil, errors.New("no fue posible consultar las facultades del secretario")
	}

	codigosCondor := codigosFacultadSecretario(resSecretario)
	if len(codigosCondor) == 0 {
		return nil, errors.New("el secretario no tiene facultades asociadas")
	}

	facultades := make([]int, 0, len(codigosCondor))
	vistos := map[int]bool{}
	for _, codigo := range codigosCondor {
		var resHomologacion map[string]interface{}
		if err := request.GetJsonWSO2(wso2Service+homologacionService+"/facultad_oikos_gedep/"+codigo, &resHomologacion); err != nil {
			continue
		}
		idOikos := idOikosHomologacion(resHomologacion)
		if idOikos > 0 && !vistos[idOikos] {
			vistos[idOikos] = true
			facultades = append(facultades, idOikos)
		}
	}
	if len(facultades) == 0 {
		return nil, errors.New("no se encontraron facultades Oikos para el secretario")
	}
	return requestresponse.APIResponseDTO(true, 200, map[string]interface{}{"Facultades": facultades}), nil
}

func FacultadesOikosDecano(documento string) (interface{}, error) {
	if documento == "" {
		return nil, errors.New("documento del decano requerido")
	}
	wso2Service := beego.AppConfig.String("Wso2Service")
	academicaService := beego.AppConfig.String("Wso2AcademicaService")
	homologacionService := beego.AppConfig.String("Wso2HomologacionService")
	if wso2Service == "" || academicaService == "" || homologacionService == "" {
		return nil, errors.New("servicios WSO2 no configurados")
	}

	var resDecano map[string]interface{}
	if err := request.GetJsonWSO2(wso2Service+academicaService+"/decano/"+documento, &resDecano); err != nil {
		return nil, errors.New("no fue posible consultar las facultades del decano")
	}

	codigosCondor := codigosFacultadDecano(resDecano)
	if len(codigosCondor) == 0 {
		return nil, errors.New("el decano no tiene facultades asociadas")
	}

	facultades := make([]int, 0, len(codigosCondor))
	vistos := map[int]bool{}
	for _, codigo := range codigosCondor {
		var resHomologacion map[string]interface{}
		if err := request.GetJsonWSO2(wso2Service+homologacionService+"/facultad_oikos_gedep/"+codigo, &resHomologacion); err != nil {
			continue
		}
		idOikos := idOikosHomologacion(resHomologacion)
		if idOikos > 0 && !vistos[idOikos] {
			vistos[idOikos] = true
			facultades = append(facultades, idOikos)
		}
	}
	if len(facultades) == 0 {
		return nil, errors.New("no se encontraron facultades Oikos para el decano")
	}
	return requestresponse.APIResponseDTO(true, 200, map[string]interface{}{"Facultades": facultades}), nil
}

func codigosFacultadSecretario(res map[string]interface{}) []string {
	codigos := []string{}
	facultades, ok := res["facultades"].(map[string]interface{})
	if !ok {
		return codigos
	}
	secretaria, ok := facultades["secretaria"].([]interface{})
	if !ok {
		if facultad, ok := facultades["secretaria"].(map[string]interface{}); ok {
			if codigo, ok := facultad["SEC_DEP_COD"].(string); ok && codigo != "" {
				codigos = append(codigos, codigo)
			}
		}
		return codigos
	}
	for _, item := range secretaria {
		facultad, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if codigo, ok := facultad["SEC_DEP_COD"].(string); ok && codigo != "" {
			codigos = append(codigos, codigo)
		}
	}
	return codigos
}

func codigosFacultadDecano(res map[string]interface{}) []string {
	codigos := []string{}
	facultad, ok := res["facultad"].(map[string]interface{})
	if !ok {
		return codigos
	}
	decano, ok := facultad["decano"].([]interface{})
	if !ok {
		if item, ok := facultad["decano"].(map[string]interface{}); ok {
			return appendCodigoFacultadDecano(codigos, item)
		}
		return codigos
	}
	for _, item := range decano {
		facultadDecano, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		codigos = appendCodigoFacultadDecano(codigos, facultadDecano)
	}
	return codigos
}

func appendCodigoFacultadDecano(codigos []string, facultad map[string]interface{}) []string {
	for _, key := range []string{"codigo_facultad", "CODIGO_FACULTAD", "DEC_DEP_COD", "DEP_COD"} {
		if codigo, ok := facultad[key].(string); ok && codigo != "" {
			return append(codigos, codigo)
		}
	}
	return codigos
}

func idOikosHomologacion(res map[string]interface{}) int {
	homologacion, ok := res["homologacion"].(map[string]interface{})
	if !ok {
		return 0
	}
	if id, ok := helpers.InterfaceToInt(homologacion["id_oikos"]); ok {
		return id
	}
	return 0
}
