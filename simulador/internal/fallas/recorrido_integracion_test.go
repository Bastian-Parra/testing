package fallas

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestPanicSeConvierteEnFallaPermanente(t *testing.T) {
	r := prepararRecorrido(t, 0, 0)
	r.PanicEnTick = func(id, tramo string, avance int) bool { return avance == 1 }
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := r.Recorrer(ctx, "C1", []string{"A", "B"})
	if err == nil || !strings.Contains(err.Error(), "panic recuperado") {
		t.Fatalf("panic sin recuperar: %v", err)
	}
	camion, e := r.Controlador.Circulacion.ConsultarCamion("T1", "A->B", "C1")
	if e != nil || !camion.FallaPermanente || camion.Avance != 1 {
		t.Fatalf("posicion de panic perdida: %+v, %v", camion, e)
	}
	if !r.Controlador.Gestor.SnapshotDireccional()["T1"]["A->B"] {
		t.Fatal("panic no bloqueo sentido")
	}
	if e := r.Controlador.IniciarTramo(r.Mina.Tramos["T1"], "A->B", "C2"); e == nil {
		t.Fatal("permitio ingresar al sentido bloqueado")
	}
	if e := r.Controlador.IniciarTramo(r.Mina.Tramos["T1"], "B->A", "C3"); e != nil {
		t.Fatalf("bloqueo indebidamente sentido contrario: %v", e)
	}
}
func TestCancelacionLimpiaCamionSano(t *testing.T) {
	r := prepararRecorrido(t, 0, 0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := r.Recorrer(ctx, "C1", []string{"A", "B"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("se esperaba cancelacion: %v", err)
	}
	if cola := r.Controlador.Circulacion.SnapshotOrden("T1", "A->B"); len(cola) != 0 {
		t.Fatalf("camion sano quedo registrado: %v", cola)
	}
}
func TestFIFOImpideAdelantarPosicion(t *testing.T) {
	c := NuevaCirculacion()
	if err := c.Entrar("T1", "A->B", "C1"); err != nil {
		t.Fatal(err)
	}
	if err := c.Entrar("T1", "A->B", "C2"); err != nil {
		t.Fatal(err)
	}
	if c.PuedeAvanzar("T1", "A->B", "C2") {
		t.Fatal("segundo camion adelanto al primero")
	}
	if err := c.AvanzarTick("T1", "A->B", "C1"); err != nil {
		t.Fatal(err)
	}
	if !c.PuedeAvanzar("T1", "A->B", "C2") {
		t.Fatal("segundo camion no pudo avanzar tras el primero")
	}
	if err := c.AvanzarTick("T1", "A->B", "C2"); err != nil {
		t.Fatal(err)
	}
	if err := c.RegistrarFalla("T1", "A->B", "C1", true, 0); err != nil {
		t.Fatal(err)
	}
	if c.PuedeAvanzar("T1", "A->B", "C2") {
		t.Fatal("camion adelanto a camion averiado")
	}
}
