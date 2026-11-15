package fallas

import (
	"sync"
	"testing"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/mina"
)

// Prepara un controlador con probabilidades conocidas.
func prepararControlador(
	t *testing.T,
	probFalla, probPermanente float64,
) *Controlador {
	t.Helper()

	config := configPrueba()
	config.ProbabilidadFalla = probFalla
	config.ProbabilidadPermanente = probPermanente

	gestor, err := NuevoGestor(config)
	if err != nil {
		t.Fatal(err)
	}

	controlador, err := NuevoControlador(
		gestor,
		NuevaCirculacion(),
	)
	if err != nil {
		t.Fatal(err)
	}

	return controlador
}

func TestControladorSinFallas(t *testing.T) {
	c := prepararControlador(t, 0, 0)

	tramo := &mina.Tramo{
		ID:   "T1",
		Tipo: mina.TipoDoble,
	}

	if err := c.IniciarTramo(tramo, "A->B", "C1"); err != nil {
		t.Fatal(err)
	}

	resultado, err := c.EvaluarCamion(tramo, "A->B", "C1")
	if err != nil {
		t.Fatal(err)
	}

	if resultado.Tipo != SinFalla {
		t.Fatal("no deberia ocurrir una falla")
	}

	if err := c.FinalizarTramo("T1", "A->B", "C1"); err != nil {
		t.Fatal(err)
	}
}

func TestControladorFallaPermanente(t *testing.T) {
	c := prepararControlador(t, 1, 1)

	tramo := &mina.Tramo{
		ID:   "T1",
		Tipo: mina.TipoDoble,
	}

	_ = c.IniciarTramo(tramo, "A->B", "C1")
	_ = c.IniciarTramo(tramo, "A->B", "C2")

	resultado, err := c.EvaluarCamion(tramo, "A->B", "C1")
	if err != nil {
		t.Fatal(err)
	}

	if resultado.Tipo != FallaPermanente {
		t.Fatal("se esperaba una falla permanente")
	}

	if c.Circulacion.PuedeSalir("T1", "A->B", "C2") {
		t.Fatal("C2 no puede adelantar al camion averiado")
	}

	snapshot := c.Gestor.SnapshotPlanificador()

	if !snapshot["T1"] {
		t.Fatal("el tramo debe aparecer bloqueado")
	}
}

func TestControladorFallaTemporal(t *testing.T) {
	c := prepararControlador(t, 1, 0)

	tramo := &mina.Tramo{
		ID:   "T1",
		Tipo: mina.TipoDoble,
	}

	if err := c.IniciarTramo(tramo, "A->B", "C1"); err != nil {
		t.Fatal(err)
	}

	resultado, err := c.EvaluarCamion(tramo, "A->B", "C1")
	if err != nil {
		t.Fatal(err)
	}

	if resultado.Tipo != FallaTemporal {
		t.Fatal("se esperaba una falla temporal")
	}

	for i := 0; i < resultado.ReparacionTicks; i++ {
		reparado, err := c.RepararTick("T1", "A->B", "C1")
		if err != nil {
			t.Fatal(err)
		}

		if i == resultado.ReparacionTicks-1 && !reparado {
			t.Fatal("el camion deberia estar reparado")
		}
	}

	if !c.Circulacion.PuedeSalir("T1", "A->B", "C1") {
		t.Fatal("el camion reparado deberia continuar")
	}
}

func TestCamionNoPuedeEstarEnDosTramos(t *testing.T) {
	c := prepararControlador(t, 0, 0)

	tramo1 := &mina.Tramo{
		ID:   "T1",
		Tipo: mina.TipoDoble,
	}

	tramo2 := &mina.Tramo{
		ID:   "T2",
		Tipo: mina.TipoDoble,
	}

	// Registramos el camion en el primer tramo.
	err := c.IniciarTramo(tramo1, "A->B", "C1")
	if err != nil {
		t.Fatal(err)
	}

	// Intentamos ingresar el mismo camion a otro tramo.
	err = c.IniciarTramo(tramo2, "B->C", "C1")

	if err == nil {
		t.Fatal("el camion no deberia estar en dos tramos")
	}

	// Comprobamos que siga solamente en el primero.
	orden := c.Circulacion.SnapshotOrden("T1", "A->B")

	if len(orden) != 1 || orden[0] != "C1" {
		t.Fatal("el camion deberia seguir en el primer tramo")
	}
}

func TestCamionAveriadoNoSeEvaluaOtraVez(t *testing.T) {
	c := prepararControlador(t, 1, 1)

	tramo := &mina.Tramo{
		ID:   "T1",
		Tipo: mina.TipoDoble,
	}

	err := c.IniciarTramo(tramo, "A->B", "C1")
	if err != nil {
		t.Fatal(err)
	}

	// Forzamos una falla permanente.
	resultado, err := c.EvaluarCamion(
		tramo, "A->B", "C1",
	)
	if err != nil {
		t.Fatal(err)
	}

	if resultado.Tipo != FallaPermanente {
		t.Fatal("se esperaba una falla permanente")
	}

	// Un camion averiado no debe volver a evaluarse.
	_, err = c.EvaluarCamion(tramo, "A->B", "C1")

	if err == nil {
		t.Fatal("el camion averiado no debe reevaluarse")
	}

	// Comprobamos que la falla siga siendo permanente.
	actual, err := c.Circulacion.ConsultarCamion(
		"T1", "A->B", "C1",
	)
	if err != nil {
		t.Fatal(err)
	}

	if !actual.EnFalla || !actual.FallaPermanente {
		t.Fatal("la falla permanente debe mantenerse")
	}
}

func TestEvaluacionConcurrenteCamion(t *testing.T) {
	c := prepararControlador(t, 1, 1)

	tramo := &mina.Tramo{
		ID:   "T1",
		Tipo: mina.TipoDoble,
	}

	if err := c.IniciarTramo(tramo, "A->B", "C1"); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup

	resultados := make(chan error, 10)

	// Diez goroutines intentan evaluar al mismo camion.
	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, err := c.EvaluarCamion(
				tramo, "A->B", "C1",
			)

			resultados <- err
		}()
	}

	wg.Wait()
	close(resultados)

	exitos := 0
	rechazos := 0

	for err := range resultados {
		if err == nil {
			exitos++
		} else {
			rechazos++
		}
	}

	// Solamente una evaluacion debe registrar la falla.
	if exitos != 1 || rechazos != 9 {
		t.Fatalf(
			"se esperaba 1 exito y 9 rechazos; hubo %d exitos y %d rechazos",
			exitos,
			rechazos,
		)
	}
}
