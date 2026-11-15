package fallas

import "testing"

// Comprueba que un bloqueo afecta solamente
// al sentido donde ocurrio la falla.
func TestSnapshotDireccional(t *testing.T) {
	gestor, err := NuevoGestor(configPrueba())
	if err != nil {
		t.Fatal(err)
	}

	// Un camion falla permanentemente de A hacia B.
	err = gestor.RegistrarPermanente("T1", "A->B")
	if err != nil {
		t.Fatal(err)
	}

	snapshot := gestor.SnapshotDireccional()

	if !snapshot["T1"]["A->B"] {
		t.Fatal("el sentido A->B deberia estar bloqueado")
	}

	if snapshot["T1"]["B->A"] {
		t.Fatal("el sentido B->A debe seguir disponible")
	}
}

// Comprueba que modificar una copia no afecta
// los bloqueos guardados en el gestor.
func TestSnapshotDireccionalIndependiente(t *testing.T) {
	gestor, err := NuevoGestor(configPrueba())
	if err != nil {
		t.Fatal(err)
	}

	err = gestor.RegistrarPermanente("T1", "A->B")
	if err != nil {
		t.Fatal(err)
	}

	primeraCopia := gestor.SnapshotDireccional()

	// Modificamos la copia, no el gestor.
	primeraCopia["T1"]["A->B"] = false
	primeraCopia["T1"]["B->A"] = true
	primeraCopia["T2"] = map[string]bool{
		"C->D": true,
	}

	segundaCopia := gestor.SnapshotDireccional()

	if !segundaCopia["T1"]["A->B"] {
		t.Fatal("el bloqueo original no debe cambiar")
	}

	if segundaCopia["T1"]["B->A"] {
		t.Fatal("el sentido contrario no debe bloquearse")
	}

	if _, existe := segundaCopia["T2"]; existe {
		t.Fatal("el tramo agregado a la copia no debe existir")
	}
}
