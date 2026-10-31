package mina

import "fmt"

// ClaseSimbolo indica qué representa un identificador
// dentro de la tabla de símbolos.
type ClaseSimbolo string

const (
	SimboloNodo  ClaseSimbolo = "nodo"
	SimboloTramo ClaseSimbolo = "tramo"
)

// Simbolo representa una entrada dentro de la tabla.
//
// Un símbolo puede representar un nodo o un tramo.
// Solo uno de esos dos campos tendrá un valor distinto de nil.
type Simbolo struct {
	ID    string
	Clase ClaseSimbolo
	Nodo  *Nodo
	Tramo *Tramo
}

// TablaSimbolos mantiene todos los identificadores declarados
// en el archivo de la mina.
//
// Se utiliza un map porque permite buscar rápidamente un
// identificador mediante su nombre, por ejemplo "P1" o "R2".
type TablaSimbolos struct {
	simbolos map[string]Simbolo
}

// NuevaTablaSimbolos crea una tabla de símbolos vacía.
func NuevaTablaSimbolos() *TablaSimbolos {
	return &TablaSimbolos{
		simbolos: make(map[string]Simbolo),
	}
}

// RegistrarNodo agrega un nodo a la tabla.
//
// Antes de agregarlo se verifica que su identificador
// no haya sido usado anteriormente.
func (t *TablaSimbolos) RegistrarNodo(nodo *Nodo) error {
	if nodo == nil {
		return fmt.Errorf("no se puede registrar un nodo nulo")
	}

	if t.Existe(nodo.ID) {
		return fmt.Errorf(
			"identificador %q ya fue declarado",
			nodo.ID,
		)
	}

	t.simbolos[nodo.ID] = Simbolo{
		ID:    nodo.ID,
		Clase: SimboloNodo,
		Nodo:  nodo,
	}

	return nil
}

// RegistrarTramo agrega un tramo a la tabla.
//
// El identificador también debe ser único, incluso respecto
// de los nodos.
func (t *TablaSimbolos) RegistrarTramo(tramo *Tramo) error {
	if tramo == nil {
		return fmt.Errorf("no se puede registrar un tramo nulo")
	}

	if t.Existe(tramo.ID) {
		return fmt.Errorf(
			"identificador %q ya fue declarado",
			tramo.ID,
		)
	}

	t.simbolos[tramo.ID] = Simbolo{
		ID:    tramo.ID,
		Clase: SimboloTramo,
		Tramo: tramo,
	}

	return nil
}

// Existe indica si un identificador ya está registrado.
func (t *TablaSimbolos) Existe(id string) bool {
	_, existe := t.simbolos[id]
	return existe
}

// Obtener busca un símbolo utilizando su identificador.
//
// Devuelve el símbolo y un booleano que indica si fue encontrado.
func (t *TablaSimbolos) Obtener(id string) (Simbolo, bool) {
	simbolo, existe := t.simbolos[id]
	return simbolo, existe
}

// Cantidad devuelve la cantidad total de identificadores
// almacenados en la tabla.
func (t *TablaSimbolos) Cantidad() int {
	return len(t.simbolos)
}
