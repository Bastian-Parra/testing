package despachador

import (
	"context"
	"testing"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/mina"
)

// Creamos una mina pequeña para probar el despachador
func crearMinaPrueba() *mina.Mina {
	m := mina.NuevaMina()

	m.Nodos["E1"] = &mina.Nodo{ID: "E1", Tipo: mina.TipoEstacionamiento}
	m.Nodos["P1"] = &mina.Nodo{ID: "P1", Tipo: mina.TipoPala, Capacidad: 2}
	m.Nodos["P2"] = &mina.Nodo{ID: "P2", Tipo: mina.TipoPala, Capacidad: 2}
	m.Nodos["S1"] = &mina.Nodo{ID: "S1", Tipo: mina.TipoStockpile}

	m.Tramos["T1"] = &mina.Tramo{
		ID: "T1", NodoA: "E1", NodoB: "P1",
		Tipo: mina.TipoDoble, Tiempo: 2,
	}
	m.Tramos["T2"] = &mina.Tramo{
		ID: "T2", NodoA: "E1", NodoB: "P2",
		Tipo: mina.TipoDoble, Tiempo: 3,
	}
	m.Tramos["T3"] = &mina.Tramo{
		ID: "T3", NodoA: "E1", NodoB: "S1",
		Tipo: mina.TipoDoble, Tiempo: 4,
	}

	return m
}

func TestRoundRobinPalas(t *testing.T) {
	d := NuevoDespachador(crearMinaPrueba())

	esperados := []string{"P1", "P2", "P1", "P2"}

	for i, esperado := range esperados {
		resultado := d.asignar(Solicitud{
			CamionID: "C1",
			Origen:   "E1",
		})

		if resultado.Err != nil {
			t.Fatalf("solicitud %d: %v", i+1, resultado.Err)
		}

		if resultado.Destino != esperado {
			t.Errorf("esperaba %s, recibí %s",
				esperado, resultado.Destino)
		}
	}
}

func TestCamionCargado(t *testing.T) {
	d := NuevoDespachador(crearMinaPrueba())

	resultado := d.asignar(Solicitud{
		CamionID: "C1",
		Origen:   "E1",
		Cargado:  true,
	})

	if resultado.Err != nil {
		t.Fatal(resultado.Err)
	}

	if resultado.Destino != "S1" {
		t.Errorf("esperaba S1, recibí %s", resultado.Destino)
	}
}

func TestDestinoInalcanzable(t *testing.T) {
	d := NuevoDespachador(crearMinaPrueba())

	// Bloqueamos el camino hacia P1.
	resultado := d.asignar(Solicitud{
		CamionID: "C1",
		Origen:   "E1",
		Fallas:   map[string]bool{"T1": true},
	})

	if resultado.Err != nil {
		t.Fatal(resultado.Err)
	}

	if resultado.Destino != "P2" {
		t.Errorf("esperaba P2, recibí %s", resultado.Destino)
	}
}

func TestSinRuta(t *testing.T) {
	d := NuevoDespachador(crearMinaPrueba())

	resultado := d.asignar(Solicitud{
		CamionID: "C1",
		Origen:   "E1",
		Fallas: map[string]bool{
			"T1": true,
			"T2": true,
		},
	})

	if resultado.Err == nil {
		t.Error("se esperaba un error porque no hay palas alcanzables")
	}
}

func TestGoroutineDespachador(t *testing.T) {
	d := NuevoDespachador(crearMinaPrueba())

	ctx, cancelar := context.WithCancel(context.Background())
	defer cancelar()

	solicitudes := make(chan Solicitud)
	respuestas := make(chan Resultado, 1)

	go d.Ejecutar(ctx, solicitudes)

	solicitudes <- Solicitud{
		CamionID:  "C1",
		Origen:    "E1",
		Respuesta: respuestas,
	}

	resultado := <-respuestas

	if resultado.Err != nil {
		t.Fatal(resultado.Err)
	}

	if resultado.Destino != "P1" {
		t.Errorf("esperaba P1, recibí %s", resultado.Destino)
	}
}
