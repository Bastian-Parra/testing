package ruteo

import (
	"testing"
	"time"

	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/fallas"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/mina"
)

// Comprueba que una falla permanente se considere
// al calcular nuevas rutas, sin bloquear el sentido contrario.
func TestIntegracionFallasRuteoDireccional(t *testing.T) {
	m := &mina.Mina{
		Nodos: map[string]*mina.Nodo{
			"A": {ID: "A"},
			"B": {ID: "B"},
			"C": {ID: "C"},
		},
		Tramos: map[string]*mina.Tramo{
			"T1": {
				ID: "T1", NodoA: "A", NodoB: "B",
				Tipo: mina.TipoDoble, Tiempo: 2,
			},
			"T2": {
				ID: "T2", NodoA: "A", NodoB: "C",
				Tipo: mina.TipoDoble, Tiempo: 5,
			},
			"T3": {
				ID: "T3", NodoA: "C", NodoB: "B",
				Tipo: mina.TipoDoble, Tiempo: 5,
			},
		},
	}

	config := fallas.Configuracion{
		ProbabilidadFalla:      1,
		ProbabilidadPermanente: 1,
		ReparacionMinTicks:     3,
		ReparacionMaxTicks:     8,
		Tick:                   time.Millisecond,
		Semilla:                123,
	}

	gestor, err := fallas.NuevoGestor(config)
	if err != nil {
		t.Fatal(err)
	}

	controlador, err := fallas.NuevoControlador(
		gestor,
		fallas.NuevaCirculacion(),
	)
	if err != nil {
		t.Fatal(err)
	}

	planificador := NuevoPlanificador(m)
	tramo := m.Tramos["T1"]

	// El camion C1 comienza a circular desde A hacia B.
	if err := controlador.IniciarTramo(tramo, "A->B", "C1"); err != nil {
		t.Fatal(err)
	}

	// Forzamos una falla permanente mediante las probabilidades.
	resultado, err := controlador.EvaluarCamion(
		tramo, "A->B", "C1",
	)
	if err != nil {
		t.Fatal(err)
	}

	if resultado.Tipo != fallas.FallaPermanente {
		t.Fatal("se esperaba una falla permanente")
	}

	// Obtenemos los bloqueos registrados por el gestor.
	bloqueos := gestor.SnapshotDireccional()

	if !bloqueos["T1"]["A->B"] {
		t.Fatal("el sentido A->B debe estar bloqueado")
	}

	if bloqueos["T1"]["B->A"] {
		t.Fatal("el sentido B->A debe seguir disponible")
	}

	// Una ruta nueva desde A debe evitar el tramo bloqueado.
	rutaIda, err := planificador.CalcularRutaDireccional(
		"A", "B", false, bloqueos,
	)
	if err != nil {
		t.Fatal(err)
	}

	esperadaIda := []string{"A", "C", "B"}

	if !rutasIgualesIntegracion(rutaIda, esperadaIda) {
		t.Fatalf(
			"ruta de ida incorrecta: %v, esperada: %v",
			rutaIda, esperadaIda,
		)
	}

	// El regreso puede utilizar el mismo tramo en sentido contrario.
	rutaVuelta, err := planificador.CalcularRutaDireccional(
		"B", "A", false, bloqueos,
	)
	if err != nil {
		t.Fatal(err)
	}

	esperadaVuelta := []string{"B", "A"}

	if !rutasIgualesIntegracion(rutaVuelta, esperadaVuelta) {
		t.Fatalf(
			"ruta de vuelta incorrecta: %v, esperada: %v",
			rutaVuelta, esperadaVuelta,
		)
	}
}

// Compara los nodos de dos rutas respetando su orden.
func rutasIgualesIntegracion(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
