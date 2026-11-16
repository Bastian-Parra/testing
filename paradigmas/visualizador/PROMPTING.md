# Registro de prompting — Certamen III

## Iteración 1: Diseño POO y lectura de logs

### Objetivo
Generar una primera versión del visualizador utilizando Python 3.11 y programación orientada a objetos.

### Prompt utilizado
Genera en Python 3.11 un visualizador orientado a objetos para un simulador de mina. Utiliza una clase base Equipo y clases Camion, Pala y Stockpile. Representa los eventos mediante una clase Evento. Lee un CSV consolidado con las columnas Seq,Tick,Entidad,Evento,Recurso,Toneladas, manteniendo el orden global de Seq. Permite toneladas vacías, detecta secuencias duplicadas y no inventes estados ni posiciones cuando los eventos no entreguen información suficiente. Separa la lectura de datos de la lógica de simulación.

### Código generado
- `modelos.py`: clases de equipos y eventos.
- `lector_logs.py`: lectura y validación del CSV.
- `simulacion.py`: procesamiento de eventos.
- `main.py`: ejecución del programa.

### Prueba realizada
Comando:

`python paradigmas/visualizador/main.py paradigmas/visualizador/prueba.csv`

Resultado:

```text
Eventos cargados: 6
Equipos encontrados: 2
C-01: estado=sale, recurso=T1_A->B
C-02: estado=entra, recurso=T2_A->B
```

### Análisis crítico
La primera versión consiguió leer el CSV de prueba e identificar correctamente los dos camiones.

Sin embargo, encontramos las siguientes limitaciones:

- El estado del equipo corresponde directamente al nombre del último evento.
- No se representan posiciones físicas ni desplazamientos.
- Todavía no se distinguen correctamente los estados operacionales de los camiones.
- No se ha validado el funcionamiento con un log consolidado real.
- Aún no existen animación, diagrama de Gantt ni gráficos comparativos.

### Correcciones solicitadas para la siguiente iteración
Mejorar la interpretación de los eventos, conservar el historial de estados por tick y desarrollar una primera visualización de los movimientos y fallas.

### Estado
Primera prueba ejecutada correctamente con datos ficticios. Validación completa pendiente.