package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/camion"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/despachador"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/fallas"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/mina"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/rampa"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/turno"
)

func main() {
	escenario := flag.String("escenario", "sin_fallas", "sin_fallas o con_fallas")
	pf := flag.Float64("pf", 0.05, "probabilidad de falla por tick")
	pp := flag.Float64("pperm", 0.20, "probabilidad de falla permanente")
	seed := flag.Int64("seed", 42, "semilla reproducible")
	reparMin := flag.Int("repar-min", 3, "reparacion minima en ticks")
	reparMax := flag.Int("repar-max", 8, "reparacion maxima en ticks")
	ticks := flag.Int("ticks", 720, "duracion del turno en ticks")
	ms := flag.Int("tick-ms", 5, "duracion real de un tick en milisegundos")
	panicCam := flag.String("panic-camion", "", "ID del camion que tendra un panic controlado")
	panicAvance := flag.Int("panic-avance", 1, "avance del camion en que ocurre el panic")
	salida := flag.String("log", "logs/issue21.csv", "archivo CSV del escenario")
	flag.Parse()
	switch *escenario {
	case "sin_fallas":
		ejecutarSinFallas()
	case "con_fallas":
		ejecutarConFallas(*pf, *pp, *seed, *reparMin, *reparMax, *ticks, *ms, *panicCam, *panicAvance, *salida)
	default:
		log.Fatalf("escenario no reconocido: %s", *escenario)
	}
}

// Simulacion original del equipo: se conserva su funcionamiento.
func ejecutarSinFallas() {
	fmt.Println("Iniciando Simulación: Escenario sin_fallas")
	ctx, cancelar := context.WithCancel(context.Background())
	defer cancelar()
	var wg sync.WaitGroup

	// 1. Inicializar Pala
	pala1, err := mina.NuevaPala("Pala-01", 30*time.Millisecond, 2)
	if err != nil {
		fmt.Println("Error fatal creando la pala:", err)
		return
	}
	wg.Add(1)
	go pala1.Ejecutar(ctx, &wg)

	// 2. Inicializar Camiones
	configCam := camion.Configuracion{
		Tick:            10 * time.Millisecond,
		TicksViajeVacio: 2, TicksCarga: 3, TicksTransporte: 2, TicksDescarga: 2,
	}

	cam1, err := camion.NuevoCamion("C-01", configCam)
	if err != nil {
		fmt.Println("Error fatal creando el camión:", err)
		return
	}

	wg.Add(1)
	go func() { defer wg.Done(); cam1.Ejecutar(ctx) }()

	fmt.Println("camion a la pala")
	pala1.Cola <- cam1

	// simulación corriendo por un tiempo determinado
	time.Sleep(2 * time.Second)

	// finalizar simulación
	cancelar()
	wg.Wait()
	fmt.Println("Simulación finalizada. Logs generados.")
}

