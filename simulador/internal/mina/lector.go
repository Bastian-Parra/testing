package mina

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// lineaEntrada guarda una línea útil del archivo junto a su número.
// El número de línea sirve para entregar errores más fáciles de entender.
type lineaEntrada struct {
	numero int
	campos []string
}

// LeerArchivo abre un archivo de mina desde el disco
// y utiliza Leer para interpretar su contenido.
func LeerArchivo(ruta string) (*Mina, error) {
	archivo, err := os.Open(ruta)
	if err != nil {
		return nil, fmt.Errorf(
			"no se pudo abrir el archivo %q: %w",
			ruta,
			err,
		)
	}
	defer archivo.Close()

	return Leer(archivo)
}

// Leer recibe cualquier fuente que implemente io.Reader.
//
// Esto nos permite usar archivos reales durante la simulación
// y strings durante las pruebas.
func Leer(origen io.Reader) (*Mina, error) {
	lineas, err := leerLineas(origen)
	if err != nil {
		return nil, err
	}

	mina := NuevaMina()

	// Primera pasada:
	// primero registramos todos los nodos.
	//
	// Esto permite que los tramos puedan aparecer incluso antes
	// que los nodos en el archivo sin generar problemas.
	for _, linea := range lineas {
		if linea.campos[0] != "NODO" {
			continue
		}

		nodo, err := interpretarNodo(linea)
		if err != nil {
			return nil, err
		}

		if err := mina.Tabla.RegistrarNodo(nodo); err != nil {
			return nil, fmt.Errorf(
				"línea %d: %w",
				linea.numero,
				err,
			)
		}

		mina.Nodos[nodo.ID] = nodo
	}

	// Segunda pasada:
	// ahora que conocemos todos los nodos podemos comprobar
	// correctamente los extremos de cada tramo.
	for _, linea := range lineas {
		if linea.campos[0] != "TRAMO" {
			continue
		}

		tramo, err := interpretarTramo(linea, mina)
		if err != nil {
			return nil, err
		}

		if err := mina.Tabla.RegistrarTramo(tramo); err != nil {
			return nil, fmt.Errorf(
				"línea %d: %w",
				linea.numero,
				err,
			)
		}

		mina.Tramos[tramo.ID] = tramo
	}

	return mina, nil
}

// leerLineas lee y prepara las declaraciones del archivo.
//
// Ignora:
//   - líneas vacías;
//   - comentarios comenzados con #.
//
// También verifica que cada declaración comience con
// NODO o TRAMO.
func leerLineas(origen io.Reader) ([]lineaEntrada, error) {
	scanner := bufio.NewScanner(origen)

	lineas := make([]lineaEntrada, 0)
	numeroLinea := 0

	for scanner.Scan() {
		numeroLinea++

		texto := strings.TrimSpace(scanner.Text())

		// Si existe un comentario, conservamos solamente
		// el texto que aparece antes del símbolo #.
		if posicionComentario := strings.Index(texto, "#"); posicionComentario >= 0 {
			texto = strings.TrimSpace(texto[:posicionComentario])
		}

		if texto == "" {
			continue
		}

		campos := strings.Fields(texto)

		if campos[0] != "NODO" && campos[0] != "TRAMO" {
			return nil, fmt.Errorf(
				"línea %d: declaración desconocida %q",
				numeroLinea,
				campos[0],
			)
		}

		lineas = append(lineas, lineaEntrada{
			numero: numeroLinea,
			campos: campos,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf(
			"error leyendo el archivo de la mina: %w",
			err,
		)
	}

	return lineas, nil
}

// interpretarNodo convierte una línea NODO en una estructura Nodo.
//
// Ejemplos:
//
//	NODO E1 estacionamiento
//	NODO I1 interseccion
//	NODO P1 pala 3
//	NODO S1 stockpile
func interpretarNodo(linea lineaEntrada) (*Nodo, error) {
	campos := linea.campos

	if len(campos) < 3 || len(campos) > 4 {
		return nil, fmt.Errorf(
			"línea %d: formato NODO inválido",
			linea.numero,
		)
	}

	id := campos[1]
	tipo := TipoNodo(campos[2])

	if !EsTipoNodoValido(tipo) {
		return nil, fmt.Errorf(
			"línea %d: tipo de nodo %q no válido",
			linea.numero,
			tipo,
		)
	}

	capacidad := 0

	// La capacidad solamente se utiliza para una pala.
	if tipo == TipoPala {
		if len(campos) != 4 {
			return nil, fmt.Errorf(
				"línea %d: la pala %q debe indicar capacidad",
				linea.numero,
				id,
			)
		}

		valor, err := strconv.Atoi(campos[3])
		if err != nil || valor <= 0 {
			return nil, fmt.Errorf(
				"línea %d: capacidad inválida para la pala %q",
				linea.numero,
				id,
			)
		}

		capacidad = valor
	} else if len(campos) != 3 {
		return nil, fmt.Errorf(
			"línea %d: el nodo %q de tipo %q no utiliza capacidad",
			linea.numero,
			id,
			tipo,
		)
	}

	return &Nodo{
		ID:        id,
		Tipo:      tipo,
		Capacidad: capacidad,
	}, nil
}

// interpretarTramo convierte una línea TRAMO en una estructura Tramo.
//
// Además comprueba que ambos nodos extremos existan.
func interpretarTramo(
	linea lineaEntrada,
	mina *Mina,
) (*Tramo, error) {
	campos := linea.campos

	if len(campos) != 6 {
		return nil, fmt.Errorf(
			"línea %d: formato TRAMO inválido",
			linea.numero,
		)
	}

	id := campos[1]
	nodoA := campos[2]
	nodoB := campos[3]
	tipo := TipoTramo(campos[4])

	if _, existe := mina.Nodos[nodoA]; !existe {
		return nil, fmt.Errorf(
			"línea %d: nodo %q no declarado",
			linea.numero,
			nodoA,
		)
	}

	if _, existe := mina.Nodos[nodoB]; !existe {
		return nil, fmt.Errorf(
			"línea %d: nodo %q no declarado",
			linea.numero,
			nodoB,
		)
	}

	if !EsTipoTramoValido(tipo) {
		return nil, fmt.Errorf(
			"línea %d: tipo de tramo %q no válido",
			linea.numero,
			tipo,
		)
	}

	tiempo, err := strconv.Atoi(campos[5])
	if err != nil || tiempo <= 0 {
		return nil, fmt.Errorf(
			"línea %d: tiempo inválido para el tramo %q",
			linea.numero,
			id,
		)
	}

	return &Tramo{
		ID:     id,
		NodoA:  nodoA,
		NodoB:  nodoB,
		Tipo:   tipo,
		Tiempo: tiempo,
	}, nil
}
