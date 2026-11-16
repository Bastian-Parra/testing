package camion

import (
	"context"
	"fmt"
	"os"
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
	mu     sync.RWMutex // protege el acceso al estado
	estado Estado

	// permisoCarga solamente se escucha cuando el camión
	// se encuentra esperando en la cola de una pala.
	//
	// permisoDescarga solamente se escucha cuando el camión
	// está esperando en un stockpile.
	permisoCarga    chan struct{}
	permisoDescarga chan struct{}
	permisoTramo    chan struct{} // canal para que el tramo avise que puede avanzar

	// Se cierra cuando la goroutine del camión termina.
	finalizado chan struct{}

	logFile *os.File // se modifica esto para agregar logging -> Vinculado a C3
	seq int // contador para logging -> Vinculado a C3

	cargado bool // indica si el camión está cargado o no
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

	rutaLog := fmt.Sprintf("../paradigmas/logs/sin_fallas/log_%s.csv", id)
	os.MkdirAll("../paradigmas/logs/sin_fallas", os.ModePerm)
	file, err	:= os.Create(rutaLog)
	if err != nil {
		return nil, err
	}

	file.WriteString("Seq,Tick,Entidad,Evento,Recurso,Toneladas\n")

	return &Camion{
		ID:              id,
		config:          config,
		estado:          EstadoViajandoVacio,
		permisoCarga:    make(chan struct{}),
		permisoDescarga: make(chan struct{}),
		permisoTramo:    make(chan struct{}),
		finalizado:      make(chan struct{}),
		logFile: file,
		seq: 1,
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
		c.registrarEvento(2, "libera", "Tramo-Norte", 0)
		// falta por hacer: integrar con el ruteador luego
		c.cambiarEstado(EstadoEnColaPala)

	case EstadoCargando:
		c.cargado = true
		c.cambiarEstado(EstadoEnColaTramo)
		// falta por hacer: integrar con el ruteador luego
		c.registrarEvento(4, "adquiere", "Tramo-Sur", 50)

	case EstadoTransportando:
		c.registrarEvento(6, "libera", "Tramo-Sur", 50)
		c.cambiarEstado(EstadoEnColaStockpile)

	case EstadoDescargando:
		c.cargado = false
		c.registrarEvento(8, "libera", "Stockpile-1", 0)
		c.cambiarEstado(EstadoEnColaTramo)
		c.registrarEvento(9, "adquiere", "Tramo-Norte", 0)
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

	if (c.logFile != nil) {
		defer c.logFile.Close()
	}

	c.registrarEvento(1, "adquiere", "Tramo-Norte", 0)

	for {
		estado := c.Estado()
		var temporizador <-chan time.Time

		if duracion := c.duracionEstado(estado); duracion > 0 {
			temporizador = time.After(duracion)
		}

		// -------------------------------------------------
		// Guardas de los canales
		// -------------------------------------------------

		var canalCarga <-chan struct{}
		var canalDescarga <-chan struct{}
		var canalTramo <-chan struct{}

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

		// solo se escucha si está esperando entrar a uno
		if estado == EstadoEnColaTramo {
			canalTramo = c.permisoTramo
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

		// FIFO cumplido: el primer camión en cola es autorizado a avanzar.
		case <-canalTramo:
		if c.cargado { 
				c.cambiarEstado(EstadoTransportando)
			} else {
				c.cambiarEstado(EstadoViajandoVacio)
			}
			c.registrarEvento(3, "adquiere", "Tramo", 0)
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

// registrarEvento agrega un evento al log del camión. (relacionado al C3)
func (c *Camion) registrarEvento(tick int, evento string, recurso string, toneladas int) {
	linea := fmt.Sprintf("%d,%d,%s,%s,%s,%d\n", c.seq, tick, c.ID, evento, recurso, toneladas)
	if c.logFile != nil {
		c.logFile.WriteString(linea)
	}
	c.seq++
}

// es llamado por goroutine del tramo	para avisarle al camión que puede avanzar
func (c *Camion) AutorizarTramo(ctx context.Context) {
	select {
	case c.permisoTramo <- struct{}{}:
	case	<-ctx.Done():
	}
}

// expone el canal de lectura para bloquear camion
func (c *Camion) EsperarPermisoTramo() <-chan struct{} {
	return c.permisoTramo
}