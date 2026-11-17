DEVLOG - C2 & C3

# Semana 1 (20 - 25 oct)
**Avance realizado:**
- Lectura y análisis de los enunciados de los certámenes 2 y 3.
- Adaptación y configuración del repositorio ya creado con anterioridad para el certamen 1 con el objetivo de cumplir con las evaluaciones 2 y 3.
- Definición de la arquitectura base del simulador, roles del equipo y elaboración de la estructura inicial de directorios.

**Problemas encontrados:**
- Tuvimos dudas sobre la correcta vinculación de los commits con los usuarios de GitHub para asegurar que el trabajo de todos quedara registrado según la rúbrica.

**Solución:**
- Los tres integrantes configuraron correstamente sus perfiles mediante `git config` y subieron los acuses de recibo modificando el archivo COMPROMISO.md y estructurando el README.md

# Semana 2 (26 oct - 01 nov)
**Avance realizado:**
- Desarrollo de la base y pasos iniciales del simulador.
- Luca implementó el lector del archivo del mapa, la tabla de símbolos y creó la primera iteración de la goroutine de los camiones con su máquina de estados, además de agregar la lógica de los protocolos de acceso a las rampas.
- Nicolás se encargó de investigar e importar limpiamente el algoritmo de Dijkstra desde una fuente externa, documentando los orígenes.
- Bastian adaptó el algoritmo de Dijkstra para que el planificador aplique correctamente los pesos por carga y excluya de las rutas los tramos con fallas permanentes.

**Problemas encontrados:**
- Al adaptar Dijkstra, notamos que si no extraíamos bien la información de la tabla de símbolos concurrente, el cálculo de rutas fallaba al encontrarse con tramos temporalmente inactivos o nodos sin inicializar.

**Solución:**
- Se ajustó el adaptador de Dijkstra para leer correctamente el estado actual de la topología (pesos variables) y filtrar explícitamente los tramos marcados con fallas permanentes antes de iniciar la búsqueda del camino más corto.

# Semana 3 (02 - 08 nov)
**Avance realizado:**
- Implementación de la coordinación central.
- Nicolás diseñó e implementó el despachador concurrente utilizando una política de asignación Round-Robin.
- Bastián desarrolló las goroutines independientes para las Palas y los Stockpiles, integrando la comunicación pura por canales para evitar bloqueos.

**Problemas encontrados:**
- En las primeras pruebas de integración entre las goroutines de los camiones y las palas, los camiones se quedaban atascados esperando que el canal de la pala estuviera listo, lo que generaba un cuello de botella y rompía el ciclo.

**Solución:**
- Se rediseño el ciclo principal de los camiones utilizando la instrucción `select` con comandos con guarda. Se deshabilitaron dinámicamente los casos (asignando `nil` al canal) cuando el camión aún no cumplía los requisitos para interactuar con la pala o el despachador.

# Seamana 4 (9 al 15 de noviembre)
**Avance realizado:**
- Cierre de la funcionalidad concurrente e inicio de la Fase 3.
- Luca finalizó la fusión de los protocolos de rampa.
- Bastián hizo un arreglo/modificación del código de palas y camiones para generar dinámicamente los logs en formato .csv requeridos para el C3 y creó la estructura de archivos en la carpeta `paradigmas/`. 

**Problemas encontrados:**
- Al correr las pruebas de carga del simulador utilizando la bandera `go run -race`, el detector reportó advertencias críticas de Data Race.
- El error surgía en la arquitectura de fallas, ya que múltiples goroutines de camiones intentaban leer y registrar sus eventos en el `Gestor` y la `Circulacion` al mismo tiempo, colisionando en memoria compartida.
- Además, en los tests FIFO, el camión perdía su estado de "cargado" al transitar entre nodos.

**Solución:**
- Para el data race: Se implementó exlusión mutua en las estructuras de gestión de fallas y secuencias utilizando `sync.Mutex`. Las lecturas y escrituras críticas se protegieron envolviéndolas con `c.mu.Lock()` y `defer.mu.Unlock`, esto garantiza que el acceso sea seguro y eliminó los warnings del compilador.
- Para la pérdida de estado: Se añadio una variable de memoria booleana `c.cargado` en la estructura interna del camión, permitiendo que la máquina de estados retome la ruta de transporte correcta tras esperar en la cola del tramo.
