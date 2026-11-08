package mina

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/camion"
)

type Stockpile struct {
	ID string
	Cola chan*camion.Camion
	tiempoDescarga	time.Duration
}

func NuevaStockpile(id string, tiempoDescarga time.Duration, capacidadCola int) (*Stockpile, error) {
	if id == "" {
		return nil, fmt.Errorf("el stockpile debe tener un identificador")
	}

	if tiempoDescarga <= 0 {
		return nil, fmt.Errorf("el tiempo de descarga debe ser mayor a cero")
	}

	return &Stockpile{
		ID: id,
		Cola: make(chan *camion.Camion, capacidadCola),
		tiempoDescarga: tiempoDescarga,
	}, nil
}

	func (s *Stockpile) Ejecutar(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case	camionActual := <-s.Cola:
				err := camionActual.AutorizarDescarga(ctx)
				if	err != nil {
					continue
			}
			select {
				case	<-time.After(s.tiempoDescarga): 
					// descarga terminada
				case	<-ctx.Done():
					return
			}
		}
	}
}