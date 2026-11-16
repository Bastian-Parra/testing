package ruteo

import (
	"testing"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/mina"
)

// Creamos una mina pequeña para probar las rutas.
func minaPruebaDireccional() *mina.Mina {
	return &mina.Mina{
		Nodos: map[string]*mina.Nodo{
			"A": {ID: "A", Tipo: mina.TipoInterseccion},
			"B": {ID: "B", Tipo: mina.TipoInterseccion},
			"C": {ID: "C", Tipo: mina.TipoInterseccion},
		},
		Tramos: map[string]*mina.Tramo{
			"T1": {
				ID: "T1", NodoA: "A", NodoB: "B",
				Tipo: mina.TipoDoble, Tiempo: 2,
			},
			"T2": {
				ID: "T2", NodoA: "A", NodoB: "C",
				Tipo: mina.TipoDoble, Tiempo: 5,
			},
			"T3": {
				ID: "T3", NodoA: "C", NodoB: "B",
				Tipo: mina.TipoDoble, Tiempo: 5,
			},
		},
	}
}

func TestRutaDireccionalSinBloqueos(t *testing.T) {
	p := NuevoPlanificador(minaPruebaDireccional())

	ruta, err := p.CalcularRutaDireccional(
		"A", "B", false, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(ruta) != 2 || ruta[0] != "A" || ruta[1] != "B" {
		t.Fatalf("ruta inesperada: %v", ruta)
	}
}

func TestRutaDireccionalBloqueoUnSentido(t *testing.T) {
	p := NuevoPlanificador(minaPruebaDireccional())

	bloqueos := map[string]map[string]bool{
		"T1": {"A->B": true},
	}

	// De A hacia B debe buscar una ruta alternativa.
	rutaAB, err := p.CalcularRutaDireccional(
		"A", "B", false, bloqueos,
	)
	if err != nil {
		t.Fatal(err)
	}

	esperada := []string{"A", "C", "B"}

	if len(rutaAB) != len(esperada) {
		t.Fatalf("ruta incorrecta: %v", rutaAB)
	}

	for i := range esperada {
		if rutaAB[i] != esperada[i] {
			t.Fatalf("ruta incorrecta: %v", rutaAB)
		}
	}

	// El sentido contrario debe seguir funcionando.
	rutaBA, err := p.CalcularRutaDireccional(
		"B", "A", false, bloqueos,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(rutaBA) != 2 || rutaBA[0] != "B" || rutaBA[1] != "A" {
		t.Fatalf("el sentido contrario fue bloqueado: %v", rutaBA)
	}
}

func TestRutaDireccionalBloqueoAmbosSentidos(t *testing.T) {
	p := NuevoPlanificador(minaPruebaDireccional())

	bloqueos := map[string]map[string]bool{
		"T1": {
			"A->B": true,
			"B->A": true,
		},
	}

	ruta, err := p.CalcularRutaDireccional(
		"B", "A", false, bloqueos,
	)
	if err != nil {
		t.Fatal(err)
	}

	esperada := []string{"B", "C", "A"}

	if len(ruta) != len(esperada) {
		t.Fatalf("ruta incorrecta: %v", ruta)
	}

	for i := range esperada {
		if ruta[i] != esperada[i] {
			t.Fatalf("ruta incorrecta: %v", ruta)
		}
	}
}
