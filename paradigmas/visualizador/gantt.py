
import argparse
import tkinter as tk
from tkinter import ttk

from lector_logs import LectorLogs


# Colores utilizados para identificar los estados.
COLORES = {
    "circulando": "#0d9488",
    "falla temporal": "#f59e0b",
    "reparando": "#eab308",
    "falla permanente": "#dc2626",
    "cargando": "#2563eb",
    "descargando": "#8b5cf6",
    "esperando": "#64748b",
    "fuera del tramo": "#94a3b8",
    "desconocido": "#d1d5db",
    "sin datos": "#f1f5f9",
}

# Estados que indican una detencion o falla.
ESTADOS_FALLA = {
    "falla temporal",
    "reparando",
    "falla permanente",
}


def clasificar_evento(evento):
    # Traducimos los eventos del CSV a estados del Gantt.
    tipo = evento.tipo

    if tipo in ("entra", "avanza"):
        return "circulando"

    if tipo == "temporal":
        return "falla temporal"

    if tipo == "reparando":
        return "reparando"

    if tipo in ("permanente", "panic_permanente"):
        return "falla permanente"

    if tipo == "carga":
        return "cargando"

    if tipo == "descarga":
        return "descargando"

    if tipo in ("sale", "libera"):
        return "fuera del tramo"

    if tipo == "adquiere":
        return "esperando"

    return None


def unir_segmentos(segmentos):
    # Unimos bloques consecutivos del mismo estado.
    if not segmentos:
        return []

    unidos = []

    for inicio, fin, estado in segmentos:
        if fin <= inicio:
            continue

        if unidos:
            anterior_inicio, anterior_fin, anterior_estado = unidos[-1]

            if anterior_estado == estado and anterior_fin == inicio:
                unidos[-1] = (
                    anterior_inicio,
                    fin,
                    estado
                )
                continue

        unidos.append((inicio, fin, estado))

    return unidos


def construir_intervalos(eventos):
    # Agrupamos los eventos por camion.
    por_camion = {}

    for evento in sorted(eventos, key=lambda e: (e.tick, e.seq)):
        if not evento.entidad.startswith("C-"):
            continue

        por_camion.setdefault(evento.entidad, []).append(evento)

    if not por_camion:
        return {}, 0, 0

    tick_min = min(
        evento.tick
        for lista in por_camion.values()
        for evento in lista
    )

    tick_max = max(
        evento.tick
        for lista in por_camion.values()
        for evento in lista
    )

    # Agregamos un tick al final solo para representar
    # visualmente el ultimo evento.
    fin_visual = tick_max + 1

    intervalos = {}

    for camion, lista in sorted(por_camion.items()):
        lista = sorted(lista, key=lambda e: (e.tick, e.seq))

        segmentos = []
        estado_actual = "desconocido"
        inicio = tick_min

        for evento in lista:
            nuevo_estado = clasificar_evento(evento)

            if nuevo_estado is None:
                continue

            # Si el estado sigue siendo el mismo,
            # no necesitamos abrir otro bloque.
            if nuevo_estado == estado_actual:
                continue

            # Guardamos el intervalo anterior.
            if evento.tick > inicio:
                segmentos.append(
                    (inicio, evento.tick, estado_actual)
                )

            inicio = evento.tick
            estado_actual = nuevo_estado

        # El ultimo estado solo se dibuja durante un tick.
        # No afirmamos que duro hasta el final del turno.
        fin_ultimo = min(inicio + 1, fin_visual)

        if fin_ultimo > inicio:
            segmentos.append(
                (inicio, fin_ultimo, estado_actual)
            )

        # Si quedan ticks sin observaciones posteriores,
        # los mostramos como un periodo sin datos.
        if fin_ultimo < fin_visual:
            segmentos.append(
                (fin_ultimo, fin_visual, "sin datos")
            )

        intervalos[camion] = unir_segmentos(segmentos)

    return intervalos, tick_min, fin_visual


