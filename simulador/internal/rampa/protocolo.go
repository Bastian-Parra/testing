package rampa

import (
	"context"
	"fmt"
	"time"
)

// Protocolo indica la estrategia utilizada para entrar a una rampa.
type Protocolo string

const (
	ProtocoloIngenuo    Protocolo = "ingenuo"
	ProtocoloPrevencion Protocolo = "prevencion"
)

// Paso representa los recursos que un camión retiene mientras
// atraviesa una rampa.
//
// En prevención puede contener desde el principio una reserva
// del destino.
//
// En ingenuo esa reserva se solicita al llegar al final.
type Paso struct {
	rampa   *Rampa
	sentido Sentido
	destino *RecursoDestino

	protocolo Protocolo

	rampaAdquirida   bool
	destinoReservado bool
}

// Entrar inicia el uso de una rampa según el protocolo seleccionado.
func Entrar(
	ctx context.Context,
	r *Rampa,
	sentido Sentido,
	destino *RecursoDestino,
	protocolo Protocolo,
	limite time.Duration,
) (*Paso, error) {

	if r == nil {
		return nil, fmt.Errorf("la rampa no puede ser nula")
	}

	paso := &Paso{
		rampa:     r,
		sentido:   sentido,
		destino:   destino,
		protocolo: protocolo,
	}

	switch protocolo {

	case ProtocoloIngenuo:
		// En el protocolo ingenuo solamente adquirimos la rampa.
		//
		// La reserva del destino ocurrirá al terminar el recorrido.
		if err := r.Adquirir(ctx, sentido, limite); err != nil {
			return nil, err
		}

		paso.rampaAdquirida = true
		return paso, nil

	case ProtocoloPrevencion:
		// Si el destino posee capacidad limitada, reservamos
		// primero uno de sus lugares.
		//
		// El envío al canal de RecursoDestino es atómico.
		if destino != nil {
			if err := destino.Reservar(ctx, limite); err != nil {
				return nil, err
			}

			paso.destinoReservado = true
		}

		// Solamente después de tener garantizado el destino
		// intentamos adquirir la rampa.
		if err := r.Adquirir(ctx, sentido, limite); err != nil {

			// Si no conseguimos la rampa, debemos devolver
			// la reserva para no dejar un lugar tomado.
			if paso.destinoReservado {
				_ = destino.Liberar()
				paso.destinoReservado = false
			}

			return nil, err
		}

		paso.rampaAdquirida = true
		return paso, nil

	default:
		return nil, fmt.Errorf(
			"protocolo desconocido: %q",
			protocolo,
		)
	}
}

// LlegarDestino representa el momento en que el camión termina
// de recorrer la rampa.
//
// IMPORTANTE:
//
// En el protocolo ingenuo el camión intenta reservar el destino
// mientras todavía retiene la rampa.
//
// Si no obtiene el lugar, la función devuelve el error de espera
// pero NO libera la rampa. Así el camión puede registrar la espera
// prolongada y volver a intentarlo.
//
// Esta retención es precisamente la situación que puede generar
// deadlock en el protocolo ingenuo.
func (p *Paso) LlegarDestino(
	ctx context.Context,
	limite time.Duration,
) error {

	if !p.rampaAdquirida {
		return fmt.Errorf(
			"el camion no mantiene la rampa",
		)
	}

	switch p.protocolo {

	case ProtocoloIngenuo:
		if p.destino != nil && !p.destinoReservado {

			if err := p.destino.Reservar(
				ctx,
				limite,
			); err != nil {

				// No liberamos la rampa.
				return err
			}

			p.destinoReservado = true
		}

	case ProtocoloPrevencion:
		// El destino ya fue reservado antes de entrar.
	}

	if err := p.rampa.Liberar(p.sentido); err != nil {
		return err
	}

	p.rampaAdquirida = false

	return nil
}

// LiberarDestino se utiliza cuando el camión deja posteriormente
// el lugar que había reservado en el destino.
func (p *Paso) LiberarDestino() error {
	if !p.destinoReservado || p.destino == nil {
		return fmt.Errorf(
			"el camion no mantiene una reserva de destino",
		)
	}

	if err := p.destino.Liberar(); err != nil {
		return err
	}

	p.destinoReservado = false

	return nil
}

// Cancelar libera los recursos que todavía retenga el paso.
//
// Esto será útil posteriormente al terminar el turno o al
// recuperar un camión desde un panic.
func (p *Paso) Cancelar() {
	if p.rampaAdquirida {
		_ = p.rampa.Liberar(p.sentido)
		p.rampaAdquirida = false
	}

	if p.destinoReservado && p.destino != nil {
		_ = p.destino.Liberar()
		p.destinoReservado = false
	}
}
