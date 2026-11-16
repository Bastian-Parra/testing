
import argparse
import csv
from pathlib import Path
import tkinter as tk
from tkinter import ttk


# Indicadores que queremos comparar entre escenarios.
INDICADORES = {
    "toneladas": "Toneladas transportadas",
    "utilizacion": "Utilizacion de palas (%)",
    "cola_palas": "Tiempo promedio en cola de palas",
    "cola_stockpiles": "Tiempo promedio en cola de stockpiles",
    "atrapados": "Camiones atrapados",
}

COLORES = [
    "#2563eb",
    "#0d9488",
    "#f59e0b",
    "#dc2626",
    "#8b5cf6",
]


def leer_metricas(ruta):
    # Cada archivo contiene los indicadores de un escenario.
    # No calculamos metricas que no esten disponibles.
    ruta = Path(ruta)

    if not ruta.is_file():
        raise FileNotFoundError(f"No existe: {ruta}")

    with ruta.open("r", encoding="utf-8-sig", newline="") as archivo:
        lector = csv.DictReader(archivo)

        columnas = set(lector.fieldnames or [])
        requeridas = {"Indicador", "Valor"}

        if not requeridas.issubset(columnas):
            raise ValueError(
                f"{ruta}: se necesitan las columnas Indicador,Valor"
            )

        resultado = {}

        for fila in lector:
            indicador = fila["Indicador"].strip()
            valor = fila["Valor"].strip()

            if indicador not in INDICADORES or not valor:
                continue

            resultado[indicador] = float(valor)

    return resultado


def cargar_escenarios(archivos):
    escenarios = {}

    for archivo in archivos:
        ruta = Path(archivo)
        nombre = ruta.stem

        if nombre in escenarios:
            raise ValueError(
                f"Hay dos escenarios con el mismo nombre: {nombre}"
            )

        escenarios[nombre] = leer_metricas(ruta)

    return escenarios


class ComparadorEscenarios:
    def __init__(self, escenarios):
        self.escenarios = escenarios

        self.ventana = tk.Tk()
        self.ventana.title(
            "Certamen III - Comparacion de escenarios"
        )
        self.ventana.geometry("1050x650")

        ttk.Label(
            self.ventana,
            text="Comparacion de escenarios de la mina",
            font=("Arial", 16, "bold")
        ).pack(pady=15)

        ttk.Label(
            self.ventana,
            text=(
                "Los graficos utilizan exclusivamente "
                "las metricas proporcionadas."
            )
        ).pack(pady=5)

        controles = ttk.Frame(self.ventana)
        controles.pack(pady=10)

        ttk.Label(
            controles,
            text="Indicador:"
        ).pack(side="left", padx=5)

        self.indicador = tk.StringVar(value="toneladas")

        self.selector = ttk.Combobox(
            controles,
            textvariable=self.indicador,
            values=list(INDICADORES),
            state="readonly",
            width=22
        )
        self.selector.pack(side="left", padx=5)
        self.selector.bind(
            "<<ComboboxSelected>>",
            lambda _: self.dibujar()
        )

        self.canvas = tk.Canvas(
            self.ventana,
            bg="white",
            width=950,
            height=420
        )
        self.canvas.pack(
            fill="both",
            expand=True,
            padx=20,
            pady=10
        )

        self.canvas.bind(
            "<Configure>",
            lambda _: self.dibujar()
        )

        self.dibujar()

    def dibujar(self):
        self.canvas.delete("all")

        indicador = self.indicador.get()

        datos = [
            (nombre, metricas[indicador])
            for nombre, metricas in self.escenarios.items()
            if indicador in metricas
        ]

        if not datos:
            self.canvas.create_text(
                450,
                180,
                text=(
                    "No hay datos disponibles para "
                    "este indicador."
                ),
                font=("Arial", 13)
            )
            return

        ancho_canvas = max(self.canvas.winfo_width(), 650)
        alto_canvas = max(self.canvas.winfo_height(), 350)

        margen_izquierdo = 180
        margen_derecho = 100
        margen_superior = 65

        ancho_maximo = max(
            100,
            ancho_canvas - margen_izquierdo - margen_derecho
        )

        valores = [valor for _, valor in datos]

        # Los indicadores previstos son cantidades no negativas.
        if any(valor < 0 for valor in valores):
            self.canvas.create_text(
                ancho_canvas / 2,
                180,
                text="No se admiten valores negativos."
            )
            return

        maximo = max(valores) if max(valores) > 0 else 1

        alto_fila = min(
            75,
            max(
                45,
                (alto_canvas - margen_superior - 40)
                / max(1, len(datos))
            )
        )

        self.canvas.create_text(
            ancho_canvas / 2,
            30,
            text=INDICADORES[indicador],
            font=("Arial", 14, "bold")
        )

        for indice, (nombre, valor) in enumerate(datos):
            y = margen_superior + indice * alto_fila

            longitud = (valor / maximo) * ancho_maximo

            self.canvas.create_text(
                margen_izquierdo - 15,
                y + 15,
                text=nombre,
                anchor="e",
                font=("Arial", 10)
            )

            self.canvas.create_rectangle(
                margen_izquierdo,
                y,
                margen_izquierdo + longitud,
                y + 30,
                fill=COLORES[indice % len(COLORES)],
                outline=""
            )

            self.canvas.create_text(
                margen_izquierdo + longitud + 10,
                y + 15,
                text=f"{valor:,.2f}",
                anchor="w",
                font=("Arial", 10, "bold")
            )

    def ejecutar(self):
        self.ventana.mainloop()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Comparar metricas de escenarios"
    )

    parser.add_argument(
        "archivos",
        nargs="+",
        help="Archivos CSV con columnas Indicador,Valor"
    )

    argumentos = parser.parse_args()

    try:
        escenarios = cargar_escenarios(argumentos.archivos)
    except (OSError, ValueError) as error:
        parser.error(str(error))

    ComparadorEscenarios(escenarios).ejecutar()
