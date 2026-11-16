import csv
from pathlib import Path
from modelos import Evento


class LectorLogs:
    COLUMNAS = [
        "Seq", "Tick", "Entidad",
        "Evento", "Recurso", "Toneladas"
    ]

    def leer(self, ruta):
        eventos = []
        secuencias = set()

        with Path(ruta).open(
            "r", encoding="utf-8-sig", newline=""
        ) as archivo:
            lector = csv.DictReader(archivo)

            if lector.fieldnames != self.COLUMNAS:
                raise ValueError("El CSV no tiene las columnas esperadas")

            for numero, fila in enumerate(lector, start=2):
                try:
                    seq = int(fila["Seq"])
                    tick = int(fila["Tick"])

                    if seq in secuencias:
                        raise ValueError(f"Seq duplicado: {seq}")

                    toneladas = (
                        float(fila["Toneladas"])
                        if fila["Toneladas"].strip()
                        else None
                    )

                    evento = Evento(
                        seq=seq,
                        tick=tick,
                        entidad=fila["Entidad"],
                        tipo=fila["Evento"],
                        recurso=fila["Recurso"],
                        toneladas=toneladas
                    )

                    eventos.append(evento)
                    secuencias.add(seq)

                except (TypeError, ValueError) as error:
                    raise ValueError(
                        f"Error en línea {numero}: {error}"
                    ) from error

        return sorted(eventos, key=lambda evento: evento.seq)