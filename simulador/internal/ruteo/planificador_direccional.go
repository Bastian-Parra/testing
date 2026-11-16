package ruteo

import (
	"errors"
	"fmt"
	"math"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/planners/externos/dijkstra"
)

// CalcularRutaDireccional busca una ruta evitando solamente
// los sentidos que tienen fallas permanentes.
//
// Los bloqueos se guardan usando el ID del tramo y
// el sentido con formato "NodoA->NodoB".
func (p *Planificador) CalcularRutaDireccional(
	origen, destino string,
	cargado bool,
	bloqueos map[string]map[string]bool,
) ([]string, error) {
	if p == nil || p.m == nil {
		return nil, errors.New("planificador sin mina")
	}

	g := dijkstra.NewGraph()

	strToInt := make(map[string]int)
	intToStr := make(map[int]string)

	i := 0
	for id := range p.m.Nodos {
		strToInt[id] = i
		intToStr[i] = id
		g.AddEmptyVertex(i)
		i++
	}

	if _, existe := strToInt[origen]; !existe {
		return nil, fmt.Errorf("origen %s no registrado", origen)
	}

	if _, existe := strToInt[destino]; !existe {
		return nil, fmt.Errorf("destino %s no registrado", destino)
	}

	for _, tramo := range p.m.Tramos {
		a, existeA := strToInt[tramo.NodoA]
		b, existeB := strToInt[tramo.NodoB]

		if !existeA || !existeB {
			return nil, fmt.Errorf(
				"el tramo %s tiene nodos no registrados",
				tramo.ID,
			)
		}

		if tramo.Tiempo <= 0 {
			return nil, fmt.Errorf(
				"el tramo %s tiene tiempo invalido",
				tramo.ID,
			)
		}

		peso := float64(tramo.Tiempo)

		// Un camion cargado tarda un 30% mas.
		if cargado {
			peso = math.Ceil(peso * 1.3)
		}

		sentidoAB := tramo.NodoA + "->" + tramo.NodoB
		sentidoBA := tramo.NodoB + "->" + tramo.NodoA

		// Agregamos cada sentido de manera independiente.
		if !bloqueos[tramo.ID][sentidoAB] {
			g.AddArc(a, b, uint64(peso))
		}

		if !bloqueos[tramo.ID][sentidoBA] {
			g.AddArc(b, a, uint64(peso))
		}
	}

	resultado, err := g.Shortest(
		strToInt[origen],
		strToInt[destino],
	)
	if err != nil {
		return nil, err
	}

	var ruta []string

	for _, nodo := range resultado.Path {
		ruta = append(ruta, intToStr[nodo])
	}

	return ruta, nil
}
