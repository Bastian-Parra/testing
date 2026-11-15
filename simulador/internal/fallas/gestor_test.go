package fallas

import (
	"context"
	"testing"
	"time"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/mina"
)

// Configuración pequeña para ejecutar las pruebas rápidamente.
func configPrueba() Configuracion {
	return Configuracion{
		ProbabilidadFalla:      0.05,
		ProbabilidadPermanente: 0.20,
		ReparacionMinTicks:     3,
		ReparacionMaxTicks:     8,
		Tick:                   time.Millisecond,
		Semilla:                123,
	}
}

func TestConfiguracionInvalida(t *testing.T) {
	config := configPrueba()
	config.ProbabilidadFalla = 1.5

	if _, err := NuevoGestor(config); err == nil {
		t.Fatal("se esperaba un error por probabilidad inválida")
	}
}

func TestRampaNoFalla(t *testing.T) {
	config := configPrueba()
	config.ProbabilidadFalla = 1

	gestor, err := NuevoGestor(config)
	if err != nil {
		t.Fatal(err)
	}

	tramo := &mina.Tramo{
		ID: "R1", Tipo: mina.TipoRampa,
	}

	for i := 0; i < 100; i++ {
		if resultado := gestor.EvaluarTick(tramo); resultado.Tipo != SinFalla {
			t.Fatal("una rampa no debería presentar fallas")
		}
	}
}

func TestFallaTemporal(t *testing.T) {
	config := configPrueba()
	config.ProbabilidadFalla = 1
	config.ProbabilidadPermanente = 0

	gestor, err := NuevoGestor(config)
	if err != nil {
		t.Fatal(err)
	}

	resultado := gestor.EvaluarTick(&mina.Tramo{
		ID: "T1", Tipo: mina.TipoDoble,
	})

	if resultado.Tipo != FallaTemporal {
		t.Fatalf("se esperaba falla temporal, se obtuvo %s", resultado.Tipo)
	}

	if resultado.ReparacionTicks < 3 || resultado.ReparacionTicks > 8 {
		t.Fatalf("duración fuera del rango: %d", resultado.ReparacionTicks)
	}
}

func TestFallaPermanente(t *testing.T) {
	config := configPrueba()
	config.ProbabilidadFalla = 1
	config.ProbabilidadPermanente = 1

	gestor, err := NuevoGestor(config)
	if err != nil {
		t.Fatal(err)
	}

	resultado := gestor.EvaluarTick(&mina.Tramo{
		ID: "T1", Tipo: mina.TipoDoble,
	})

	if resultado.Tipo != FallaPermanente {
		t.Fatalf("se esperaba falla permanente, se obtuvo %s", resultado.Tipo)
	}
}

func TestSnapshotIndependiente(t *testing.T) {
	gestor, err := NuevoGestor(configPrueba())
	if err != nil {
		t.Fatal(err)
	}

	if err := gestor.RegistrarPermanente("T1", "A->B"); err != nil {
		t.Fatal(err)
	}

	snapshot := gestor.SnapshotPlanificador()

	if !snapshot["T1"] {
		t.Fatal("el tramo debería aparecer bloqueado")
	}

	// Modificamos la copia, no el estado original.
	delete(snapshot, "T1")

	nuevoSnapshot := gestor.SnapshotPlanificador()

	if !nuevoSnapshot["T1"] {
		t.Fatal("el snapshot modificó el estado del gestor")
	}
}

func TestReparacionCancelada(t *testing.T) {
	gestor, err := NuevoGestor(configPrueba())
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancelar := context.WithCancel(context.Background())
	cancelar()

	err = gestor.EsperarReparacion(ctx, 5)

	if err == nil {
		t.Fatal("se esperaba error al cancelar la reparación")
	}
}
