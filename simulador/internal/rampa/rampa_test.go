package rampa

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/mina"
)

func nuevaRampaPrueba(t *testing.T) *Rampa {
	t.Helper()

	tramo := &mina.Tramo{
		ID:     "R1",
		NodoA:  "I1",
		NodoB:  "P1",
		Tipo:   mina.TipoRampa,
		Tiempo: 4,
	}

	r, err := NuevaRampa(tramo)
	if err != nil {
		t.Fatal(err)
	}

	return r
}

// Varios camiones pueden utilizar simultáneamente
// la rampa si viajan en el mismo sentido.
func TestRampaPermiteMismoSentido(t *testing.T) {
	r := nuevaRampaPrueba(t)
	ctx := context.Background()

	if err := r.Adquirir(
		ctx,
		SentidoAB,
		100*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}

	if err := r.Adquirir(
		ctx,
		SentidoAB,
		100*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}

	sentido, ocupantes := r.Estado()

	if sentido != SentidoAB {
		t.Fatalf(
			"sentido esperado %s, obtenido %s",
			SentidoAB,
			sentido,
		)
	}

	if ocupantes != 2 {
		t.Fatalf(
			"se esperaban 2 ocupantes, se obtuvieron %d",
			ocupantes,
		)
	}
}

// El sentido contrario debe esperar mientras existan
// camiones dentro de la rampa.
func TestRampaBloqueaSentidoContrario(t *testing.T) {
	r := nuevaRampaPrueba(t)
	ctx := context.Background()

	if err := r.Adquirir(
		ctx,
		SentidoAB,
		100*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}

	err := r.Adquirir(
		ctx,
		SentidoBA,
		20*time.Millisecond,
	)

	if !errors.Is(err, ErrEsperaProlongada) {
		t.Fatalf(
			"se esperaba espera prolongada, se obtuvo %v",
			err,
		)
	}

	if err := r.Liberar(SentidoAB); err != nil {
		t.Fatal(err)
	}

	if err := r.Adquirir(
		ctx,
		SentidoBA,
		100*time.Millisecond,
	); err != nil {
		t.Fatalf(
			"el sentido contrario debería poder entrar: %v",
			err,
		)
	}
}

// En el protocolo ingenuo el camión entra a la rampa
// aunque el destino esté lleno.
func TestIngenuoPuedeQuedarEsperandoConRampaTomada(
	t *testing.T,
) {
	r := nuevaRampaPrueba(t)

	destino, err := NuevoRecursoDestino("P1", 1)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Llenamos la pala antes de que llegue nuestro camión.
	if err := destino.Reservar(
		ctx,
		100*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}

	paso, err := Entrar(
		ctx,
		r,
		SentidoAB,
		destino,
		ProtocoloIngenuo,
		100*time.Millisecond,
	)

	if err != nil {
		t.Fatal(err)
	}

	// Puede entrar a la rampa aunque P1 esté llena.
	_, ocupantes := r.Estado()

	if ocupantes != 1 {
		t.Fatal(
			"el camion debería encontrarse dentro de la rampa",
		)
	}

	// Al llegar al final intenta conseguir un cupo.
	err = paso.LlegarDestino(
		ctx,
		20*time.Millisecond,
	)

	if !errors.Is(err, ErrEsperaProlongada) {
		t.Fatalf(
			"se esperaba espera prolongada, se obtuvo %v",
			err,
		)
	}

	// Como es ingenuo, continúa reteniendo la rampa.
	_, ocupantes = r.Estado()

	if ocupantes != 1 {
		t.Fatal(
			"el protocolo ingenuo debería seguir reteniendo la rampa",
		)
	}

	paso.Cancelar()

	// Liberamos también el cupo que ocupaba el otro camión.
	if err := destino.Liberar(); err != nil {
		t.Fatal(err)
	}
}

