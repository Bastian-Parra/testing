package camion

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Configuracion contiene los tiempos utilizados por el camión.
//
// Tick representa cuánto dura un minuto simulado en tiempo real.
// Los demás valores están expresados en ticks.
type Configuracion struct {
	Tick            time.Duration
	TicksViajeVacio int
	TicksCarga      int
	TicksTransporte int
	TicksDescarga   int
}

// Camion representa un camión autónomo del simulador.
//
// Cada camión posee su propio estado y sus propios canales.
// Su ciclo se ejecuta de manera independiente dentro de
// su propia goroutine.
type Camion struct {
	ID string

	config Configuracion

	// Protege el estado para que pueda ser consultado
	// desde otras goroutines sin producir carreras de datos.
	mu     sync.RWMutex
	estado Estado

	// Estos canales representan recursos/eventos externos.
	//
	// permisoCarga solamente se escucha cuando el camión
	// se encuentra esperando en la cola de una pala.
	//
	// permisoDescarga solamente se escucha cuando el camión
	// está esperando en un stockpile.
	permisoCarga    chan struct{}
	permisoDescarga chan struct{}

	// Se cierra cuando la goroutine del camión termina.
	finalizado chan struct{}
}

// NuevoCamion crea un camión listo para comenzar su ciclo.
func NuevoCamion(
	id string,
	config Configuracion,
) (*Camion, error) {

	if id == "" {
		return nil, fmt.Errorf("el camion debe tener un identificador")
	}

	if config.Tick <= 0 {
		return nil, fmt.Errorf("la duracion del tick debe ser mayor que cero")
	}

	if config.TicksViajeVacio <= 0 ||
		config.TicksCarga <= 0 ||
		config.TicksTransporte <= 0 ||
		config.TicksDescarga <= 0 {

		return nil, fmt.Errorf(
			"las duraciones del camion deben ser mayores que cero",
		)
	}

	return &Camion{
		ID:              id,
		config:          config,
		estado:          EstadoViajandoVacio,
		permisoCarga:    make(chan struct{}),
		permisoDescarga: make(chan struct{}),
		finalizado:      make(chan struct{}),
	}, nil
}

// Estado devuelve de forma segura el estado actual del camión.
func (c *Camion) Estado() Estado {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.estado
}

// cambiarEstado modifica el estado actual.
//
// Se mantiene privado para que las transiciones ocurran
// solamente dentro de la máquina de estados.
func (c *Camion) cambiarEstado(nuevo Estado) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.estado = nuevo
}

// duracionEstado devuelve cuánto tiempo debe permanecer
// automáticamente el camión en un determinado estado.
//
// Los estados de cola no tienen duración automática:
// dependen de que otro componente les conceda acceso.
func (c *Camion) duracionEstado(estado Estado) time.Duration {
	var ticks int

	switch estado {
	case EstadoViajandoVacio:
		ticks = c.config.TicksViajeVacio

	case EstadoCargando:
		ticks = c.config.TicksCarga

	case EstadoTransportando:
		ticks = c.config.TicksTransporte

	case EstadoDescargando:
		ticks = c.config.TicksDescarga

	default:
		return 0
	}

	return time.Duration(ticks) * c.config.Tick
}

// avanzarPorTiempo realiza las transiciones que ocurren
// automáticamente cuando termina el temporizador de un estado.
func (c *Camion) avanzarPorTiempo(estado Estado) {
	switch estado {

	case EstadoViajandoVacio:
		c.cambiarEstado(EstadoEnColaPala)

	case EstadoCargando:
		c.cambiarEstado(EstadoTransportando)

	case EstadoTransportando:
		c.cambiarEstado(EstadoEnColaStockpile)

	case EstadoDescargando:
		c.cambiarEstado(EstadoViajandoVacio)
	}
}

// Ejecutar contiene el ciclo principal del camión.
//
// Debe ejecutarse dentro de una goroutine:
//
//	go camion.Ejecutar(ctx)
//
// La máquina utiliza un select cuyos canales son habilitados
// o deshabilitados según el estado actual.
//
// Un canal nil nunca puede ser seleccionado por select.
// Esto permite utilizar los canales como comandos con guarda.
func (c *Camion) Ejecutar(ctx context.Context) {
	defer close(c.finalizado)

	for {
		estado := c.Estado()

		// -------------------------------------------------
		// Guarda del temporizador
		// -------------------------------------------------

		var temporizador <-chan time.Time

		if duracion := c.duracionEstado(estado); duracion > 0 {
			temporizador = time.After(duracion)
		}

		// -------------------------------------------------
		// Guardas de los canales
		// -------------------------------------------------

		var canalCarga <-chan struct{}
		var canalDescarga <-chan struct{}

		// El permiso de carga solamente tiene sentido
		// cuando el camión está esperando en una pala.
		if estado == EstadoEnColaPala {
			canalCarga = c.permisoCarga
		}

		// El permiso de descarga solamente tiene sentido
		// cuando espera en el stockpile.
		if estado == EstadoEnColaStockpile {
			canalDescarga = c.permisoDescarga
		}

		// -------------------------------------------------
		// Comando con guarda
		// -------------------------------------------------

		select {

		// Permite finalizar la goroutine ordenadamente.
		case <-ctx.Done():
			return

		// Solo está habilitado para estados que tienen
		// una duración automática.
		case <-temporizador:
			c.avanzarPorTiempo(estado)

		// canalCarga será nil en cualquier otro estado,
		// por lo que este caso queda deshabilitado.
		case <-canalCarga:
			c.cambiarEstado(EstadoCargando)

		// Lo mismo ocurre con el permiso de descarga.
		case <-canalDescarga:
			c.cambiarEstado(EstadoDescargando)
		}
	}
}

// AutorizarCarga permite que una pala indique al camión
// que puede comenzar a ser cargado.
//
// Si el camión todavía no se encuentra en EnColaPala,
// el envío queda esperando porque la guarda está deshabilitada.
func (c *Camion) AutorizarCarga(ctx context.Context) error {
	select {

	case c.permisoCarga <- struct{}{}:
		return nil

	case <-ctx.Done():
		return ctx.Err()

	case <-c.finalizado:
		return fmt.Errorf(
			"el camion %s ya finalizo",
			c.ID,
		)
	}
}

// AutorizarDescarga funciona de forma equivalente para
// la cola del stockpile.
func (c *Camion) AutorizarDescarga(ctx context.Context) error {
	select {

	case c.permisoDescarga <- struct{}{}:
		return nil

	case <-ctx.Done():
		return ctx.Err()

	case <-c.finalizado:
		return fmt.Errorf(
			"el camion %s ya finalizo",
			c.ID,
		)
	}
}

// Finalizado permite esperar el término de la goroutine
// sin exponer el canal para escritura.
func (c *Camion) Finalizado() <-chan struct{} {
	return c.finalizado
}
