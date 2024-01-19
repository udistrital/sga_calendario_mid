# Calendario Académico - SGA MID

API MID intermediaria entre el cliente SGA y la API de Eventos con los endpoints requeridos para la gestión del calendario académico.

## Especificaciones Técnicas

### Tecnologías Implementadas y Versiones
* [Golang](https://github.com/udistrital/introduccion_oas/blob/master/instalacion_de_herramientas/golang.md)
* [BeeGo](https://github.com/udistrital/introduccion_oas/blob/master/instalacion_de_herramientas/beego.md)
* [Docker](https://docs.docker.com/engine/install/ubuntu/)
* [Docker Compose](https://docs.docker.com/compose/)

### Variables de Entorno
```shell
CORE_SERVICE: [Servicio API CORE]
DOCUMENTO_SERVICE: [Servicio API DOCUMENTOS]
EVENTOS_SERVICE: [Servicio API EVENTOS]
OIKOS_SERVICE: [Servicio API OIKOS]
PARAMETRO_SERVICE: [Servicio API PARAMETRO]
PROYECTO_ACADEMICO_SERVICE: [Servicio PROYECTO ACADEMICO]
TERCEROS_SERVICE: [Servicio API TERCEROS]
```
**NOTA:** Las variables se pueden ver en el fichero conf/app.conf.


### Ejecución del Proyecto
```shell
#1. Obtener el repositorio con Go
go get github.com/udistrital/sga_mid_calendario_academico

#2. Moverse a la carpeta del repositorio
cd $GOPATH/src/github.com/udistrital/sga_mid_calendario_academico

# 3. Moverse a la rama **develop**
git pull origin develop && git checkout develop

# 4. alimentar todas las variables de entorno que utiliza el proyecto.
SGA_MID_HTTP_PORT=8080 RULER_URL=127.0.0.1:27017 SGA_MID_CALENDARIO_ACADEMICO_SOME_VARIABLE=some_value bee run
```

### Ejecución Dockerfile
```shell
# docker build --tag=sga_mid_calendario_academico . --no-cache
# docker run -p 80:80 sga_mid_calendario_academico
```

### Ejecución docker-compose
```shell
#1. Clonar el repositorio
git clone -b develop https://github.com/udistrital/sga_mid_calendario_academico

#2. Moverse a la carpeta del repositorio
cd sga_mid_calendario_academico

#3. Crear un fichero con el nombre **custom.env**
# En windows ejecutar:* ` ni custom.env`
touch custom.env

#4. Crear la network **back_end** para los contenedores
docker network create back_end

#5. Ejecutar el compose del contenedor
docker-compose up --build

#6. Comprobar que los contenedores estén en ejecución
docker ps
```

### Ejecución Pruebas

Pruebas unitarias
```shell
# En Proceso
```

## Estado CI

| Develop | Relese 0.0.1 | Master |
| -- | -- | -- |
| [![Build Status](https://hubci.portaloas.udistrital.edu.co/api/badges/udistrital/sga_mid_calendario_academico/status.svg?ref=refs/heads/develop)](https://hubci.portaloas.udistrital.edu.co/udistrital/sga_mid_calendario_academico) | [![Build Status](https://hubci.portaloas.udistrital.edu.co/api/badges/udistrital/sga_mid_calendario_academico/status.svg?ref=refs/heads/release/0.0.1)](https://hubci.portaloas.udistrital.edu.co/udistrital/sga_mid_calendario_academico) | [![Build Status](https://hubci.portaloas.udistrital.edu.co/api/badges/udistrital/sga_mid_calendario_academico/status.svg)](https://hubci.portaloas.udistrital.edu.co/udistrital/sga_mid_calendario_academico) |

## Licencia

This file is part of sga_mid_calendario_academico.

sga_mid_calendario_academico is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

sga_mid is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with sga_mid. If not, see https://www.gnu.org/licenses/.
