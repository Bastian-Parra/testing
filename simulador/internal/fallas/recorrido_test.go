package fallas

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/mina"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/rampa"
)

// prepararRecorrido crea una mina pequeña para probar
// los recorridos sin iniciar toda la simulacion.
func prepararRecorrido(t *testing.T, probFalla, probPermanente float64) *Recorrido {
	t.Helper()

	m := &mina.Mina{
		Nodos: map[string]*mina.Nodo{
			"A": {ID: "A"},
			"B": {ID: "B"},
			"C": {ID: "C"},
		},
		Tramos: map[string]*mina.Tramo{
			"T1": {
				ID: "T1", NodoA: "A", NodoB: "B",
				Tipo: mina.TipoDoble, Tiempo: 2,
			},
			"T2": {
				ID: "T2", NodoA: "B", NodoB: "C",
				Tipo: mina.TipoRampa, Tiempo: 2,
			},
		},
	}

	gestor, err := NuevoGestor(Configuracion{
		ProbabilidadFalla:      probFalla,
		ProbabilidadPermanente: probPermanente,
		ReparacionMinTicks:     2,
		ReparacionMaxTicks:     2,
		Tick:                   time.Millisecond,
		Semilla:                123,
	})
	if err != nil {
		t.Fatal(err)
	}

	controlador, err := NuevoControlador(gestor, NuevaCirculacion())
	if err != nil {
		t.Fatal(err)
	}

	recursoRampa, err := rampa.NuevaRampa(m.Tramos["T2"])
	if err != nil {
		t.Fatal(err)
	}

	return &Recorrido{
		Mina:        m,
		Controlador: controlador,
		Rampas: map[string]*rampa.Rampa{
			"T2": recursoRampa,
		},
		Tick:      time.Millisecond,
		Protocolo: rampa.ProtocoloPrevencion,
		Limite:    100 * time.Millisecond,
	}
}

func TestRecorridoSinFallas(t *testing.T) {
	recorrido := prepararRecorrido(t, 0, 0)

	ctx, cancelar := context.WithTimeout(
		context.Background(), time.Second,
	)
	defer cancelar()

	err := recorrido.Recorrer(ctx, "C1", []string{"A", "B", "C"})
	if err != nil {
		t.Fatal(err)
	}

	orden := recorrido.Controlador.Circulacion.SnapshotOrden(
		"T1", "A->B",
	)
	if len(orden) != 0 {
		t.Fatalf("el tramo deberia quedar libre: %v", orden)
	}

	_, ocupantes := recorrido.Rampas["T2"].Estado()
	if ocupantes != 0 {
		t.Fatalf("la rampa quedo ocupada: %d", ocupantes)
	}
}

func TestRecorridoFallaPermanente(t *testing.T) {
	recorrido := prepararRecorrido(t, 1, 1)

	ctx, cancelar := context.WithTimeout(
		context.Background(), 20*time.Millisecond,
	)
	defer cancelar()

	err := recorrido.Recorrer(ctx, "C1", []string{"A", "B"})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("se esperaba fin del contexto, obtenido: %v", err)
	}

	bloqueos := recorrido.Controlador.Gestor.SnapshotDireccional()
	if !bloqueos["T1"]["A->B"] {
		t.Fatal("la falla permanente no bloqueo el sentido")
	}

	orden := recorrido.Controlador.Circulacion.SnapshotOrden(
		"T1", "A->B",
	)
	if len(orden) != 1 || orden[0] != "C1" {
		t.Fatalf("el camion averiado debe permanecer: %v", orden)
	}
}

func TestRecorridoFallaTemporalCancelada(t *testing.T) {
	recorrido := prepararRecorrido(t, 1, 0)

	ctx, cancelar := context.WithTimeout(
		context.Background(), 20*time.Millisecond,
	)
	defer cancelar()

	// Con probabilidad de falla igual a 1, el camion
	// vuelve a fallar cada vez que termina su reparacion.
	err := recorrido.Recorrer(ctx, "C1", []string{"A", "B"})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("se esperaba cancelacion, obtenido: %v", err)
	}

	bloqueos := recorrido.Controlador.Gestor.SnapshotDireccional()
	if bloqueos["T1"]["A->B"] {
		t.Fatal("una falla temporal no debe bloquear permanentemente")
	}
}

func TestRecorridoRampaSinFallas(t *testing.T) {
	recorrido := prepararRecorrido(t, 1, 1)

	ctx, cancelar := context.WithTimeout(
		context.Background(), time.Second,
	)
	defer cancelar()

	// Aunque las probabilidades sean del 100%,
	// una rampa nunca debe generar fallas.
	err := recorrido.Recorrer(ctx, "C1", []string{"B", "C"})
	if err != nil {
		t.Fatal(err)
	}

	_, ocupantes := recorrido.Rampas["T2"].Estado()
	if ocupantes != 0 {
		t.Fatal("la rampa no fue liberada")
	}
}

func TestRecorridoRutaInvalida(t *testing.T) {
	recorrido := prepararRecorrido(t, 0, 0)

	err := recorrido.Recorrer(
		context.Background(), "C1", []string{"A", "C"},
	)
	if err == nil {
		t.Fatal("se esperaba error por tramo inexistente")
	}
}
