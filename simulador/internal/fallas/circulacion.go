package fallas

import (
	"fmt"
	"sync"
)

// Circulacion controla el orden de los camiones
// dentro de cada tramo y sentido.
type Circulacion struct {
	mu     sync.Mutex
	tramos map[string]map[string][]*CamionEnTramo
}

// CamionEnTramo guarda la informacion de un camion
// mientras se encuentra circulando por un tramo.
type CamionEnTramo struct {
	ID              string
	EnFalla         bool
	FallaPermanente bool
	TicksRestantes  int
	Avance          int // ticks de movimiento completados en este tramo
}

// NuevaCirculacion prepara las colas de los tramos.
func NuevaCirculacion() *Circulacion {
	return &Circulacion{
		tramos: make(map[string]map[string][]*CamionEnTramo),
	}
}

// Entrar agrega un camion al final de la cola.
// Tambien revisa que no este registrado en otro tramo.
func (c *Circulacion) Entrar(
	tramoID, sentido, camionID string,
) error {
	if tramoID == "" || sentido == "" || camionID == "" {
		return fmt.Errorf("tramo, sentido y camion son obligatorios")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Revisamos todas las colas para evitar que
	// el mismo camion este en dos tramos a la vez.
	for _, sentidos := range c.tramos {
		for _, cola := range sentidos {
			for _, camion := range cola {
				if camion.ID == camionID {
					return fmt.Errorf(
						"el camion %s ya esta circulando",
						camionID,
					)
				}
			}
		}
	}

	if c.tramos[tramoID] == nil {
		c.tramos[tramoID] = make(map[string][]*CamionEnTramo)
	}

	c.tramos[tramoID][sentido] = append(
		c.tramos[tramoID][sentido],
		&CamionEnTramo{ID: camionID},
	)

	return nil
}

// RegistrarFalla marca un camion como averiado.
// Si queda detenido, los que vienen detras no pueden adelantar.
func (c *Circulacion) RegistrarFalla(
	tramoID, sentido, camionID string,
	permanente bool,
	ticks int,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, camion := range c.tramos[tramoID][sentido] {
		if camion.ID == camionID {
			if camion.EnFalla {
				return fmt.Errorf(
					"el camion %s ya esta en falla",
					camionID,
				)
			}

			if !permanente && ticks <= 0 {
				return fmt.Errorf(
					"la reparacion necesita ticks positivos",
				)
			}

			camion.EnFalla = true
			camion.FallaPermanente = permanente
			camion.TicksRestantes = ticks

			return nil
		}
	}

	return fmt.Errorf("camion %s no encontrado", camionID)
}

// ConsultarCamion devuelve una copia de la informacion
// del camion sin exponer los datos internos.
func (c *Circulacion) ConsultarCamion(
	tramoID, sentido, camionID string,
) (CamionEnTramo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, camion := range c.tramos[tramoID][sentido] {
		if camion.ID == camionID {
			return *camion, nil
		}
	}

	return CamionEnTramo{}, fmt.Errorf(
		"camion %s no encontrado en el tramo %s",
		camionID,
		tramoID,
	)
}

// AvanzarReparacion descuenta un tick de reparacion.
// Una falla permanente nunca se repara automaticamente.
func (c *Circulacion) AvanzarReparacion(
	tramoID, sentido, camionID string,
) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, camion := range c.tramos[tramoID][sentido] {
		if camion.ID != camionID {
			continue
		}

		if !camion.EnFalla {
			return true, nil
		}

		if camion.FallaPermanente {
			return false, nil
		}

		if camion.TicksRestantes > 0 {
			camion.TicksRestantes--
		}

		if camion.TicksRestantes == 0 {
			camion.EnFalla = false
			return true, nil
		}

		return false, nil
	}

	return false, fmt.Errorf(
		"camion %s no encontrado",
		camionID,
	)
}

// PuedeSalir comprueba que el camion sea el primero
// de la cola y que no se encuentre averiado.
func (c *Circulacion) PuedeSalir(
	tramoID, sentido, camionID string,
) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	cola := c.tramos[tramoID][sentido]

	if len(cola) == 0 {
		return false
	}

	return cola[0].ID == camionID && !cola[0].EnFalla
}

// Salir permite abandonar el tramo solamente
// al primer camion de la cola.
func (c *Circulacion) Salir(
	tramoID, sentido, camionID string,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cola := c.tramos[tramoID][sentido]

	if len(cola) == 0 || cola[0].ID != camionID {
		return fmt.Errorf(
			"el camion %s no puede adelantar",
			camionID,
		)
	}

	if cola[0].EnFalla {
		return fmt.Errorf(
			"el camion %s esta en falla",
			camionID,
		)
	}

	c.tramos[tramoID][sentido] = cola[1:]

	return nil
}

// SnapshotOrden devuelve una copia del orden
// de los camiones dentro de un tramo.
func (c *Circulacion) SnapshotOrden(
	tramoID, sentido string,
) []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	cola := c.tramos[tramoID][sentido]
	orden := make([]string, 0, len(cola))

	for _, camion := range cola {
		orden = append(orden, camion.ID)
	}

	return orden
}

// PuedeAvanzar impide que un camion supere la posicion de quien va delante.
// Si el primero esta detenido, los demas esperan al alcanzar su avance.
func (c *Circulacion) PuedeAvanzar(tramoID, sentido, camionID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	cola := c.tramos[tramoID][sentido]
	for i, camion := range cola {
		if camion.ID != camionID {
			continue
		}
		if camion.EnFalla {
			return false
		}
		return i == 0 || camion.Avance < cola[i-1].Avance
	}
	return false
}

// AvanzarTick suma una unidad de avance si no se adelanta a otro camion.
func (c *Circulacion) AvanzarTick(tramoID, sentido, camionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	cola := c.tramos[tramoID][sentido]
	for i, camion := range cola {
		if camion.ID != camionID {
			continue
		}
		if camion.EnFalla || (i > 0 && camion.Avance >= cola[i-1].Avance) {
			return fmt.Errorf("camion %s no puede avanzar", camionID)
		}
		camion.Avance++
		return nil
	}
	return fmt.Errorf("camion %s no encontrado", camionID)
}

// RetirarAlTerminar retira un camion no averiado al finalizar el turno.
// No se usa durante la circulacion: no autoriza adelantamientos.
func (c *Circulacion) RetirarAlTerminar(tramoID, sentido, camionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cola := c.tramos[tramoID][sentido]
	for i, camion := range cola {
		if camion.ID == camionID && !camion.FallaPermanente {
			c.tramos[tramoID][sentido] = append(cola[:i:i], cola[i+1:]...)
			return
		}
	}
}

// MarcarPermanente convierte un camion ya registrado en obstaculo fisico.
func (c *Circulacion) MarcarPermanente(tramoID, sentido, camionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, camion := range c.tramos[tramoID][sentido] {
		if camion.ID == camionID {
			camion.EnFalla = true
			camion.FallaPermanente = true
			camion.TicksRestantes = 0
			return nil
		}
	}
	return fmt.Errorf("camion %s no encontrado", camionID)
}
