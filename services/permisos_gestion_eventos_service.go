package services

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/astaxie/beego"
	"github.com/udistrital/sga_calendario_mid/helpers"
	"github.com/udistrital/utils_oas/request"
)

const userInfoURL = "https://autenticacion.portaloas.udistrital.edu.co/oauth2/userinfo"

var userInfoRequestMutex sync.Mutex

func ValidarPermisoGestionActividad(idActividad string, authHeader string) error {
	actividad, err := obtenerActividad(idActividad)
	if err != nil {
		return err
	}
	eventoCatalogoID, err := idRelacion(actividad["EventoCatalogoId"])
	if err != nil || eventoCatalogoID == "" {
		return errors.New("no fue posible identificar el catálogo de la actividad")
	}
	return ValidarPermisoGestionEventoCatalogo(eventoCatalogoID, authHeader)
}

func ValidarPermisoGestionCalendario(idCalendario string, authHeader string) error {
	roles := rolesDesdeAuthorization(authHeader)
	perfilesUsuario, err := perfilesUsuarioDesdeRoles(roles, authHeader)
	if err != nil {
		return err
	}
	if len(perfilesUsuario) == 0 {
		return errors.New("no tiene perfiles autorizados para gestionar actividades del calendario")
	}

	var eventos []map[string]interface{}
	urlEventos := beego.AppConfig.String("EventoService") + "calendario_evento?query=Activo:true,ProcesoId__CalendarioID__Id:" + idCalendario + "&limit=0"
	if err := request.GetJson(urlEventos, &eventos); err != nil {
		return errors.New("no fue posible consultar actividades del calendario para validar permisos")
	}
	for _, evento := range eventos {
		eventoCatalogoID, err := idRelacion(evento["EventoCatalogoId"])
		if err != nil || eventoCatalogoID == "" {
			return errors.New("no fue posible identificar el catálogo de una actividad del calendario")
		}
		if !perfilGestionaEventoCatalogo(eventoCatalogoID, perfilesUsuario) {
			return fmt.Errorf("no tiene permiso para gestionar actividades del catálogo %s", eventoCatalogoID)
		}
	}
	return nil
}

func ValidarPermisoGestionEventoCatalogo(eventoCatalogoID string, authHeader string) error {
	roles := rolesDesdeAuthorization(authHeader)
	perfilesUsuario, err := perfilesUsuarioDesdeRoles(roles, authHeader)
	if err != nil {
		return err
	}
	if len(perfilesUsuario) == 0 || !perfilGestionaEventoCatalogo(eventoCatalogoID, perfilesUsuario) {
		return errors.New("no tiene permiso para gestionar esta actividad")
	}
	return nil
}

func rolesDesdeAuthorization(authHeader string) []string {
	roles := make([]string, 0)
	roles = append(roles, rolesDesdeUserInfo(userInfoDesdeBearer(authHeader))...)
	roles = append(roles, rolesDesdeJWT(authHeader)...)
	return roles
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

func rolesDesdeUserInfo(userInfo map[string]interface{}) []string {
	roles := make([]string, 0)
	if userInfo == nil {
		return roles
	}
	roles = append(roles, rolesDesdeCampo(userInfo["role"])...)
	roles = append(roles, rolesDesdeCampo(userInfo["roles"])...)
	if userService, ok := userInfo["userService"].(map[string]interface{}); ok {
		roles = append(roles, rolesDesdeCampo(userService["role"])...)
	}
	if user, ok := userInfo["user"].(map[string]interface{}); ok {
		roles = append(roles, rolesDesdeCampo(user["role"])...)
	}
	return filtrarRolesAplicacion(roles)
}

func rolesDesdeJWT(authHeader string) []string {
	claims, ok := helpers.ClaimsDesdeBearer(authHeader)
	if !ok {
		return []string{}
	}
	roles := append(rolesDesdeCampo(claims["role"]), rolesDesdeCampo(claims["roles"])...)
	return filtrarRolesAplicacion(roles)
}

func rolesDesdeCampo(valor interface{}) []string {
	switch typed := valor.(type) {
	case string:
		if typed == "" {
			return []string{}
		}
		return []string{typed}
	case []interface{}:
		roles := make([]string, 0, len(typed))
		for _, item := range typed {
			if rol := strings.TrimSpace(fmt.Sprintf("%v", item)); rol != "" {
				roles = append(roles, rol)
			}
		}
		return roles
	case []string:
		return typed
	default:
		return []string{}
	}
}

func filtrarRolesAplicacion(roles []string) []string {
	filtrados := make([]string, 0, len(roles))
	for _, rol := range roles {
		rol = strings.TrimSpace(rol)
		if rol != "" && !strings.Contains(rol, "/") {
			filtrados = append(filtrados, rol)
		}
	}
	return filtrados
}

func perfilesUsuarioDesdeRoles(roles []string, authHeader string) (map[int]bool, error) {
	perfiles, err := PerfilesConfiguracionSGA(authHeader)
	if err != nil {
		return nil, fmt.Errorf("no fue posible consultar perfiles de gestión: %w", err)
	}
	rolesNormalizados := make(map[string]bool)
	for _, rol := range roles {
		rolesNormalizados[helpers.NormalizarTextoPermiso(rol)] = true
	}
	resultado := make(map[int]bool)
	for _, perfil := range perfiles {
		if perfil.Id <= 0 {
			continue
		}
		nombre := helpers.NormalizarTextoPermiso(perfil.Nombre)
		codigo := helpers.NormalizarTextoPermiso(perfil.CodigoAbreviacion)
		if rolesNormalizados[nombre] || rolesNormalizados[codigo] {
			resultado[perfil.Id] = true
		}
	}
	return resultado, nil
}

func perfilGestionaEventoCatalogo(eventoCatalogoID string, perfilesUsuario map[int]bool) bool {
	var relaciones []map[string]interface{}
	url := beego.AppConfig.String("EventoService") + "evento_catalogo_rol_gestion?query=Activo:true,EventoCatalogoId__Id:" + eventoCatalogoID + "&limit=0"
	if err := request.GetJson(url, &relaciones); err != nil {
		return false
	}
	for _, relacion := range relaciones {
		perfilID, ok := helpers.InterfaceToInt(relacion["PerfilId"])
		if ok && perfilesUsuario[perfilID] {
			return true
		}
	}
	return false
}
