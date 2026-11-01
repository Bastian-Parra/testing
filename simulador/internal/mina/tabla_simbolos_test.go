package mina

import "testing"

// Comprueba que un nodo pueda registrarse y recuperarse.
func TestTablaRegistraNodo(t *testing.T) {
	tabla := NuevaTablaSimbolos()

	nodo := &Nodo{
		ID:   "E1",
		Tipo: TipoEstacionamiento,
	}

	err := tabla.RegistrarNodo(nodo)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	simbolo, existe := tabla.Obtener("E1")

	if !existe {
		t.Fatal("E1 debería existir en la tabla")
	}

	if simbolo.Clase != SimboloNodo {
		t.Fatalf(
			"E1 debería ser un nodo, pero se obtuvo %q",
			simbolo.Clase,
		)
	}

	if simbolo.Nodo != nodo {
		t.Fatal("el nodo recuperado no corresponde al nodo registrado")
	}
}

// Comprueba que no se pueda declarar dos veces el mismo ID.
func TestTablaRechazaNodoDuplicado(t *testing.T) {
	tabla := NuevaTablaSimbolos()

	nodo := &Nodo{
		ID:   "P1",
		Tipo: TipoPala,
	}

	if err := tabla.RegistrarNodo(nodo); err != nil {
		t.Fatalf("no se esperaba error en el primer registro: %v", err)
	}

	err := tabla.RegistrarNodo(nodo)

	if err == nil {
		t.Fatal("se esperaba error por identificador duplicado")
	}
}

// Comprueba que un nodo y un tramo tampoco puedan compartir ID.
func TestTablaRechazaIDCompartido(t *testing.T) {
	tabla := NuevaTablaSimbolos()

	nodo := &Nodo{
		ID:   "X1",
		Tipo: TipoInterseccion,
	}

	if err := tabla.RegistrarNodo(nodo); err != nil {
		t.Fatal(err)
	}

	tramo := &Tramo{
		ID:     "X1",
		NodoA:  "E1",
		NodoB:  "I1",
		Tipo:   TipoDoble,
		Tiempo: 5,
	}

	if err := tabla.RegistrarTramo(tramo); err == nil {
		t.Fatal("se esperaba error porque X1 ya estaba registrado")
	}
}
