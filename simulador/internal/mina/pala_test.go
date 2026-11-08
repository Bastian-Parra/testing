package mina

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/camion"
)

func TestPalaCamionWaitGroup(t *testing.T) {
	configCamion := camion.Configuracion{
		Tick: 10 * time.Millisecond,
		TicksViajeVacio: 1,
		TicksCarga: 2,
		TicksTransporte: 2,
		TicksDescarga: 2,
	}

	cam, _ :=	camion.NuevoCamion("C1", configCamion)
	pala, _ := NuevaPala("P1", 20*time.Millisecond, 5)

	ctx, cancelar := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	// se agrega el camión al WaitGroup y se ejecuta en una goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		cam.Ejecutar(ctx)
	}()


	//	se agrega la pala al WaitGroup y se ejecuta en una goroutine
	wg.Add(1)
	go pala.Ejecutar(ctx, &wg)

	// se envia el camión a la pala para que comience la carga
	time.Sleep(20* time.Millisecond)
	pala.Cola <- cam

	exito := false
	for i := 0; i < 50; i++ {
		estadoActual := cam.Estado()
		if estadoActual == camion.EstadoCargando || estadoActual == camion.EstadoTransportando {
			exito = true
			t.Logf("el camión ha pasado a estado %s", estadoActual)
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !exito {
		t.Fatalf("error: el camión no ha cambiado de estado")
	}

	cancelar() // se cancela el contexto para detener la ejecución de las goroutines
	wg.Wait() // se espera a que todas las goroutines finalicen
}