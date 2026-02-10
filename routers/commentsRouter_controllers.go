package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

    beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ActividadCalendarioController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ActividadCalendarioController"],
        beego.ControllerComments{
            Method: "PostActividadCalendario",
            Router: "/",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ActividadCalendarioController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ActividadCalendarioController"],
        beego.ControllerComments{
            Method: "UpdateActividadResponsables",
            Router: "/calendario/actividad/:id",
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ClonarCalendarioController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ClonarCalendarioController"],
        beego.ControllerComments{
            Method: "PostCalendario",
            Router: "/",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ClonarCalendarioController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ClonarCalendarioController"],
        beego.ControllerComments{
            Method: "PostCalendarioExtension",
            Router: "/extension",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ClonarCalendarioController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ClonarCalendarioController"],
        beego.ControllerComments{
            Method: "PostCalendarioPadre",
            Router: "/padre",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ConsultaCalendarioAcademicoController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ConsultaCalendarioAcademicoController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ConsultaCalendarioAcademicoController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ConsultaCalendarioAcademicoController"],
        beego.ControllerComments{
            Method: "GetOnePorId",
            Router: "/:id",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ConsultaCalendarioAcademicoController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ConsultaCalendarioAcademicoController"],
        beego.ControllerComments{
            Method: "PutInhabilitarCalendario",
            Router: "/calendario/academico/:id/inhabilitar",
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ConsultaCalendarioAcademicoController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ConsultaCalendarioAcademicoController"],
        beego.ControllerComments{
            Method: "PostCalendarioHijo",
            Router: "/padre",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ConsultaCalendarioAcademicoController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ConsultaCalendarioAcademicoController"],
        beego.ControllerComments{
            Method: "GetCalendarInfo",
            Router: "/v2/:id",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ConsultaCalendarioProyectoController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ConsultaCalendarioProyectoController"],
        beego.ControllerComments{
            Method: "GetCalendarByProjectId",
            Router: "/:id",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ConsultaCalendarioProyectoController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:ConsultaCalendarioProyectoController"],
        beego.ControllerComments{
            Method: "GetCalendarProject",
            Router: "/calendario/proyecto",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:EventoController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:EventoController"],
        beego.ControllerComments{
            Method: "PostEvento",
            Router: "/",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:EventoController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:EventoController"],
        beego.ControllerComments{
            Method: "PutEvento",
            Router: "/:id",
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:EventoController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:EventoController"],
        beego.ControllerComments{
            Method: "DeleteEvento",
            Router: "/:id",
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:EventoController"] = append(beego.GlobalControllerRouter["github.com/udistrital/sga_calendario_mid/controllers:EventoController"],
        beego.ControllerComments{
            Method: "GetEvento",
            Router: "/evento/persona/:persona",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
