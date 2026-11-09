# Implementación del despachador - Referencias y justificación

**Asignatura:** Lenguajes de Programación II  
**Responsable:** Nicolás Flores  
**Lenguaje:** Go

## 1. Objetivo

Implementar un despachador concurrente que reciba solicitudes de camiones y asigne destinos de tipo pala o stockpile mediante una política Round Robin, comprobando previamente que exista una ruta válida.

## 2. Justificación de la implementación

### Goroutines y canales

Se utilizó una goroutine para que el despachador pueda ejecutarse concurrentemente con los demás componentes del simulador, las solicitudes se reciben mediante un canal de Go. Cada solicitud contiene otro canal para devolver el resultado al camión correspondiente

**Referencias:**
- https://go.dev/tour/concurrency/1
- https://go.dev/tour/concurrency/2

### Política Round Robin

Se eligió Round Robin por su simplicidad y porque permite repartir las asignaciones entre distintos destinos mediante un turno circular, se mantienen contadores independientes para las palas y los stockpiles. Los destinos se ordenan por identificador para evitar depender del orden de recorrido de los mapas de Go

**Referencia sobre ordenamiento:**
- https://pkg.go.dev/sort#Strings

La elección de Round Robin corresponde a una decisión de diseño para cumplir los requisitos del certamen.

### Validación de rutas

Antes de asignar un destino se utiliza el método `CalcularRuta()` del paquete `internal/ruteo`.

Este método emplea el algoritmo Dijkstra incorporado previamente al proyecto y considera los tramos marcados como fallidos.

Si no existe una ruta válida, el despachador intenta con otro destino. Si ninguno es alcanzable, devuelve un error.

**Fuente:** `simulador/internal/ruteo/planificador.go`, implementación existente del proyecto.

### Control de ejecución

Se utilizó `context.Context` para permitir la cancelación del despachador y `select` para atender solicitudes o finalizar su ejecución.

**Referencias:**
- https://pkg.go.dev/context
- https://go.dev/tour/concurrency/5

## 3. Pruebas realizadas

Se implementaron cinco pruebas unitarias:

1. Alternancia Round Robin entre palas.
2. Asignación de stockpile a un camión cargado.
3. Exclusión de un destino inalcanzable.
4. Manejo de ausencia de rutas disponibles.
5. Comunicación con la goroutine del despachador.

Los comandos `go test ./...` y `go vet ./...` finalizaron correctamente.

La prueba con `go test -race` quedó pendiente porque el entorno Windows requiere habilitar CGO y disponer de un compilador C compatible.

## 4. Uso de inteligencia artificial

Se utilizó ChatGPT como apoyo para generar pruebas, explicar las herramientas de concurrencia de Go y revisar los resultados de compilación.

## 5. Limitaciones y trabajo pendiente

La implementación actual no controla la ocupación ni las colas de las palas, y todavía debe integrarse con las goroutines reales de los camiones, también queda pendiente verificar su comportamiento ante cambios concurrentes del mapa y ejecutar el detector de condiciones de carrera.