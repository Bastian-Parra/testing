package main

import (
	"context"
	"fmt"
	"sync"
	"time"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/camion"
	"github.com/Bastian-Parra/ici425-carabelli-flores-parra/simulador/internal/mina"
)

func main() {
	fmt.Println("Iniciando Simulación: Escenario sin_fallas")
	ctx, cancelar := context.WithCancel(context.Background())
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