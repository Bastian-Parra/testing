package mina

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/camion"
)

// pala agrupa las propiedades de la estación de carga con struct
type Pala struct {
	ID string
	Cola chan *camion.Camion
	tiempoCarga time.Duration
}

// constructor de la pala, retorna la estructura y maneja errores
func NuevaPala(id string, tiempoCarga time.Duration, capacidaCola int) (*Pala, error) {
	if id == "" {
		return nil, fmt.Errorf("la pala debe tener un identificador")
	}
	if tiempoCarga <= 0 {
		return nil, fmt.Errorf("el tiempo de carga debe ser mayor a cero")
	}

	return &Pala{
		ID: id,
		Cola: make(chan *camion.Camion, capacidaCola),
		tiempoCarga: tiempoCarga,
	}, nil
}

// se recibe un puntero (p *Pala) y recibe un WaitGroup para	sincronizar la ejecución de goroutines

func (p *Pala) Ejecutar(ctx context.Context, wg *sync.WaitGroup) {

	// el contador de WaitGroup se decrementa al finalizar la ejecución de la goroutine
	defer wg.Done()

	for {
		// multiples canales simultaneamente, se espera a que un camión llegue a la pala o que el contexto se cancele
		select {
			case	<-ctx.Done():
				return
			case camionActual := <-p.Cola:
				err := camionActual.AutorizarCarga(ctx)
				if err != nil {
					continue // si hay error, vuelve al ciclo for	para esperar otro camión
				}
				select {
				case <-time.After(p.tiempoCarga):
					// carga terminada
				case <-ctx.Done():
					return
				}
		}
	}
}