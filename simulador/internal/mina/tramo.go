package mina

import (
	"context"
	"sync"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/camion"
)


// tramo representa una parte de la ruta con capacidad limitada

func (t *Tramo) InicializarMonitor(capacidad int) {
	t.Capacidad = capacidad
	t.entrada = make(chan *camion.Camion)
	t.salida = make(chan *camion.Camion)
	t.cola = make([]*camion.Camion, 0)
	t.ocupantes = 0
}

// gorutine monitora del tramo
// select infinito: no hay condiciones de carrera
func (t *Tramo) Ejecutar(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case cam := <-t.entrada:
			// un camion quiere entrar al tramo
			if t.ocupantes < t.Capacidad {
				// hay espacio: camion entra al tramo
				t.ocupantes++
				cam.AutorizarTramo(ctx)
			} else {
				// esta lleno: camion queda en cola
				t.cola = append(t.cola, cam)
			}
		case <-t.salida:
			// un camion salio
			t.ocupantes--

			// si hay camiones en cola, avanza el primero (fifo)
			if len(t.cola) > 0 {
				siguientes := t.cola[0]
				t.cola = t.cola[1:]
				t.ocupantes++
				siguientes.AutorizarTramo(ctx)
			}
		}
	}
}

// es invocado por el camión para solicitar entrar al tramo:
func (t *Tramo) Entrar(cam *camion.Camion) {
	t.entrada <- cam
	// se bloquea hasta que el tramo le de permiso
	// <-cam.EsperarPermisoTramo() PENDIENTE: se integrará con el ruteador luego
}

// es invocado por el camión cuando termina de recorrer el tramo:
func (t *Tramo) Salir(cam *camion.Camion) {
	t.salida <- cam
}