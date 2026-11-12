package mina

import (
	"context"
	"fmt"
	"sync"
	"os"
	"time"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/camion"
)

// pala agrupa las propiedades de la estación de carga con struct
type Pala struct {
	ID string
	Cola chan *camion.Camion
	tiempoCarga time.Duration
	logFile	*os.File // se modifica esto para agregar logging -> Vinculado a C3
	seq int // contador para logging -> Vinculado a C3
}

// constructor de la pala, retorna la estructura y maneja errores
func NuevaPala(id string, tiempoCarga time.Duration, capacidaCola int) (*Pala, error) {
	if id == "" {
		return nil, fmt.Errorf("la pala debe tener un identificador")
	}
	if tiempoCarga <= 0 {
		return nil, fmt.Errorf("el tiempo de carga debe ser mayor a cero")
	}

	rutaLog := fmt.Sprintf("../paradigmas/logs/sin_fallas/log_%s.csv", id)
	file, err := os.Create(rutaLog)
	if err != nil {
		return nil, err
	}
	file.WriteString("Seq,Tick,Entidad,Evento,Recurso,Toneladas\n")

	return &Pala{
		ID: id,
		Cola: make(chan *camion.Camion, capacidaCola),
		tiempoCarga: tiempoCarga,
		logFile: file,
		seq: 1,
	}, nil
}

// se recibe un puntero (p *Pala) y recibe un WaitGroup para	sincronizar la ejecución de goroutines

func (p *Pala) Ejecutar(ctx context.Context, wg *sync.WaitGroup) {

	// el contador de WaitGroup se decrementa al finalizar la ejecución de la goroutine
	defer wg.Done()

	if p.logFile != nil {
		defer p.logFile.Close()
	}

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
				p.registrarEvento(10, "adquiere", camionActual.ID, 0)
				time.Sleep(1 * time.Millisecond) // simula el tiempo de llegada del camión a la pala
				p.registrarEvento(40, "carga", camionActual.ID, 0)
				p.registrarEvento(41, "libera", camionActual.ID, 0)
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

// registrarEvento agrega un evento al log de la pala.
func (p *Pala) registrarEvento(tick int, evento string, recurso string, toneladas int) {
	linea := fmt.Sprintf("%d,%d,%s,%s,%s,%d\n", p.seq, tick, p.ID, evento, recurso, toneladas)
	p.logFile.WriteString(linea)
	p.seq++
}