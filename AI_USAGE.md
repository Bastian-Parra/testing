# AI_USAGE.md

## Registro de código externo

**Issue:** #15 - Fase 1: Importación pura del algoritmo de ruteo.

**Responsable:** Nicolás Flores.

**Algoritmo:** Dijkstra.

**Repositorio original:** https://github.com/RyanCarrier/dijkstra

**Autoría original:** Ryan Carrier y colaboradores.

**Licencia:** MIT.

**Lenguaje:** Go.

### Archivos importados

- `simulador/planners/externos/dijkstra/dijkstra.go`
- `simulador/planners/externos/dijkstra/graph.go`
- `simulador/planners/externos/dijkstra/list.go`
- `simulador/planners/externos/dijkstra/errors.go`
- `simulador/planners/externos/dijkstra/LICENSE`

### Procedimiento de importación

Se descargó el repositorio original y se seleccionaron los archivos necesarios para utilizar el algoritmo de Dijkstra. Se agregaron comentarios de procedencia al inicio de los archivos Go y se conservó la licencia original.

La lógica del algoritmo importado no fue modificada.

### Uso de inteligencia artificial

Se utilizó ChatGPT como apoyo para comprender los requisitos de importación, interpretar los comandos de Git y Go, y orientar las verificaciones de compilación.

El código del algoritmo fue obtenido del repositorio externo indicado y no generado mediante inteligencia artificial.

### Commit de importación

`d33cf232f3729662e5e13ca81370b788642ff41f`

### Verificación

Se comprobó la compilación de los cuatro archivos Go importados mediante `go test`, sin errores de compilación. No se ejecutaron pruebas unitarias en esta verificación.