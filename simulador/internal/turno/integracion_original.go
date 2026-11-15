package turno

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/camion"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/despachador"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/fallas"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/mina"
)

// EjecutarCamionOriginal usa el camion, la pala y el stockpile del equipo.
// El despachador decide la ruta y el controlador se encarga de las fallas.
func EjecutarCamionOriginal(ctx context.Context, c *camion.Camion,
	rec *fallas.Recorrido, gestor *fallas.Gestor,
	solicitudes chan<- despachador.Solicitud,
	pala *mina.Pala, stockpile *mina.Stockpile, tick time.Duration,
	panicAvance int, provocarPanic bool,
	registrar func(string, string, string, string, int)) error {

	if c == nil || rec == nil || gestor == nil || pala == nil || stockpile == nil || tick <= 0 {
		return fmt.Errorf("faltan componentes para ejecutar el camion")
	}
	return c.EjecutarConFallas(ctx, func(cambiarEstado func(camion.Estado)) error {
		origen := "A"
		cargado := false
		panicUsado := false
		for ctx.Err() == nil {
			if cargado {
				cambiarEstado(camion.EstadoTransportando)
			} else {
				cambiarEstado(camion.EstadoViajandoVacio)
			}
			respuesta := make(chan despachador.Resultado, 1)
			solicitud := despachador.Solicitud{
				CamionID: c.ID, Origen: origen, Cargado: cargado,
				BloqueosDireccionales: gestor.SnapshotDireccional(), Respuesta: respuesta,
			}
			select {
			case solicitudes <- solicitud:
			case <-ctx.Done():
				return ctx.Err()
			}
			var asignacion despachador.Resultado
			select {
			case asignacion = <-respuesta:
			case <-ctx.Done():
				return ctx.Err()
			}
			if asignacion.Err != nil {
				registrar(c.ID, "sin_ruta", "", "", 0)
				cambiarEstado(camion.EstadoEsperando)
				return asignacion.Err
			}
			copia := *rec
			copia.Evento = func(id, evento, tramo, sentido string, avance int) {
				switch evento {
				case "temporal", "permanente", "panic_permanente", "reparando":
					cambiarEstado(camion.EstadoEnFalla)
				case "entra", "avanza", "sale":
					if cargado {
						cambiarEstado(camion.EstadoTransportando)
					} else {
						cambiarEstado(camion.EstadoViajandoVacio)
					}
				}
				registrar(id, evento, tramo, sentido, avance)
			}
			if provocarPanic && !panicUsado {
				copia.PanicEnTick = func(id, tramo string, avance int) bool {
					if avance == panicAvance {
						panicUsado = true
						return true
					}
					return false
				}
			}
			if err := copia.Recorrer(ctx, c.ID, asignacion.Ruta); err != nil {
				// Una falla permanente conserva EnFalla; no cambiamos su estado.
				return err
			}
			origen = asignacion.Destino
			if cargado {
				cambiarEstado(camion.EstadoEnColaStockpile)
				select {
				case stockpile.Cola <- c:
				case <-ctx.Done():
					return ctx.Err()
				}
				if err := c.EsperarDescarga(ctx); err != nil {
					return err
				}
				cambiarEstado(camion.EstadoDescargando)
				registrar(c.ID, "descarga", origen, "", 0)
			} else {
				cambiarEstado(camion.EstadoEnColaPala)
				select {
				case pala.Cola <- c:
				case <-ctx.Done():
					return ctx.Err()
				}
				if err := c.EsperarCarga(ctx); err != nil {
					return err
				}
				cambiarEstado(camion.EstadoCargando)
				registrar(c.ID, "carga", origen, "", 0)
			}
			// Los recursos originales conceden permiso al comenzar el servicio.
			// El camion espera la duracion del servicio antes de salir.
			timer := time.NewTimer(3 * tick)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			}
			cargado = !cargado
		}
		return ctx.Err()
	})
}

// EsCancelacion evita informar como error el cierre normal del turno.
func EsCancelacion(err error) bool {
	return err == nil || errors.Is(err, context.Canceled)
}
