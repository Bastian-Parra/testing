package fallas

import "testing"

// Comprueba que los camiones no puedan adelantarse.
func TestCirculacionFIFO(t *testing.T) {
	c := NuevaCirculacion()

	for _, id := range []string{"C1", "C2", "C3"} {
		if err := c.Entrar("T1", "A->B", id); err != nil {
			t.Fatal(err)
		}
	}

	if c.PuedeSalir("T1", "A->B", "C2") {
		t.Fatal("C2 no puede adelantar a C1")
	}

	if err := c.Salir("T1", "A->B", "C1"); err != nil {
		t.Fatal(err)
	}

	if !c.PuedeSalir("T1", "A->B", "C2") {
		t.Fatal("C2 deberia poder salir despues de C1")
	}
}

// Una falla permanente bloquea a los camiones posteriores.
func TestFallaPermanenteBloqueaCola(t *testing.T) {
	c := NuevaCirculacion()

	_ = c.Entrar("T1", "A->B", "C1")
	_ = c.Entrar("T1", "A->B", "C2")

	if err := c.RegistrarFalla(
		"T1", "A->B", "C1", true, 0,
	); err != nil {
		t.Fatal(err)
	}

	if c.PuedeSalir("T1", "A->B", "C1") {
		t.Fatal("C1 no puede salir porque esta averiado")
	}

	if c.PuedeSalir("T1", "A->B", "C2") {
		t.Fatal("C2 no puede adelantar al camion averiado")
	}
}

// Una falla temporal permite continuar al repararse.
func TestFallaTemporalLiberaCola(t *testing.T) {
	c := NuevaCirculacion()

	_ = c.Entrar("T1", "A->B", "C1")
	_ = c.Entrar("T1", "A->B", "C2")

	if err := c.RegistrarFalla(
		"T1", "A->B", "C1", false, 3,
	); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		reparado, err := c.AvanzarReparacion(
			"T1", "A->B", "C1",
		)
		if err != nil {
			t.Fatal(err)
		}
		if reparado {
			t.Fatal("el camion se reparo antes de tiempo")
		}
	}

	reparado, err := c.AvanzarReparacion(
		"T1", "A->B", "C1",
	)
	if err != nil || !reparado {
		t.Fatal("el camion deberia estar reparado")
	}

	if !c.PuedeSalir("T1", "A->B", "C1") {
		t.Fatal("C1 deberia poder continuar")
	}

	if c.PuedeSalir("T1", "A->B", "C2") {
		t.Fatal("C2 debe esperar su turno")
	}
}

// El bloqueo en un sentido no impide circular en el contrario.
func TestSentidosIndependientes(t *testing.T) {
	c := NuevaCirculacion()

	_ = c.Entrar("T1", "A->B", "C1")
	_ = c.Entrar("T1", "B->A", "C2")

	_ = c.RegistrarFalla("T1", "A->B", "C1", true, 0)

	if !c.PuedeSalir("T1", "B->A", "C2") {
		t.Fatal("el sentido contrario no deberia bloquearse")
	}
}

// La copia del orden no puede modificar la cola original.
func TestSnapshotOrdenIndependiente(t *testing.T) {
	c := NuevaCirculacion()

	_ = c.Entrar("T1", "A->B", "C1")

	orden := c.SnapshotOrden("T1", "A->B")
	orden[0] = "OTRO"

	nuevo := c.SnapshotOrden("T1", "A->B")

	if nuevo[0] != "C1" {
		t.Fatal("el snapshot modifico la cola original")
	}
}