// ==========================================
// Sistema de fallas - Issue #21
// ==========================================
// Aqui ejecutamos el escenario de fallas con el despachador y el
// controlador de circulacion. El main original queda disponible.
func ejecutarConFallas(pf, pp float64, seed int64, reparMin, reparMax, ticks, ms int, panicCam string, panicAvance int, salida string) {
	if ticks <= 0 || ms <= 0 {
		log.Fatal("ticks y tick-ms deben ser positivos")
	}
	tick := time.Duration(ms) * time.Millisecond
	gestor, err := fallas.NuevoGestor(fallas.Configuracion{ProbabilidadFalla: pf, ProbabilidadPermanente: pp, ReparacionMinTicks: reparMin, ReparacionMaxTicks: reparMax, Tick: tick, Semilla: seed})
	if err != nil {
		log.Fatal(err)
	}
	m := &mina.Mina{Nodos: map[string]*mina.Nodo{"A": {ID: "A", Tipo: mina.TipoEstacionamiento}, "B": {ID: "B", Tipo: mina.TipoPala}, "C": {ID: "C", Tipo: mina.TipoInterseccion}, "D": {ID: "D", Tipo: mina.TipoStockpile}}, Tramos: map[string]*mina.Tramo{"T1": {ID: "T1", NodoA: "A", NodoB: "B", Tipo: mina.TipoDoble, Tiempo: 8}, "T2": {ID: "T2", NodoA: "A", NodoB: "C", Tipo: mina.TipoRampa, Tiempo: 5}, "T3": {ID: "T3", NodoA: "C", NodoB: "B", Tipo: mina.TipoDoble, Tiempo: 6}, "T4": {ID: "T4", NodoA: "B", NodoB: "D", Tipo: mina.TipoDoble, Tiempo: 4}, "T5": {ID: "T5", NodoA: "D", NodoB: "A", Tipo: mina.TipoDoble, Tiempo: 7}}}
	ram, err := rampa.NuevaRampa(m.Tramos["T2"])
	if err != nil {
		log.Fatal(err)
	}
	controlador, err := fallas.NuevoControlador(gestor, fallas.NuevaCirculacion())
	if err != nil {
		log.Fatal(err)
	}
	rec := &fallas.Recorrido{Mina: m, Controlador: controlador, Rampas: map[string]*rampa.Rampa{"T2": ram}, Tick: tick, Protocolo: rampa.ProtocoloPrevencion, Limite: 100 * tick}
	desp := despachador.NuevoDespachador(m)
	if err := os.MkdirAll(filepath.Dir(salida), 0755); err != nil {
		log.Fatal(err)
	}
	archivo, err := os.Create(salida)
	if err != nil {
		log.Fatal(err)
	}
	defer archivo.Close()

	// Formato de logs
	fmt.Fprintln(archivo, "Seq,Tick,Entidad,Evento,Recurso,Toneladas")

	var logMu sync.Mutex
	inicio := time.Now()
	secuencia := 0

	registrar := func(id, evento, tramo, sentido string, avance int) {
		t := int(time.Since(inicio) / tick)

		if t > ticks {
			t = ticks
		}

		logMu.Lock()
		defer logMu.Unlock()

		// Cada evento recibe un numero consecutivo.
		secuencia++

		recurso := tramo
		if sentido != "" {
			recurso = tramo + "_" + sentido
		}

		toneladas := ""

		fmt.Fprintf(archivo, "%d,%d,%s,%s,%s,%s\n",
			secuencia, t, id, evento, recurso, toneladas)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	// El despachador original asigna destinos; cada nueva solicitud
	// recibe una copia actual de los bloqueos direccionales.
	solicitudes := make(chan despachador.Solicitud)
	wg.Add(1)
	go func() { defer wg.Done(); desp.Ejecutar(ctx, solicitudes) }()
	// Sistema de fallas - Issue #21: reutilizamos los recursos originales.
	// Se crean sus carpetas de logs antes de iniciar las goroutines.
	if err := os.MkdirAll("../paradigmas/logs/sin_fallas", 0755); err != nil {
		log.Fatal(err)
	}
	pala, err := mina.NuevaPala("Pala-01", 3*tick, 3)
	if err != nil {
		log.Fatal(err)
	}
	stockpile, err := mina.NuevaStockpile("Stockpile-01", 3*tick, 3)
	if err != nil {
		log.Fatal(err)
	}
	wg.Add(2)
	go pala.Ejecutar(ctx, &wg)
	go stockpile.Ejecutar(ctx, &wg)

	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("C-%02d", i)
		// Estos son los mismos camiones usados en el escenario original.
		c, err := camion.NuevoCamion(id, camion.Configuracion{
			Tick: tick, TicksViajeVacio: 2, TicksCarga: 3,
			TicksTransporte: 2, TicksDescarga: 3,
		})
		if err != nil {
			log.Fatal(err)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := turno.EjecutarCamionOriginal(ctx, c, rec, gestor,
				solicitudes, pala, stockpile, tick,
				panicAvance, id == panicCam, registrar)
			if err != nil && !turno.EsCancelacion(err) && ctx.Err() == nil {
				registrar(id, "termina_por_error", "", "", 0)
				fmt.Fprintln(os.Stderr, id, err)
			}
		}()
	}
	// El turno se detiene despues de los ticks solicitados.
	timer := time.NewTimer(time.Duration(ticks) * tick)
	<-timer.C
	cancel()
	wg.Wait()
	fmt.Printf("Turno finalizado: %d ticks. Log: %s\n", ticks, salida)
	fmt.Printf("Bloqueos finales: %v\n", gestor.SnapshotDireccional())
	for _, tramo := range []string{"T1", "T3", "T4", "T5"} {
		for _, sentido := range []string{"A->B", "B->A", "C->B", "B->C", "B->D", "D->B", "D->A", "A->D"} {
			if orden := controlador.Circulacion.SnapshotOrden(tramo, sentido); len(orden) > 0 {
				fmt.Printf("Tramo %s sentido %s: %v\n", tramo, sentido, orden)
			}
		}
	}
}
