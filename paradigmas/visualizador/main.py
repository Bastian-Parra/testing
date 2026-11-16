import argparse
from lector_logs import LectorLogs
from simulacion import Simulacion


def main():
    parser = argparse.ArgumentParser(
        description="Visualizador de eventos de la mina"
    )
    parser.add_argument("archivo", help="Ruta del log consolidado")
    args = parser.parse_args()

    lector = LectorLogs()
    eventos = lector.leer(args.archivo)

    simulacion = Simulacion()
    simulacion.procesar(eventos)

    print(f"Eventos cargados: {len(eventos)}")
    print(f"Equipos encontrados: {len(simulacion.equipos)}")

    for identificador, equipo in simulacion.equipos.items():
        print(
            f"{identificador}: "
            f"estado={equipo.estado}, recurso={equipo.recurso}"
        )


if __name__ == "__main__":
    main()