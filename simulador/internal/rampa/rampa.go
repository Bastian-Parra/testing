package rampa

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/mina"
)

// Sentido representa el sentido en que un camión recorre una rampa.
type Sentido string

const (
	SentidoAB Sentido = "A->B"
	SentidoBA Sentido = "B->A"
)

// ErrEsperaProlongada permite reconocer cuando una adquisición
// superó el tiempo máximo permitido.
var ErrEsperaProlongada = errors.New("espera prolongada")

// ErrorEspera indica qué recurso estaba esperando el camión
// cuando venció su tiempo límite.
//
// Más adelante la issue de logs podrá utilizar esta información.
type ErrorEspera struct {
	Recurso string
}

func (e *ErrorEspera) Error() string {
	return fmt.Sprintf(
		"%v esperando recurso %s",
		ErrEsperaProlongada,
		e.Recurso,
	)
}

func (e *ErrorEspera) Unwrap() error {
	return ErrEsperaProlongada
}

// Rampa representa un tramo de una sola vía.
//
// No existe un límite de camiones dentro de ella, pero solamente
// puede existir un sentido activo al mismo tiempo.
type Rampa struct {
	ID    string
	NodoA string
	NodoB string

	mu sync.Mutex

	sentidoActual Sentido
	ocupantes     int

	// cambio se cierra cada vez que la rampa queda disponible.
	// Los camiones que están esperando pueden despertarse
	// sin tener que consultar constantemente el estado.
	cambio chan struct{}
}

// NuevaRampa crea la representación concurrente de un tramo rampa.
func NuevaRampa(tramo *mina.Tramo) (*Rampa, error) {
	if tramo == nil {
		return nil, fmt.Errorf("el tramo no puede ser nulo")
	}

	if tramo.Tipo != mina.TipoRampa {
		return nil, fmt.Errorf(
			"el tramo %s no es una rampa",
			tramo.ID,
		)
	}

	return &Rampa{
		ID:     tramo.ID,
		NodoA:  tramo.NodoA,
		NodoB:  tramo.NodoB,
		cambio: make(chan struct{}),
	}, nil
}

// sentidoValido comprueba los dos únicos sentidos posibles.
func sentidoValido(sentido Sentido) bool {
	return sentido == SentidoAB || sentido == SentidoBA
}

// puedeEntrar debe ejecutarse con el mutex adquirido.
//
// Un camión puede entrar cuando:
//
//  1. la rampa está vacía; o
//  2. ya existen camiones viajando en su mismo sentido.
func (r *Rampa) puedeEntrar(sentido Sentido) bool {
	return r.ocupantes == 0 || r.sentidoActual == sentido
}

// Adquirir espera hasta poder ingresar a la rampa.
//
// Si supera el límite indicado devuelve ErrEsperaProlongada.
//
// Importante: el timeout no libera recursos que el camión
// pudiera haber reservado antes. Esa responsabilidad pertenece
// al protocolo que coordina los recursos.
func (r *Rampa) Adquirir(
	ctx context.Context,
	sentido Sentido,
	limite time.Duration,
) error {

	if !sentidoValido(sentido) {
		return fmt.Errorf("sentido de rampa inválido: %q", sentido)
	}

	if limite <= 0 {
		return fmt.Errorf("el limite de espera debe ser mayor que cero")
	}

	temporizador := time.NewTimer(limite)
	defer temporizador.Stop()

	for {
		r.mu.Lock()

		if r.puedeEntrar(sentido) {
			if r.ocupantes == 0 {
				r.sentidoActual = sentido
			}

			r.ocupantes++

			r.mu.Unlock()
			return nil
		}

		// Guardamos el canal actual antes de liberar el mutex.
		//
		// Cuando el último camión salga, ese canal será cerrado
		// y despertará a quienes estén esperando.
		cambio := r.cambio

		r.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-temporizador.C:
			return &ErrorEspera{
				Recurso: r.ID,
			}

		case <-cambio:
			// El estado cambió. Volvemos al inicio y revisamos
			// nuevamente si nuestro sentido puede entrar.
		}
	}
}

// Liberar indica que un camión terminó de utilizar la rampa.
func (r *Rampa) Liberar(sentido Sentido) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.ocupantes == 0 {
		return fmt.Errorf(
			"la rampa %s ya se encuentra vacía",
			r.ID,
		)
	}

	if r.sentidoActual != sentido {
		return fmt.Errorf(
			"no se puede liberar %s en sentido %s; el sentido activo es %s",
			r.ID,
			sentido,
			r.sentidoActual,
		)
	}

	r.ocupantes--

	// Mientras quede al menos un camión, el sentido sigue reservado.
	if r.ocupantes > 0 {
		return nil
	}

	// Al salir el último camión la rampa vuelve a quedar libre.
	r.sentidoActual = ""

	// Despertamos a todos los que estaban esperando.
	close(r.cambio)

	// Creamos un nuevo canal para el siguiente cambio de estado.
	r.cambio = make(chan struct{})

	return nil
}

// Estado devuelve una fotografía coherente de la ocupación.
//
// Será útil posteriormente para la visualización.
func (r *Rampa) Estado() (Sentido, int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.sentidoActual, r.ocupantes
}
