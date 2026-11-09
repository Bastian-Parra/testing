package despachador

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/mina"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/ruteo"
)

// Solicitud contiene la información que envía un camión.
type Solicitud struct {
	CamionID  string
	Origen    string
	Cargado   bool
	Fallas    map[string]bool
	Respuesta chan Resultado
}

// Resultado indica el destino asignado y la ruta.
type Resultado struct {
	CamionID string
	Destino  string
	Ruta     []string
	Err      error
}

// Despachador guarda la mina y el turno de asignación.
type Despachador struct {
	mina           *mina.Mina
	planificador   *ruteo.Planificador
	turnoPalas     int
	turnoStockpile int
}

// NuevoDespachador prepara el componente.
func NuevoDespachador(m *mina.Mina) *Despachador {
	return &Despachador{
		mina:         m,
		planificador: ruteo.NuevoPlanificador(m),
	}
}

// destinos obtiene los nodos del tipo solicitado.
// Se ordenan para que Round Robin siempre tenga
// el mismo orden de referencia.
func (d *Despachador) destinos(tipo mina.TipoNodo) []string {
	var ids []string

	for id, nodo := range d.mina.Nodos {
		if nodo != nil && nodo.Tipo == tipo {
			ids = append(ids, id)
		}
	}

	sort.Strings(ids)
	return ids
}

// asignar busca el siguiente destino alcanzable.
func (d *Despachador) asignar(s Solicitud) Resultado {
	resultado := Resultado{CamionID: s.CamionID}

	if d.mina == nil || d.planificador == nil {
		resultado.Err = errors.New("despachador sin mina")
		return resultado
	}

	if _, existe := d.mina.Nodos[s.Origen]; !existe {
		resultado.Err = fmt.Errorf("origen %s no existe", s.Origen)
		return resultado
	}

	// Un camión cargado va al stockpile.
	// Uno vacío va a una pala.
	tipo := mina.TipoPala
	turno := &d.turnoPalas

	if s.Cargado {
		tipo = mina.TipoStockpile
		turno = &d.turnoStockpile
	}

	candidatos := d.destinos(tipo)

	if len(candidatos) == 0 {
		resultado.Err = errors.New("no existen destinos del tipo solicitado")
		return resultado
	}

	// Revisamos los destinos en orden circular.
	for i := 0; i < len(candidatos); i++ {
		indice := (*turno + i) % len(candidatos)
		destino := candidatos[indice]

		ruta, err := d.planificador.CalcularRuta(
			s.Origen, destino, s.Cargado, s.Fallas,
		)

		if err != nil || len(ruta) == 0 {
			continue
		}

		// Avanzamos el turno solo cuando asignamos.
		*turno = (indice + 1) % len(candidatos)

		resultado.Destino = destino
		resultado.Ruta = ruta
		return resultado
	}

	resultado.Err = errors.New("no hay destinos alcanzables")
	return resultado
}

// Ejecutar recibe solicitudes hasta que se cierre
// el canal o se cancele el contexto.
// Esta función se debe iniciar como goroutine.
func (d *Despachador) Ejecutar(
	ctx context.Context,
	solicitudes <-chan Solicitud,
) {
	for {
		select {
		case <-ctx.Done():
			return

		case solicitud, abierto := <-solicitudes:
			if !abierto {
				return
			}

			resultado := d.asignar(solicitud)

			// Si se cancela el simulador, no quedamos
			// esperando indefinidamente una respuesta.
			if solicitud.Respuesta != nil {
				select {
				case solicitud.Respuesta <- resultado:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}
