package camion

import (
	"context"
	"testing"
	"time"
)

// esperarEstado espera hasta que el camión alcance el estado
// solicitado o hasta que venza el tiempo máximo.
//
// Evita utilizar Sleep con tiempos exactos en las pruebas.
func esperarEstado(
	t *testing.T,
	c *Camion,
	esperado Estado,
	maximo time.Duration,
) {

	t.Helper()

	limite := time.Now().Add(maximo)

	for time.Now().Before(limite) {

		if c.Estado() == esperado {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatalf(
		"se esperaba estado %s, estado actual: %s",
		esperado,
		c.Estado(),
	)
}

// configuración rápida para las pruebas.
func configuracionPrueba() Configuracion {
	return Configuracion{
		Tick:            10 * time.Millisecond,
		TicksViajeVacio: 2,
		TicksCarga:      3,
		TicksTransporte: 3,
		TicksDescarga:   3,
	}
}

// Comprueba el estado inicial del camión.
func TestCamionComienzaViajandoVacio(t *testing.T) {

	c, err := NuevoCamion(
		"C01",
		configuracionPrueba(),
	)

	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if c.Estado() != EstadoViajandoVacio {
		t.Fatalf(
			"estado inicial incorrecto: %s",
			c.Estado(),
		)
	}
}

// Comprueba un ciclo básico completo.
func TestCamionCompletaCiclo(t *testing.T) {

	c, err := NuevoCamion(
		"C01",
		configuracionPrueba(),
	)

	if err != nil {
		t.Fatal(err)
	}

	ctx, cancelar := context.WithCancel(context.Background())
	defer cancelar()

	go c.Ejecutar(ctx)

	// Viaja vacío hasta llegar a la pala.
	esperarEstado(
		t,
		c,
		EstadoEnColaPala,
		200*time.Millisecond,
	)

	// La pala autoriza comenzar la carga.
	if err := c.AutorizarCarga(ctx); err != nil {
		t.Fatal(err)
	}

	esperarEstado(
		t,
		c,
		EstadoCargando,
		100*time.Millisecond,
	)

	// Al terminar la carga comienza automáticamente
	// el transporte hacia el stockpile.
	esperarEstado(
		t,
		c,
		EstadoTransportando,
		200*time.Millisecond,
	)

	esperarEstado(
		t,
		c,
		EstadoEnColaStockpile,
		200*time.Millisecond,
	)

	// El stockpile autoriza la descarga.
	if err := c.AutorizarDescarga(ctx); err != nil {
		t.Fatal(err)
	}

	esperarEstado(
		t,
		c,
		EstadoDescargando,
		100*time.Millisecond,
	)

	// Después de descargar comienza un nuevo ciclo.
	esperarEstado(
		t,
		c,
		EstadoViajandoVacio,
		200*time.Millisecond,
	)
}

// Esta prueba comprueba específicamente el comportamiento
// de los canales utilizados como comandos con guarda.
//
// Intentamos autorizar una carga mientras el camión todavía
// está viajando. El envío no debería completarse hasta que
// llegue al estado EnColaPala.
func TestCanalCargaFuncionaComoGuarda(t *testing.T) {

	config := configuracionPrueba()

	// Hacemos que el viaje dure más tiempo para poder
	// comprobar claramente que la autorización queda bloqueada.
	config.TicksViajeVacio = 5

	c, err := NuevoCamion("C01", config)

	if err != nil {
		t.Fatal(err)
	}

	ctx, cancelar := context.WithCancel(context.Background())
	defer cancelar()

	go c.Ejecutar(ctx)

	resultado := make(chan error, 1)

	go func() {
		resultado <- c.AutorizarCarga(ctx)
	}()

	// Durante los primeros milisegundos el camión sigue
	// viajando, así que el case de carga debe estar deshabilitado.
	select {

	case <-resultado:
		t.Fatal(
			"la carga fue autorizada antes de llegar a la cola de la pala",
		)

	case <-time.After(15 * time.Millisecond):
		// Correcto: la operación sigue esperando.
	}

	// Cuando llegue a EnColaPala, el canal deja de ser nil
	// y el envío puede completarse.
	select {

	case err := <-resultado:
		if err != nil {
			t.Fatalf(
				"no se esperaba error: %v",
				err,
			)
		}

	case <-time.After(200 * time.Millisecond):
		t.Fatal(
			"la autorización de carga nunca fue recibida",
		)
	}
}

// Comprueba que la goroutine pueda finalizar ordenadamente.
func TestCamionFinalizaConContexto(t *testing.T) {

	c, err := NuevoCamion(
		"C01",
		configuracionPrueba(),
	)

	if err != nil {
		t.Fatal(err)
	}

	ctx, cancelar := context.WithCancel(context.Background())

	go c.Ejecutar(ctx)

	cancelar()

	select {

	case <-c.Finalizado():
		// Correcto.

	case <-time.After(200 * time.Millisecond):
		t.Fatal(
			"la goroutine del camion no termino",
		)
	}
}