// Prevención no permite ingresar a la rampa sin reservar
// antes el lugar del destino.
func TestPrevencionNoEntraSiDestinoLleno(t *testing.T) {
	r := nuevaRampaPrueba(t)

	destino, err := NuevoRecursoDestino("P1", 1)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	if err := destino.Reservar(
		ctx,
		100*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}

	_, err = Entrar(
		ctx,
		r,
		SentidoAB,
		destino,
		ProtocoloPrevencion,
		20*time.Millisecond,
	)

	if !errors.Is(err, ErrEsperaProlongada) {
		t.Fatalf(
			"se esperaba espera prolongada, se obtuvo %v",
			err,
		)
	}

	// La diferencia fundamental:
	// el camión nunca entró a R1.
	_, ocupantes := r.Estado()

	if ocupantes != 0 {
		t.Fatal(
			"prevención no debería ocupar la rampa sin reserva de destino",
		)
	}

	if err := destino.Liberar(); err != nil {
		t.Fatal(err)
	}
}

// Comprueba el flujo completo del protocolo de prevención.
func TestPrevencionReservaAntesDeEntrar(t *testing.T) {
	r := nuevaRampaPrueba(t)

	destino, err := NuevoRecursoDestino("P1", 1)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	paso, err := Entrar(
		ctx,
		r,
		SentidoAB,
		destino,
		ProtocoloPrevencion,
		100*time.Millisecond,
	)

	if err != nil {
		t.Fatal(err)
	}

	if destino.Ocupacion() != 1 {
		t.Fatal(
			"el lugar debería estar reservado antes de recorrer la rampa",
		)
	}

	if err := paso.LlegarDestino(
		ctx,
		100*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}

	_, ocupantes := r.Estado()

	if ocupantes != 0 {
		t.Fatal(
			"la rampa debería liberarse al llegar al destino",
		)
	}

	if destino.Ocupacion() != 1 {
		t.Fatal(
			"el camión debe conservar su lugar en la pala",
		)
	}

	if err := paso.LiberarDestino(); err != nil {
		t.Fatal(err)
	}

	if destino.Ocupacion() != 0 {
		t.Fatal(
			"el lugar debería quedar libre",
		)
	}
}

// Si prevención reserva un destino pero no consigue la rampa,
// debe devolver la reserva.
func TestPrevencionDevuelveReservaSiFallaRampa(t *testing.T) {
	r := nuevaRampaPrueba(t)

	destino, err := NuevoRecursoDestino("P1", 1)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Bloqueamos R1 en el sentido contrario.
	if err := r.Adquirir(
		ctx,
		SentidoBA,
		100*time.Millisecond,
	); err != nil {
		t.Fatal(err)
	}

	_, err = Entrar(
		ctx,
		r,
		SentidoAB,
		destino,
		ProtocoloPrevencion,
		20*time.Millisecond,
	)

	if !errors.Is(err, ErrEsperaProlongada) {
		t.Fatalf(
			"se esperaba espera prolongada, se obtuvo %v",
			err,
		)
	}

	// La reserva realizada antes de intentar tomar la rampa
	// debe haberse devuelto.
	if destino.Ocupacion() != 0 {
		t.Fatal(
			"la reserva del destino quedó retenida indebidamente",
		)
	}

	if err := r.Liberar(SentidoBA); err != nil {
		t.Fatal(err)
	}
}

// Comprueba que la reserva de un destino sea atómica.
//
// Dos goroutines intentan reservar el único lugar disponible.
// Solamente una debe conseguirlo.
func TestReservaDestinoEsAtomica(t *testing.T) {
	destino, err := NuevoRecursoDestino("P1", 1)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	resultados := make(chan error, 2)

	for i := 0; i < 2; i++ {
		go func() {
			resultados <- destino.Reservar(
				ctx,
				30*time.Millisecond,
			)
		}()
	}

	exitos := 0
	esperas := 0

	for i := 0; i < 2; i++ {
		err := <-resultados

		switch {
		case err == nil:
			exitos++

		case errors.Is(err, ErrEsperaProlongada):
			esperas++

		default:
			t.Fatalf(
				"se obtuvo un error inesperado: %v",
				err,
			)
		}
	}

	if exitos != 1 {
		t.Fatalf(
			"solo una goroutine debía reservar el lugar; éxitos: %d",
			exitos,
		)
	}

	if esperas != 1 {
		t.Fatalf(
			"una goroutine debía quedar esperando; esperas: %d",
			esperas,
		)
	}

	if destino.Ocupacion() != 1 {
		t.Fatalf(
			"la ocupación esperada era 1, se obtuvo %d",
			destino.Ocupacion(),
		)
	}

	if err := destino.Liberar(); err != nil {
		t.Fatal(err)
	}
}
