package fallas

import (
	"context"
	"fmt"
	"time"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/mina"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/rampa"
)

// Recorrido permite que un camion avance por los tramos
// de una ruta, respetando las fallas y las rampas.
type Recorrido struct {
	Mina        *mina.Mina
	Controlador *Controlador
	Rampas      map[string]*rampa.Rampa

	Tick      time.Duration
	Protocolo rampa.Protocolo
	Limite    time.Duration
	// Evento permite notificar el estado al camion sin cambiar su ruta.
	Evento func(camionID, evento, tramoID, sentido string, avance int)
	// PanicEnTick sirve para probar una falla inesperada de forma controlada.
	PanicEnTick func(camionID, tramoID string, avance int) bool
}

// Recorrer ejecuta una ruta que ya fue asignada.
// No cambia la ruta si aparece una falla.
func (r *Recorrido) Recorrer(
	ctx context.Context,
	camionID string,
	ruta []string,
) error {
	if r == nil || r.Mina == nil || r.Controlador == nil {
		return fmt.Errorf("recorrido sin componentes")
	}
	if r.Tick <= 0 || r.Limite <= 0 {
		return fmt.Errorf("tiempos de recorrido invalidos")
	}
	if camionID == "" || len(ruta) < 2 {
		return fmt.Errorf("camion o ruta invalidos")
	}

	for i := 0; i < len(ruta)-1; i++ {
		origen := ruta[i]
		destino := ruta[i+1]

		tramo, sentido, err := r.buscarTramo(origen, destino)
		if err != nil {
			return err
		}

		switch tramo.Tipo {
		case mina.TipoDoble:
			err = r.recorrerDoble(ctx, camionID, tramo, sentido)

		case mina.TipoRampa:
			err = r.recorrerRampa(ctx, tramo, sentido)

		default:
			err = fmt.Errorf("tipo de tramo desconocido: %s", tramo.Tipo)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// buscarTramo encuentra el tramo que conecta dos nodos.
// El sentido se representa usando los identificadores reales.
func (r *Recorrido) buscarTramo(
	origen, destino string,
) (*mina.Tramo, string, error) {
	if r.Mina.Nodos[origen] == nil || r.Mina.Nodos[destino] == nil {
		return nil, "", fmt.Errorf("nodo no encontrado")
	}

	for _, tramo := range r.Mina.Tramos {
		if tramo == nil {
			continue
		}

		if tramo.NodoA == origen && tramo.NodoB == destino {
			return tramo, origen + "->" + destino, nil
		}

		if tramo.NodoB == origen && tramo.NodoA == destino {
			return tramo, origen + "->" + destino, nil
		}
	}

	return nil, "", fmt.Errorf(
		"no existe tramo entre %s y %s", origen, destino,
	)
}

// esperarTick permite avanzar un tick sin ignorar
// la cancelacion del simulador.
func (r *Recorrido) esperarTick(ctx context.Context) error {
	timer := time.NewTimer(r.Tick)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// recorrerDoble controla el avance de un camion
// dentro de un tramo doble.
func (r *Recorrido) recorrerDoble(
	ctx context.Context,
	camionID string,
	tramo *mina.Tramo,
	sentido string,
) (retErr error) {
	if tramo.Tiempo <= 0 {
		return fmt.Errorf("tiempo invalido en %s", tramo.ID)
	}

	// El camion entra a la cola del sentido correspondiente.
	if err := r.Controlador.IniciarTramo(
		tramo, sentido, camionID,
	); err != nil {
		return err
	}

	// Al cancelar, los camiones sanos salen del registro; los
	// permanentemente averiados conservan su posicion.
	defer r.Controlador.Circulacion.RetirarAlTerminar(tramo.ID, sentido, camionID)
	// recover debe ejecutarse en la misma goroutine que circula.
	// Se registra antes de retirar el camion para conservar el obstaculo.
	defer func() {
		if valor := recover(); valor != nil {
			retErr = fmt.Errorf("panic recuperado en camion %s: %v", camionID, valor)
			if err := r.Controlador.RegistrarPanic(tramo.ID, sentido, camionID); err != nil {
				retErr = fmt.Errorf("%w; registro: %v", retErr, err)
			}
			if r.Evento != nil {
				r.Evento(camionID, "panic_permanente", tramo.ID, sentido, 0)
			}
		}
	}()
	if r.Evento != nil {
		r.Evento(camionID, "entra", tramo.ID, sentido, 0)
	}

	// Si el turno termina, no eliminamos una falla
	// permanente mientras se ejecuta la simulacion.
	for avance := 0; avance < tramo.Tiempo; {
		if err := r.esperarTick(ctx); err != nil {
			return err
		}

		actual, err := r.Controlador.Circulacion.ConsultarCamion(
			tramo.ID, sentido, camionID,
		)
		if err != nil {
			return err
		}

		if r.PanicEnTick != nil && r.PanicEnTick(camionID, tramo.ID, avance) {
			panic("panic controlado en camion " + camionID)
		}
		// Una falla permanente deja detenido al camion.
		if actual.EnFalla && actual.FallaPermanente {
			<-ctx.Done()
			return ctx.Err()
		}

		// Una falla temporal necesita completar su reparacion.
		if actual.EnFalla {
			if r.Evento != nil {
				r.Evento(camionID, "reparando", tramo.ID, sentido, avance)
			}
			_, err := r.Controlador.RepararTick(
				tramo.ID, sentido, camionID,
			)
			if err != nil {
				return err
			}
			continue
		}

		// Un camion no puede superar la posicion del que va delante.
		if !r.Controlador.Circulacion.PuedeAvanzar(tramo.ID, sentido, camionID) {
			continue
		}
		// Solo evaluamos nuevas fallas cuando el camion
		// esta funcionando normalmente.
		resultado, err := r.Controlador.EvaluarCamion(
			tramo, sentido, camionID,
		)
		if err != nil {
			return err
		}

		if resultado.Tipo != SinFalla {
			if r.Evento != nil {
				r.Evento(camionID, string(resultado.Tipo), tramo.ID, sentido, avance)
			}
			continue
		}
		if err := r.Controlador.Circulacion.AvanzarTick(tramo.ID, sentido, camionID); err != nil {
			return err
		}
		avance++
		if r.Evento != nil {
			r.Evento(camionID, "avanza", tramo.ID, sentido, avance)
		}
	}

	// Aunque termine sus ticks, no puede adelantar
	// a otro camion que este delante.
	for !r.Controlador.Circulacion.PuedeSalir(
		tramo.ID, sentido, camionID,
	) {
		if err := r.esperarTick(ctx); err != nil {
			return err
		}
	}

	if r.Evento != nil {
		r.Evento(camionID, "sale", tramo.ID, sentido, tramo.Tiempo)
	}
	return r.Controlador.FinalizarTramo(
		tramo.ID, sentido, camionID,
	)
}

// recorrerRampa reutiliza el protocolo que ya existe.
// En las rampas no se generan fallas aleatorias.
func (r *Recorrido) recorrerRampa(
	ctx context.Context,
	tramo *mina.Tramo,
	sentido string,
) error {
	if tramo.Tiempo <= 0 {
		return fmt.Errorf("tiempo invalido en %s", tramo.ID)
	}

	recurso := r.Rampas[tramo.ID]
	if recurso == nil {
		return fmt.Errorf("rampa %s no registrada", tramo.ID)
	}

	direccion := rampa.SentidoAB
	if sentido == tramo.NodoB+"->"+tramo.NodoA {
		direccion = rampa.SentidoBA
	}

	paso, err := rampa.Entrar(
		ctx, recurso, direccion, nil,
		r.Protocolo, r.Limite,
	)
	if err != nil {
		return err
	}

	// Cancelar libera la rampa si el recorrido termina
	// antes de tiempo.
	defer paso.Cancelar()

	for i := 0; i < tramo.Tiempo; i++ {
		if err := r.esperarTick(ctx); err != nil {
			return err
		}
	}

	return paso.LlegarDestino(ctx, r.Limite)
}
