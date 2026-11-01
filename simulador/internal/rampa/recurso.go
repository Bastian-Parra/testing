package rampa

import (
	"context"
	"fmt"
	"time"
)

// RecursoDestino representa los lugares disponibles en un
// destino limitado, por ejemplo una pala.
//
// El canal funciona como un semáforo:
// su capacidad corresponde a la cantidad máxima de lugares.
type RecursoDestino struct {
	ID string

	ocupados chan struct{}
}

// NuevoRecursoDestino crea un recurso con una cantidad fija
// de lugares disponibles.
func NuevoRecursoDestino(
	id string,
	capacidad int,
) (*RecursoDestino, error) {

	if id == "" {
		return nil, fmt.Errorf(
			"el recurso debe tener identificador",
		)
	}

	if capacidad <= 0 {
		return nil, fmt.Errorf(
			"la capacidad debe ser mayor que cero",
		)
	}

	return &RecursoDestino{
		ID:       id,
		ocupados: make(chan struct{}, capacidad),
	}, nil
}

// Reservar intenta ocupar un lugar.
//
// El envío al canal bufferizado realiza de forma atómica
// la comprobación y la reserva.
//
// No hacemos:
//
//	if hayLugar {
//	    ocuparLugar()
//	}
//
// porque dos goroutines podrían observar simultáneamente
// el mismo lugar como disponible.
//
// En cambio, solamente una puede completar el envío cuando
// queda el último lugar.
func (r *RecursoDestino) Reservar(
	ctx context.Context,
	limite time.Duration,
) error {

	if limite <= 0 {
		return fmt.Errorf(
			"el limite de espera debe ser mayor que cero",
		)
	}

	temporizador := time.NewTimer(limite)
	defer temporizador.Stop()

	select {
	case r.ocupados <- struct{}{}:
		return nil

	case <-ctx.Done():
		return ctx.Err()

	case <-temporizador.C:
		return &ErrorEspera{
			Recurso: r.ID,
		}
	}
}

// Liberar devuelve un lugar al recurso.
func (r *RecursoDestino) Liberar() error {
	select {
	case <-r.ocupados:
		return nil

	default:
		return fmt.Errorf(
			"el recurso %s no tiene reservas que liberar",
			r.ID,
		)
	}
}

// Ocupacion devuelve la cantidad actual de lugares utilizados.
func (r *RecursoDestino) Ocupacion() int {
	return len(r.ocupados)
}

// Capacidad devuelve la cantidad máxima de lugares.
func (r *RecursoDestino) Capacidad() int {
	return cap(r.ocupados)
}
