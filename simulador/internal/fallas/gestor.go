package fallas

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/mina"
)

// TipoFalla indica el resultado de evaluar un tick.
type TipoFalla string

const (
	SinFalla        TipoFalla = "sin_falla"
	FallaTemporal   TipoFalla = "temporal"
	FallaPermanente TipoFalla = "permanente"
)

// Configuracion permite cambiar las probabilidades
// sin modificar el codigo del simulador.
type Configuracion struct {
	ProbabilidadFalla      float64
	ProbabilidadPermanente float64
	ReparacionMinTicks     int
	ReparacionMaxTicks     int
	Tick                   time.Duration
	Semilla                int64
}

// Resultado representa una falla de un camion.
type Resultado struct {
	Tipo            TipoFalla
	ReparacionTicks int
}

// Gestor administra las fallas de manera segura.
// El mutex protege el generador aleatorio y los bloqueos.
type Gestor struct {
	mu        sync.Mutex
	config    Configuracion
	aleatorio *rand.Rand
	bloqueos  map[string]map[string]bool
}

// NuevoGestor valida la configuracion inicial.
func NuevoGestor(config Configuracion) (*Gestor, error) {
	if config.ProbabilidadFalla < 0 ||
		config.ProbabilidadFalla > 1 ||
		config.ProbabilidadPermanente < 0 ||
		config.ProbabilidadPermanente > 1 {
		return nil, errors.New("las probabilidades deben estar entre 0 y 1")
	}

	if config.ReparacionMinTicks <= 0 ||
		config.ReparacionMaxTicks < config.ReparacionMinTicks {
		return nil, errors.New("rango de reparacion invalido")
	}

	if config.Tick <= 0 {
		return nil, errors.New("el tick debe ser positivo")
	}

	return &Gestor{
		config:    config,
		aleatorio: rand.New(rand.NewSource(config.Semilla)),
		bloqueos:  make(map[string]map[string]bool),
	}, nil
}

// EvaluarTick aplica la probabilidad de falla.
// Solo los tramos dobles pueden presentar fallas.
func (g *Gestor) EvaluarTick(tramo *mina.Tramo) Resultado {
	if tramo == nil || tramo.Tipo != mina.TipoDoble {
		return Resultado{Tipo: SinFalla}
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.aleatorio.Float64() >= g.config.ProbabilidadFalla {
		return Resultado{Tipo: SinFalla}
	}

	if g.aleatorio.Float64() < g.config.ProbabilidadPermanente {
		return Resultado{Tipo: FallaPermanente}
	}

	rango := g.config.ReparacionMaxTicks -
		g.config.ReparacionMinTicks + 1

	ticks := g.config.ReparacionMinTicks +
		g.aleatorio.Intn(rango)

	return Resultado{
		Tipo:            FallaTemporal,
		ReparacionTicks: ticks,
	}
}

// RegistrarPermanente conserva el bloqueo de un tramo.
// Se identifica el sentido para no bloquear automaticamente
// a los camiones que circulan en sentido contrario.
func (g *Gestor) RegistrarPermanente(
	tramoID, sentido string,
) error {
	if tramoID == "" || sentido == "" {
		return fmt.Errorf("tramo y sentido son obligatorios")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.bloqueos[tramoID] == nil {
		g.bloqueos[tramoID] = make(map[string]bool)
	}

	g.bloqueos[tramoID][sentido] = true
	return nil
}

// SnapshotPlanificador conserva compatibilidad
// con el planificador original de nuestros compañeros.
// Si un sentido esta bloqueado, marca todo el tramo.
func (g *Gestor) SnapshotPlanificador() map[string]bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	copia := make(map[string]bool)

	for tramoID, sentidos := range g.bloqueos {
		for _, bloqueado := range sentidos {
			if bloqueado {
				copia[tramoID] = true
				break
			}
		}
	}

	return copia
}

// SnapshotDireccional devuelve una copia de los bloqueos
// separados por tramo y sentido.
//
// Por ejemplo, si A->B esta bloqueado, B->A puede
// seguir disponible para calcular nuevas rutas.
//
// Copiamos tambien los mapas internos para evitar
// que otra goroutine modifique los datos originales.
func (g *Gestor) SnapshotDireccional() map[string]map[string]bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	copia := make(
		map[string]map[string]bool,
		len(g.bloqueos),
	)

	for tramoID, sentidos := range g.bloqueos {
		copia[tramoID] = make(
			map[string]bool,
			len(sentidos),
		)

		for sentido, bloqueado := range sentidos {
			copia[tramoID][sentido] = bloqueado
		}
	}

	return copia
}

// EsperarReparacion detiene al camion durante X ticks.
// La espera puede interrumpirse al finalizar el turno.
func (g *Gestor) EsperarReparacion(
	ctx context.Context,
	ticks int,
) error {
	if ticks <= 0 {
		return errors.New("los ticks deben ser positivos")
	}

	temporizador := time.NewTimer(
		time.Duration(ticks) * g.config.Tick,
	)
	defer temporizador.Stop()

	select {
	case <-temporizador.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
