package ruteo

import (
	"errors"
	"math"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/mina"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/planners/externos/dijkstra"
)

// planificador guarda el mapa de la mina
// empieza con mayúscula para poder usarlo desde otros paquetes
type Planificador struct {
	m *mina.Mina // usamos un puntero para no copiar toda la estructura en cada llamada
}

func NuevoPlanificador(m *mina.Mina) *Planificador {
	return &Planificador{m: m}
}

// calcularRuta devuelve la ruta y un error si algo sale mal
func (p *Planificador) CalcularRuta(origen, destino string, cargado bool, fallas map[string]bool) ([]string, error) {
	g := dijkstra.NewGraph()

	// make deja listos los mapas para usarlos
	strToInt := make(map[string]int)
	intToStr := make(map[int]string)

	// estos nombres son cortos porque son variables que se usan solo acá
	i := 0
	for id := range p.m.Nodos {
		strToInt[id] = i
		intToStr[i] = id
		g.AddEmptyVertex(i)
		i++
	}

	for _, t := range p.m.Tramos {
		// si el tramo está fallando, lo saltamos y seguimos con el siguiente
		if fallas[t.ID] {
			continue
		}

		// lo pasamos a float64 para poder usar math.ceil
		peso := float64(t.Tiempo)
		if cargado {
			peso = math.Ceil(peso * 1.3)
		}

		a := strToInt[t.NodoA]
		b := strToInt[t.NodoB]

		g.AddArc(a, b, uint64(peso))
		g.AddArc(b, a, uint64(peso))
	}

	origenID, ok1 := strToInt[origen]
	destinoID, ok2 := strToInt[destino]

	// si no existe el origen o el destino, avisamos y terminamos
	if !ok1 || !ok2 {
		return nil, errors.New("origen o destino no registrados")
	}

	best, err := g.Shortest(origenID, destinoID)
	if err != nil {
		return nil, err
	}

	// el slice parte vacío y append se encarga de ir agregando los nodos
	var ruta []string
	for _, n := range best.Path {
		ruta = append(ruta, intToStr[n])
	}

	return ruta, nil
}