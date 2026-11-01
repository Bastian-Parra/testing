package mina

import (
	"strings"
	"testing"
)

// Comprueba que una mina correctamente escrita pueda cargarse.
func TestLeerMinaValida(t *testing.T) {
	entrada := strings.NewReader(`
		NODO E1 estacionamiento
		NODO I1 interseccion
		NODO P1 pala 3
		NODO S1 stockpile

		TRAMO T1 E1 I1 doble 5
		TRAMO R1 I1 P1 rampa 4
		TRAMO T2 I1 S1 doble 8
	`)

	m, err := Leer(entrada)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if len(m.Nodos) != 4 {
		t.Fatalf(
			"se esperaban 4 nodos, se obtuvieron %d",
			len(m.Nodos),
		)
	}

	if len(m.Tramos) != 3 {
		t.Fatalf(
			"se esperaban 3 tramos, se obtuvieron %d",
			len(m.Tramos),
		)
	}

	if m.Nodos["P1"].Capacidad != 3 {
		t.Fatalf(
			"P1 debería tener capacidad 3, tiene %d",
			m.Nodos["P1"].Capacidad,
		)
	}

	if m.Tramos["R1"].Tipo != TipoRampa {
		t.Fatal("R1 debería ser una rampa")
	}

	if m.Tabla.Cantidad() != 7 {
		t.Fatalf(
			"se esperaban 7 símbolos, se obtuvieron %d",
			m.Tabla.Cantidad(),
		)
	}
}

// Comprueba que comentarios y líneas vacías se ignoren.
func TestLeerIgnoraComentarios(t *testing.T) {
	entrada := strings.NewReader(`
		# comentario inicial

		NODO E1 estacionamiento
		NODO I1 interseccion

		TRAMO T1 E1 I1 doble 5 # comentario final
	`)

	m, err := Leer(entrada)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if len(m.Tramos) != 1 {
		t.Fatalf(
			"se esperaba 1 tramo, se obtuvieron %d",
			len(m.Tramos),
		)
	}
}

// Comprueba que no se puedan repetir identificadores de nodo.
func TestLeerRechazaNodoDuplicado(t *testing.T) {
	entrada := strings.NewReader(`
		NODO E1 estacionamiento
		NODO E1 interseccion
	`)

	_, err := Leer(entrada)

	if err == nil {
		t.Fatal("se esperaba error por nodo duplicado")
	}
}

// Comprueba que todos los nodos usados por un tramo existan.
func TestLeerRechazaNodoInexistente(t *testing.T) {
	entrada := strings.NewReader(`
		NODO E1 estacionamiento

		TRAMO T1 E1 I99 doble 5
	`)

	_, err := Leer(entrada)

	if err == nil {
		t.Fatal("se esperaba error porque I99 no fue declarado")
	}
}

// Comprueba que una pala tenga una capacidad definida.
func TestLeerRechazaPalaSinCapacidad(t *testing.T) {
	entrada := strings.NewReader(`
		NODO P1 pala
	`)

	_, err := Leer(entrada)

	if err == nil {
		t.Fatal("se esperaba error porque P1 no tiene capacidad")
	}
}

// Comprueba que el tiempo de recorrido sea mayor que cero.
func TestLeerRechazaTiempoInvalido(t *testing.T) {
	entrada := strings.NewReader(`
		NODO E1 estacionamiento
		NODO I1 interseccion

		TRAMO T1 E1 I1 doble 0
	`)

	_, err := Leer(entrada)

	if err == nil {
		t.Fatal("se esperaba error por tiempo inválido")
	}
}

// Comprueba que solo existan los tipos de tramo permitidos.
func TestLeerRechazaTipoTramoInvalido(t *testing.T) {
	entrada := strings.NewReader(`
		NODO E1 estacionamiento
		NODO I1 interseccion

		TRAMO T1 E1 I1 autopista 5
	`)

	_, err := Leer(entrada)

	if err == nil {
		t.Fatal("se esperaba error por tipo de tramo inválido")
	}
}

// Comprueba nuestra decisión de hacer la lectura en dos pasadas.
//
// El tramo puede aparecer antes de sus nodos porque primero se
// recopilan todos los nodos y recién después se procesan los tramos.
func TestLeerPermiteTramoAntesDeNodos(t *testing.T) {
	entrada := strings.NewReader(`
		TRAMO T1 E1 I1 doble 5

		NODO E1 estacionamiento
		NODO I1 interseccion
	`)

	m, err := Leer(entrada)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if _, existe := m.Tramos["T1"]; !existe {
		t.Fatal("T1 debería haber sido registrado")
	}
}
