package turno

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/camion"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/despachador"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/fallas"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/mina"
)

func TestCamionOriginalCargaYDescargaConRutas(t *testing.T) {
	if err := os.MkdirAll("../paradigmas/logs/sin_fallas", 0755); err != nil {
		t.Fatal(err)
	}
	tick := time.Millisecond
	m := &mina.Mina{
		Nodos: map[string]*mina.Nodo{
			"A": {ID: "A", Tipo: mina.TipoEstacionamiento},
			"B": {ID: "B", Tipo: mina.TipoPala},
			"D": {ID: "D", Tipo: mina.TipoStockpile},
		},
		Tramos: map[string]*mina.Tramo{
			"T1": {ID: "T1", NodoA: "A", NodoB: "B", Tipo: mina.TipoDoble, Tiempo: 2},
			"T2": {ID: "T2", NodoA: "B", NodoB: "D", Tipo: mina.TipoDoble, Tiempo: 2},
			"T3": {ID: "T3", NodoA: "D", NodoB: "A", Tipo: mina.TipoDoble, Tiempo: 2},
		},
	}
	gestor, err := fallas.NuevoGestor(fallas.Configuracion{ProbabilidadFalla: 0, ProbabilidadPermanente: 0, ReparacionMinTicks: 3, ReparacionMaxTicks: 8, Tick: tick, Semilla: 1})
	if err != nil {
		t.Fatal(err)
	}
	ctrl, err := fallas.NuevoControlador(gestor, fallas.NuevaCirculacion())
	if err != nil {
		t.Fatal(err)
	}
	rec := &fallas.Recorrido{Mina: m, Controlador: ctrl, Tick: tick, Limite: 100 * tick}
	pala, err := mina.NuevaPala("Pala-Test-21", 3*tick, 2)
	if err != nil {
		t.Fatal(err)
	}
	stock, err := mina.NuevaStockpile("Stock-Test-21", 3*tick, 2)
	if err != nil {
		t.Fatal(err)
	}
	c, err := camion.NuevoCamion("C-Test-21", camion.Configuracion{Tick: tick, TicksViajeVacio: 2, TicksCarga: 3, TicksTransporte: 2, TicksDescarga: 3})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(2)
	go pala.Ejecutar(ctx, &wg)
	go stock.Ejecutar(ctx, &wg)
	solicitudes := make(chan despachador.Solicitud)
	wg.Add(1)
	go func() { defer wg.Done(); despachador.NuevoDespachador(m).Ejecutar(ctx, solicitudes) }()
	eventos := make(chan string, 50)
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = EjecutarCamionOriginal(ctx, c, rec, gestor, solicitudes, pala, stock, tick, 1, false,
			func(_, evento, _, _ string, _ int) {
				select {
				case eventos <- evento:
				default:
				}
			})
	}()
	<-ctx.Done()
	wg.Wait()
	close(eventos)
	carga, descarga := false, false
	for evento := range eventos {
		if evento == "carga" {
			carga = true
		}
		if evento == "descarga" {
			descarga = true
		}
	}
	if !carga || !descarga {
		t.Fatalf("no se usaron ambos recursos originales: carga=%v descarga=%v", carga, descarga)
	}
	select {
	case <-c.Finalizado():
	default:
		t.Fatal("camion original no finalizo")
	}
}
