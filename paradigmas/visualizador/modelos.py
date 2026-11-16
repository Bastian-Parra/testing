from dataclasses import dataclass
from typing import Optional


@dataclass
class Evento:
    seq: int
    tick: int
    entidad: str
    tipo: str
    recurso: str
    toneladas: Optional[float] = None


class Equipo:
    def __init__(self, identificador):
        self.identificador = identificador
        self.estado = "desconocido"
        self.recurso = None


class Camion(Equipo):
    pass


class Pala(Equipo):
    pass


class Stockpile(Equipo):
    pass