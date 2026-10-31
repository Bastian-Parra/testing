package mina

// TipoNodo representa los tipos de nodos permitidos en la mina.
type TipoNodo string

const (
	TipoEstacionamiento TipoNodo = "estacionamiento"
	TipoInterseccion    TipoNodo = "interseccion"
	TipoPala            TipoNodo = "pala"
	TipoStockpile       TipoNodo = "stockpile"
)

// TipoTramo representa los tipos de caminos permitidos.
type TipoTramo string

const (
	TipoDoble TipoTramo = "doble"
	TipoRampa TipoTramo = "rampa"
)

// Nodo representa un punto dentro del grafo de la mina.
//
// La capacidad solo se utiliza en los nodos de tipo pala.
// Para los otros tipos queda en 0.
type Nodo struct {
	ID        string
	Tipo      TipoNodo
	Capacidad int
}

// Tramo representa una conexión entre dos nodos.
//
// Tiempo indica cuántos ticks tarda un camión vacío
// en recorrer ese tramo.
type Tramo struct {
	ID     string
	NodoA  string
	NodoB  string
	Tipo   TipoTramo
	Tiempo int
}

// Mina contiene la representación completa del mapa.
type Mina struct {
	Nodos  map[string]*Nodo
	Tramos map[string]*Tramo
	Tabla  *TablaSimbolos
}

// NuevaMina crea una mina vacía.
func NuevaMina() *Mina {
	return &Mina{
		Nodos:  make(map[string]*Nodo),
		Tramos: make(map[string]*Tramo),
		Tabla:  NuevaTablaSimbolos(),
	}
}

// EsTipoNodoValido verifica si el tipo recibido pertenece
// a los tipos permitidos por el formato de la mina.
func EsTipoNodoValido(tipo TipoNodo) bool {
	switch tipo {
	case TipoEstacionamiento,
		TipoInterseccion,
		TipoPala,
		TipoStockpile:
		return true
	default:
		return false
	}
}

// EsTipoTramoValido verifica si un tramo es doble o rampa.
func EsTipoTramoValido(tipo TipoTramo) bool {
	switch tipo {
	case TipoDoble, TipoRampa:
		return true
	default:
		return false
	}
}
