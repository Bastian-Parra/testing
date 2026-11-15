package fallas

import (
	"fmt"
	"sync"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/mina"
)

// Controlador conecta el gestor de fallas
// con las colas de circulacion.
type Controlador struct {
	mu          sync.Mutex
	Gestor      *Gestor
	Circulacion *Circulacion
}

// NuevoControlador utiliza los componentes
// que ya fueron creados y probados.
func NuevoControlador(
	gestor *Gestor,
	circulacion *Circulacion,
) (*Controlador, error) {
	if gestor == nil || circulacion == nil {
		return nil, fmt.Errorf(
			"el gestor y la circulacion son obligatorios",
		)
	}

	return &Controlador{
		Gestor:      gestor,
		Circulacion: circulacion,
	}, nil
}

// IniciarTramo registra un camion al entrar
// a un tramo doble de la mina.
func (c *Controlador) IniciarTramo(
	tramo *mina.Tramo,
	sentido, camionID string,
) error {
	if tramo == nil {
		return fmt.Errorf("el tramo no puede ser nil")
	}

	if tramo.Tipo != mina.TipoDoble {
		return fmt.Errorf(
			"el control FIFO de fallas solo admite tramos dobles",
		)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Gestor.SnapshotDireccional()[tramo.ID][sentido] {
		return fmt.Errorf("sentido %s del tramo %s bloqueado", sentido, tramo.ID)
	}
	return c.Circulacion.Entrar(
		tramo.ID,
		sentido,
		camionID,
	)
}

// EvaluarCamion comprueba si ocurre una falla durante un tick.
// El mutex evita que dos evaluaciones del controlador
// se ejecuten al mismo tiempo.
func (c *Controlador) EvaluarCamion(
	tramo *mina.Tramo,
	sentido, camionID string,
) (Resultado, error) {
	if tramo == nil || tramo.Tipo != mina.TipoDoble {
		return Resultado{Tipo: SinFalla},
			fmt.Errorf("tramo doble invalido")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Comprobamos que el camion este registrado.
	actual, err := c.Circulacion.ConsultarCamion(
		tramo.ID,
		sentido,
		camionID,
	)
	if err != nil {
		return Resultado{Tipo: SinFalla}, err
	}

	// No volvemos a evaluar un camion averiado.
	if actual.EnFalla {
		return Resultado{Tipo: SinFalla},
			fmt.Errorf(
				"el camion %s ya esta en falla",
				camionID,
			)
	}

	// Evaluamos la probabilidad de falla del tick.
	resultado := c.Gestor.EvaluarTick(tramo)

	if resultado.Tipo == SinFalla {
		return resultado, nil
	}

	permanente := resultado.Tipo == FallaPermanente

	// Registramos la falla en la circulacion.
	err = c.Circulacion.RegistrarFalla(
		tramo.ID,
		sentido,
		camionID,
		permanente,
		resultado.ReparacionTicks,
	)
	if err != nil {
		return Resultado{Tipo: SinFalla}, err
	}

	// Una falla permanente bloquea este sentido.
	if permanente {
		err = c.Gestor.RegistrarPermanente(
			tramo.ID,
			sentido,
		)
		if err != nil {
			return resultado, err
		}
	}

	return resultado, nil
}

// RepararTick descuenta un tick de reparacion.
func (c *Controlador) RepararTick(
	tramoID, sentido, camionID string,
) (bool, error) {
	return c.Circulacion.AvanzarReparacion(
		tramoID,
		sentido,
		camionID,
	)
}

// FinalizarTramo permite salir respetando
// el orden FIFO de los camiones.
func (c *Controlador) FinalizarTramo(
	tramoID, sentido, camionID string,
) error {
	return c.Circulacion.Salir(
		tramoID,
		sentido,
		camionID,
	)
}

// RegistrarPanic deja un camion averiado permanentemente en el tramo.
func (c *Controlador) RegistrarPanic(tramoID, sentido, camionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.Circulacion.MarcarPermanente(tramoID, sentido, camionID); err != nil {
		return err
	}
	return c.Gestor.RegistrarPermanente(tramoID, sentido)
}
