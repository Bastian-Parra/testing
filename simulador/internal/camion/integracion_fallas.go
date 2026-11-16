package camion

import (
	"context"
)

// EjecutarConFallas mantiene el ciclo de vida del camion original.
// El recorrido lo organiza el paquete turno, para no mezclar rutas con
// la maquina de estados ni crear dependencias circulares con mina.
func (c *Camion) EjecutarConFallas(ctx context.Context, ciclo func(func(Estado)) error) error {
	defer close(c.finalizado)
	if c.logFile != nil {
		defer c.logFile.Close()
	}
	return ciclo(c.cambiarEstado)
}

// EsperarCarga recibe la autorizacion enviada por la pala original.
func (c *Camion) EsperarCarga(ctx context.Context) error {
	select {
	case <-c.permisoCarga:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// EsperarDescarga recibe la autorizacion enviada por el stockpile original.
func (c *Camion) EsperarDescarga(ctx context.Context) error {
	select {
	case <-c.permisoDescarga:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
