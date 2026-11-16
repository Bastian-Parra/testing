from modelos import Camion, Pala, Stockpile


class Simulacion:
    def __init__(self):
        self.equipos = {}
        self.historial = []

    def obtener_equipo(self, identificador):
        if identificador not in self.equipos:
            if identificador.startswith("C-"):
                equipo = Camion(identificador)
            elif identificador.lower().startswith("pala"):
                equipo = Pala(identificador)
            elif identificador.lower().startswith("stockpile"):
                equipo = Stockpile(identificador)
            else:
                return None

            self.equipos[identificador] = equipo

        return self.equipos[identificador]

    def procesar(self, eventos):
        for evento in eventos:
            equipo = self.obtener_equipo(evento.entidad)

            if equipo is not None:
                equipo.estado = evento.tipo
                if evento.recurso:
                    equipo.recurso = evento.recurso

            self.historial.append(evento)