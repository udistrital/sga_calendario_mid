package services

import (
	"net/url"
	"sync"

	"github.com/astaxie/beego"
	"github.com/udistrital/sga_calendario_mid/helpers"
	"github.com/udistrital/sga_calendario_mid/models"
	"github.com/udistrital/utils_oas/request"
)

const userInfoURL = "https://autenticacion.portaloas.udistrital.edu.co/oauth2/userinfo"

var userInfoRequestMutex sync.Mutex

func RegistrarAuditoria(entidadTipo string, entidadId int, operacion string, valorAnterior, valorNuevo interface{}, usuario, endpoint string) {
	payload := models.AuditoriaEventoPayload{
		EntidadTipo:   entidadTipo,
		Operacion:     operacion,
		ValorAnterior: helpers.ToJSONOrNil(valorAnterior),
		ValorNuevo:    helpers.ToJSONOrNil(valorNuevo),
		TerceroId:     helpers.TerceroIDOrNil(usuario),
		Endpoint:      endpoint,
		Activo:        true,
	}

	var resultado interface{}
	request.SendJson(beego.AppConfig.String("EventoService")+"auditoria_eventos", "POST", &resultado, payload)
	_ = resultado
}

func ExtraerUsuario(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	if terceroID := resolverTerceroDesdeBearer(authHeader); terceroID != "" {
		return terceroID
	}
	if sub := helpers.ExtraerSubJWT(authHeader); sub != "" {
		return sub
	}
	return ""
}

func resolverTerceroDesdeBearer(authHeader string) string {
	userInfo := userInfoDesdeBearer(authHeader)
	if userInfo == nil {
		return ""
	}

	sub, _ := userInfo["sub"].(string)
	email, _ := userInfo["email"].(string)
	documento, _ := userInfo["documento"].(string)
	if id := buscarTerceroPorDocumento(documento, sub, email); id != "" {
		return id
	}
	if id := buscarTerceroPorUsuarioWSO2(sub); id != "" {
		return id
	}
	return buscarTerceroPorUsuarioWSO2(email)
}

func userInfoDesdeBearer(authHeader string) map[string]interface{} {
	var userInfo map[string]interface{}
	userInfoRequestMutex.Lock()
	request.SetHeader(authHeader)
	err := request.GetJson(userInfoURL, &userInfo)
	request.SetHeader("")
	userInfoRequestMutex.Unlock()
	if err != nil || userInfo == nil {
		return nil
	}
	return userInfo
}

func buscarTerceroPorDocumento(documento string, usuario string, correo string) string {
	if documento == "" {
		return ""
	}
	var identificaciones []map[string]interface{}
	endpoint := beego.AppConfig.String("TercerosService") + "datos_identificacion?query=Activo:true,Numero:" + url.QueryEscape(documento) + "&sortby=FechaCreacion&order=desc"
	if err := request.GetJson(endpoint, &identificaciones); err != nil || len(identificaciones) == 0 {
		return ""
	}
	for _, identificacion := range identificaciones {
		tercero, ok := identificacion["TerceroId"].(map[string]interface{})
		if !ok || tercero == nil {
			continue
		}
		usuarioWSO2, _ := tercero["UsuarioWSO2"].(string)
		if usuarioWSO2 == usuario || usuarioWSO2 == correo {
			if id, ok := helpers.IDToString(tercero["Id"]); ok {
				return id
			}
		}
	}
	if tercero, ok := identificaciones[0]["TerceroId"].(map[string]interface{}); ok {
		if id, ok := helpers.IDToString(tercero["Id"]); ok {
			return id
		}
	}
	return ""
}

func buscarTerceroPorUsuarioWSO2(usuario string) string {
	if usuario == "" {
		return ""
	}
	var terceros []map[string]interface{}
	endpoint := beego.AppConfig.String("TercerosService") + "tercero?query=UsuarioWSO2:" + url.QueryEscape(usuario)
	if err := request.GetJson(endpoint, &terceros); err != nil || len(terceros) == 0 {
		return ""
	}
	if id, ok := helpers.IDToString(terceros[0]["Id"]); ok {
		return id
	}
	return ""
}
