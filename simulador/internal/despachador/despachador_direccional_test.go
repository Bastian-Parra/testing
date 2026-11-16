package despachador

import (
	"context"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/mina"
	"testing"
)

func TestDespachadorEvitaSentidoBloqueado(t *testing.T) {
	m := &mina.Mina{Nodos: map[string]*mina.Nodo{
		"A": {ID: "A", Tipo: mina.TipoEstacionamiento},
		"B": {ID: "B", Tipo: mina.TipoPala},
		"C": {ID: "C", Tipo: mina.TipoInterseccion},
	}, Tramos: map[string]*mina.Tramo{
		"T1": {ID: "T1", NodoA: "A", NodoB: "B", Tipo: mina.TipoDoble, Tiempo: 1},
		"T2": {ID: "T2", NodoA: "A", NodoB: "C", Tipo: mina.TipoDoble, Tiempo: 1},
		"T3": {ID: "T3", NodoA: "C", NodoB: "B", Tipo: mina.TipoDoble, Tiempo: 1},
	}}
	d := NuevoDespachador(m)
	solicitudes := make(chan Solicitud)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	terminado := make(chan struct{})
	go func() { defer close(terminado); d.Ejecutar(ctx, solicitudes) }()
	respuesta := make(chan Resultado, 1)
	solicitudes <- Solicitud{CamionID: "C-01", Origen: "A", BloqueosDireccionales: map[string]map[string]bool{"T1": {"A->B": true}}, Respuesta: respuesta}
	r := <-respuesta
	if r.Err != nil {
		t.Fatal(r.Err)
	}
	if len(r.Ruta) != 3 || r.Ruta[0] != "A" || r.Ruta[1] != "C" || r.Ruta[2] != "B" {
		t.Fatalf("ruta inesperada: %v", r.Ruta)
	}
	cancel()
	<-terminado
}
