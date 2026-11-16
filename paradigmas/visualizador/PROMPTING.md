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

----------------------------------------

## Iteración 2: Animación de la mina y representación de fallas

### Objetivo
Generar una interfaz gráfica en Python utilizando tkinter para visualizar los camiones, los tramos y las fallas del simulador.

### Prompt utilizado
Genera una animación de la mina en Python utilizando tkinter. Representa los seis nodos y cinco tramos definidos en mina_prueba.txt. Permite avanzar y retroceder por ticks, mostrar camiones sobre tramos reconocidos y resaltar las fallas temporales y permanentes. Utiliza el CSV de seis columnas y evita inventar posiciones exactas.

### Código generado
- `animacion.py`: interfaz gráfica, mapa de la mina, controles de navegación y representación de eventos.

### Prueba realizada
Comando:

`python paradigmas/visualizador/animacion.py paradigmas/visualizador/prueba.csv`

Se comprobó visualmente el tick 3:

- C-01 aparece asociado a T1.
- C-02 aparece asociado a T2.
- T1 aparece en rojo debido al evento `temporal`.
- El historial muestra cuatro eventos hasta ese tick.

### Errores y limitaciones detectadas
1. Las fallas permanecen activas visualmente aunque posteriormente puedan existir eventos de recuperación.
2. Los camiones se muestran en el último tramo registrado sin comprobar si ya salieron de él.
3. Las posiciones son esquemáticas, no posiciones físicas calculadas.
4. Todavía no se ha probado el visualizador con un log consolidado real.

### Correcciones solicitadas
Implementar un seguimiento de estados por tick, diferenciar las fallas activas de las reparadas y evitar mostrar camiones dentro de tramos que ya abandonaron.

### Estado
Primera prueba gráfica realizada correctamente con datos ficticios. Correcciones y validación con datos reales pendientes.

### Corrección y segunda prueba de la Iteración 2

**Prompt de corrección solicitado a la IA:**

Corrige el visualizador de Python para que reconstruya los estados de los camiones desde el historial de eventos. Diferencia las fallas temporales y permanentes, mantén las temporales durante `reparando`, permite que `avanza` confirme la reanudación del movimiento y elimina la ubicación del camión cuando ocurre `sale`. Conserva la navegación por ticks y no inventes posiciones físicas.

**Problemas corregidos:**

- Los camiones ya no permanecen visualmente dentro de un tramo después de registrar su salida.
- Las fallas temporales se muestran en naranja y las permanentes en rojo.
- El programa reconstruye los estados al seleccionar cada tick.
- Se incorporó un resumen del estado actual de cada camión.

**Pruebas realizadas:**

Se ejecutó:

`python paradigmas/visualizador/animacion.py paradigmas/visualizador/prueba.csv`

Se revisaron visualmente los ticks 0, 1, 2, 3, 4 y 5.

Los resultados coincidieron con los eventos ficticios del CSV: C-01 ingresó a T1, C-02 ingresó a T2, se representó la falla temporal y posteriormente C-01 abandonó T1.

**Evaluación crítica:**

La corrección mejoró la interpretación de los eventos y la representación de fallas. Sin embargo, las posiciones continúan siendo esquemáticas y todavía no se ha comprobado el funcionamiento con un log consolidado real.

El evento `sale` confirma que el camión abandonó el tramo, pero no proporciona el instante exacto en que finalizó la reparación. Por eso, la visualización no debe interpretarse como una reconstrucción física exacta del proceso.

**Resultado:** correcciones verificadas con seis capturas del escenario ficticio.

------------------------------------

