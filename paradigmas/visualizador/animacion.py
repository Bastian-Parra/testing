
import argparse
import tkinter as tk
from tkinter import ttk

from lector_logs import LectorLogs


# Coordenadas esquematicas del mapa mina_prueba.txt.
# Estas posiciones son solo para visualizar la mina.
NODOS = {
    "E1": (90, 210),
    "I1": (250, 210),
    "I2": (430, 210),
    "S1": (250, 70),
    "P1": (370, 350),
    "P2": (510, 350),
}

# Tramos definidos en el archivo de la mina.
TRAMOS = {
    "T1": ("E1", "I1"),
    "T2": ("I1", "S1"),
    "T3": ("I1", "I2"),
    "R1": ("I2", "P1"),
    "R2": ("I2", "P2"),
}


def identificar_tramo(recurso):
    # Reconocemos T1 y tambien formatos como T1_A->B.
    # No inventamos equivalencias para recursos desconocidos.
    nombre = (recurso or "").split("_", 1)[0]
    return nombre if nombre in TRAMOS else None


class AnimacionMina:
    def __init__(self, eventos):
        self.eventos = sorted(eventos, key=lambda e: e.seq)

        # Mostramos el tick inicial y los ticks registrados.
        self.ticks = sorted({0} | {e.tick for e in eventos})
        self.indice = 0

        self.ventana = tk.Tk()
        self.ventana.title("Certamen III - Visualizador de mina")
        self.ventana.geometry("900x650")

        # Controles para navegar entre los ticks.
        controles = ttk.Frame(self.ventana)
        controles.pack(pady=10)

        ttk.Button(
            controles,
            text="Anterior",
            command=self.anterior
        ).pack(side="left", padx=5)

        self.etiqueta_tick = ttk.Label(controles)
        self.etiqueta_tick.pack(side="left", padx=15)

        ttk.Button(
            controles,
            text="Siguiente",
            command=self.siguiente
        ).pack(side="left", padx=5)

        # Espacio donde dibujamos el mapa.
        self.canvas = tk.Canvas(
            self.ventana,
            width=850,
            height=420,
            bg="white"
        )
        self.canvas.pack(padx=10, pady=10)

        # Historial de eventos observados.
        self.detalle = tk.Text(
            self.ventana,
            height=9,
            width=105,
            state="disabled"
        )
        self.detalle.pack(padx=10, pady=5)

        self.dibujar()

    def anterior(self):
        if self.indice > 0:
            self.indice -= 1
            self.dibujar()

    def siguiente(self):
        if self.indice < len(self.ticks) - 1:
            self.indice += 1
            self.dibujar()

    def reconstruir_estados(self, tick_actual):
        # Reconstruimos todo desde el comienzo.
        # Asi los botones funcionan en ambas direcciones.
        estados = {}
        fallas = {}
        visibles = []

        for evento in self.eventos:
            if evento.tick > tick_actual:
                continue

            visibles.append(evento)

            # Por ahora seguimos los estados de los camiones.
            if not evento.entidad.startswith("C-"):
                continue

            camion = evento.entidad
            tipo = evento.tipo
            recurso = evento.recurso
            tramo = identificar_tramo(recurso)

            if camion not in estados:
                estados[camion] = {
                    "tramo": None,
                    "estado": "desconocido"
                }

            estado = estados[camion]

            if tipo == "entra":
                estado["tramo"] = tramo
                estado["estado"] = "circulando"

            elif tipo == "adquiere":
                # Adquiere puede referirse a una pala u otro
                # recurso. Solo ubicamos el camion si el
                # recurso corresponde a un tramo conocido.
                if tramo is not None:
                    estado["tramo"] = tramo
                estado["estado"] = "adquiere recurso"

            elif tipo == "avanza":
                estado["tramo"] = tramo
                estado["estado"] = "circulando"

                # Si vuelve a avanzar despues de una falla
                # temporal, consideramos que se recupero.
                falla = fallas.get(camion)

                if falla is not None:
                    if falla["tipo"] == "temporal":
                        if falla["recurso"] == recurso:
                            fallas.pop(camion, None)

            elif tipo == "temporal":
                estado["tramo"] = tramo
                estado["estado"] = "falla temporal"

                fallas[camion] = {
                    "tipo": "temporal",
                    "recurso": recurso,
                    "tramo": tramo
                }

            elif tipo == "reparando":
                # Reparando no significa que la reparacion
                # ya haya terminado.
                estado["estado"] = "reparando"

            elif tipo in ("permanente", "panic_permanente"):
                estado["tramo"] = tramo
                estado["estado"] = "falla permanente"

                fallas[camion] = {
                    "tipo": "permanente",
                    "recurso": recurso,
                    "tramo": tramo
                }

            elif tipo == "sale":
                # Sale confirma que abandono el tramo.
                estado["tramo"] = None
                estado["estado"] = "fuera del tramo"

                # Si salio despues de una falla temporal,
                # ya no sigue detenido en ese tramo.
                falla = fallas.get(camion)

                if falla is not None:
                    if falla["tipo"] == "temporal":
                        if falla["recurso"] == recurso:
                            fallas.pop(camion, None)

            elif tipo == "libera":
                # Libera no siempre significa salir de un
                # tramo, porque puede liberar una pala.
                if tramo is not None:
                    if estado["tramo"] == tramo:
                        estado["tramo"] = None

                estado["estado"] = "libera recurso"

            elif tipo == "carga":
                estado["estado"] = "cargando"

            elif tipo == "descarga":
                estado["estado"] = "descargando"

            elif tipo == "sin_ruta":
                estado["estado"] = "sin ruta"

        return estados, fallas, visibles

    def dibujar_mapa(self):
        # Dibujamos los tramos.
        for nombre, (origen, destino) in TRAMOS.items():
            x1, y1 = NODOS[origen]
            x2, y2 = NODOS[destino]

            self.canvas.create_line(
                x1, y1, x2, y2,
                width=4,
                fill="#94a3b8"
            )

            self.canvas.create_text(
                (x1 + x2) / 2,
                (y1 + y2) / 2 - 12,
                text=nombre,
                fill="#475569"
            )

        # Dibujamos los nodos.
        for nombre, (x, y) in NODOS.items():
            color = "#2563eb"

            if nombre.startswith("P"):
                color = "#16a34a"

            elif nombre.startswith("S"):
                color = "#d97706"

            self.canvas.create_oval(
                x - 20, y - 20,
                x + 20, y + 20,
                fill=color,
                outline=""
            )

            self.canvas.create_text(
                x, y,
                text=nombre,
                fill="white"
            )

    def dibujar_fallas(self, fallas):
        # Naranja: falla temporal.
        # Rojo: falla permanente.
        # El color identifica el tramo afectado, pero
        # no implica bloqueo en ambos sentidos.
        for falla in fallas.values():
            tramo = falla["tramo"]

            if tramo is None:
                continue

            origen, destino = TRAMOS[tramo]
            x1, y1 = NODOS[origen]
            x2, y2 = NODOS[destino]

            if falla["tipo"] == "permanente":
                color = "#dc2626"
            else:
                color = "#f59e0b"

            self.canvas.create_line(
                x1, y1, x2, y2,
                width=6,
                fill=color
            )

    def dibujar_camiones(self, estados):
        # Dibujamos los camiones que siguen asociados
        # a un tramo conocido.
        for indice, (nombre, estado) in enumerate(
            sorted(estados.items())
        ):
            tramo = estado["tramo"]

            if tramo is None:
                continue

            origen, destino = TRAMOS[tramo]
            x1, y1 = NODOS[origen]
            x2, y2 = NODOS[destino]

            # Posicion esquematica en el tramo.
            x = (x1 + x2) / 2
            y = (y1 + y2) / 2 + 16 + indice * 18

            self.canvas.create_rectangle(
                x - 24, y - 10,
                x + 24, y + 10,
                fill="#0f766e",
                outline=""
            )

            self.canvas.create_text(
                x, y,
                text=nombre,
                fill="white",
                font=("Arial", 9)
            )

    def mostrar_historial(self, tick_actual, visibles, estados):
        lineas = [
            f"Eventos observados hasta el tick {tick_actual}:",
            ""
        ]

        for evento in visibles[-12:]:
            lineas.append(
                f"Seq {evento.seq} | Tick {evento.tick} | "
                f"{evento.entidad} | {evento.tipo} | "
                f"{evento.recurso}"
            )

        lineas.append("")
        lineas.append("Estados actuales:")

        for nombre, estado in sorted(estados.items()):
            ubicacion = estado["tramo"] or "sin tramo conocido"

            lineas.append(
                f"{nombre}: {estado['estado']} | "
                f"Ubicacion: {ubicacion}"
            )

        self.detalle.config(state="normal")
        self.detalle.delete("1.0", tk.END)
        self.detalle.insert("1.0", "\n".join(lineas))
        self.detalle.config(state="disabled")

    def dibujar(self):
        self.canvas.delete("all")

        tick_actual = self.ticks[self.indice]
        self.etiqueta_tick.config(text=f"Tick: {tick_actual}")

        # Obtenemos los estados hasta el tick seleccionado.
        estados, fallas, visibles = self.reconstruir_estados(
            tick_actual
        )

        self.dibujar_mapa()
        self.dibujar_fallas(fallas)
        self.dibujar_camiones(estados)

        self.mostrar_historial(
            tick_actual,
            visibles,
            estados
        )

    def ejecutar(self):
        self.ventana.mainloop()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Animacion de la mina del Certamen III"
    )

    parser.add_argument(
        "archivo",
        help="Ruta del CSV consolidado"
    )

    argumentos = parser.parse_args()

    eventos = LectorLogs().leer(argumentos.archivo)

    animacion = AnimacionMina(eventos)
    animacion.ejecutar()