class DiagramaGantt:
    def __init__(self, eventos):
        self.intervalos, self.inicio, self.fin = construir_intervalos(
            eventos
        )

        self.ventana = tk.Tk()
        self.ventana.title("Certamen III - Diagrama de Gantt")
        self.ventana.geometry("1100x650")

        titulo = ttk.Label(
            self.ventana,
            text="Estados de los camiones durante la simulacion",
            font=("Arial", 15, "bold")
        )
        titulo.pack(pady=12)

        descripcion = ttk.Label(
            self.ventana,
            text=(
                "Intervalos inferidos desde eventos del CSV. "
                "Los periodos sin observaciones se indican como sin datos."
            )
        )
        descripcion.pack(pady=3)

        contenedor = ttk.Frame(self.ventana)
        contenedor.pack(
            fill="both",
            expand=True,
            padx=15,
            pady=10
        )

        self.canvas = tk.Canvas(
            contenedor,
            bg="white",
            height=400
        )

        scroll_y = ttk.Scrollbar(
            contenedor,
            orient="vertical",
            command=self.canvas.yview
        )

        scroll_x = ttk.Scrollbar(
            contenedor,
            orient="horizontal",
            command=self.canvas.xview
        )

        self.canvas.configure(
            yscrollcommand=scroll_y.set,
            xscrollcommand=scroll_x.set
        )

        self.canvas.grid(
            row=0,
            column=0,
            sticky="nsew"
        )

        scroll_y.grid(
            row=0,
            column=1,
            sticky="ns"
        )

        scroll_x.grid(
            row=1,
            column=0,
            sticky="ew"
        )

        contenedor.rowconfigure(0, weight=1)
        contenedor.columnconfigure(0, weight=1)

        # Leyenda de colores.
        leyenda = ttk.Frame(self.ventana)
        leyenda.pack(pady=10)

        for estado, color in COLORES.items():
            color_texto = "white"

            if estado in ("desconocido", "sin datos"):
                color_texto = "black"

            tk.Label(
                leyenda,
                text=estado,
                bg=color,
                fg=color_texto,
                padx=7,
                pady=4
            ).pack(
                side="left",
                padx=3
            )

        self.dibujar()

    def dibujar(self):
        self.canvas.delete("all")

        if not self.intervalos:
            self.canvas.create_text(
                300,
                100,
                text="No se encontraron eventos de camiones."
            )
            return

        margen_x = 110
        margen_y = 70
        alto_fila = 65
        alto_barra = 30

        duracion = max(
            1,
            self.fin - self.inicio
        )

        # Para turnos grandes utilizamos desplazamiento.
        escala = max(
            2,
            min(65, 850 / duracion)
        )

        ancho = max(
            950,
            margen_x + duracion * escala + 100
        )

        alto = max(
            400,
            margen_y + len(self.intervalos) * alto_fila + 80
        )

        # Dibujamos el eje horizontal de ticks.
        salto = max(
            1,
            duracion // 12
        )

        for tick in range(
            self.inicio,
            self.fin + 1,
            salto
        ):
            x = margen_x + (tick - self.inicio) * escala

            self.canvas.create_line(
                x,
                45,
                x,
                alto - 40,
                fill="#e2e8f0",
                dash=(3, 4)
            )

            self.canvas.create_text(
                x,
                30,
                text=str(tick),
                fill="#475569"
            )

        # Una fila por camion.
        for indice, (camion, segmentos) in enumerate(
            self.intervalos.items()
        ):
            y = margen_y + indice * alto_fila

            self.canvas.create_text(
                65,
                y + alto_barra / 2,
                text=camion,
                font=("Arial", 11, "bold")
            )

            for inicio, fin, estado in segmentos:
                if fin <= inicio:
                    continue

                x1 = margen_x + (
                    inicio - self.inicio
                ) * escala

                x2 = margen_x + (
                    fin - self.inicio
                ) * escala

                color = COLORES.get(
                    estado,
                    COLORES["desconocido"]
                )

                self.canvas.create_rectangle(
                    x1,
                    y,
                    x2,
                    y + alto_barra,
                    fill=color,
                    outline="white"
                )

                # Solo escribimos la etiqueta si cabe.
                # Asi evitamos textos superpuestos.
                ancho_bloque = x2 - x1
                ancho_minimo = max(
                    75,
                    len(estado) * 7
                )

                if ancho_bloque >= ancho_minimo:
                    color_texto = "white"

                    if estado in ("desconocido", "sin datos"):
                        color_texto = "black"

                    self.canvas.create_text(
                        (x1 + x2) / 2,
                        y + alto_barra / 2,
                        text=estado,
                        fill=color_texto,
                        font=("Arial", 9)
                    )

        self.canvas.configure(
            scrollregion=(0, 0, ancho, alto)
        )

    def ejecutar(self):
        self.ventana.mainloop()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Diagrama de Gantt de los camiones"
    )

    parser.add_argument(
        "archivo",
        help="Ruta del CSV consolidado"
    )

    argumentos = parser.parse_args()

    eventos = LectorLogs().leer(argumentos.archivo)

    gantt = DiagramaGantt(eventos)
    gantt.ejecutar()
