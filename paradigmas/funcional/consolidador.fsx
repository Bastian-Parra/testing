open System
open System.IO
open System.Globalization

// Representa una fila de los logs del Certamen II.
type Evento = {
    Seq: int
    Tick: int
    Entidad: string
    TipoEvento: string
    Recurso: string
    Toneladas: float
    Origen: string
}


// ---------------------------------------------------------
// PARSEO DEL CSV
// ---------------------------------------------------------

// Convierte una linea CSV en un Evento.
// Esta funcion es pura: recibe texto y devuelve un resultado.
let parsearLinea origen numeroLinea (linea: string) =
    let partes = linea.Split(',')

    if partes.Length <> 6 then
        Error(
            sprintf
                "%s, linea %d: se esperaban 6 columnas y se encontraron %d"
                origen
                numeroLinea
                partes.Length
        )
    else
        let seqOk, seq =
            Int32.TryParse(partes.[0])

        let tickOk, tick =
            Int32.TryParse(partes.[1])

        let toneladasOk, toneladas =
            Double.TryParse(
                partes.[5],
                NumberStyles.Float,
                CultureInfo.InvariantCulture
            )

        if seqOk && tickOk && toneladasOk then
            Ok {
                Seq = seq
                Tick = tick
                Entidad = partes.[2]
                TipoEvento = partes.[3]
                Recurso = partes.[4]
                Toneladas = toneladas
                Origen = origen
            }
        else
            Error(
                sprintf
                    "%s, linea %d: hay valores numericos invalidos"
                    origen
                    numeroLinea
            )


// Convierte todas las lineas del archivo en una lista de eventos.
// La primera linea corresponde a la cabecera.
let parsearContenido origen lineas =
    match lineas with
    | [] ->
        Error [
            sprintf "%s: archivo vacio" origen
        ]

    | _cabecera :: datos ->

        let resultados =
            datos
            |> List.mapi (fun indice linea ->
                parsearLinea origen (indice + 2) linea
            )

        let errores =
            resultados
            |> List.choose (function
                | Error mensaje -> Some mensaje
                | Ok _ -> None
            )

        let eventos =
            resultados
            |> List.choose (function
                | Ok evento -> Some evento
                | Error _ -> None
            )

        if List.isEmpty errores then
            Ok eventos
        else
            Error errores


// ---------------------------------------------------------
// VALIDACION DEL ORDEN
// ---------------------------------------------------------

// Comprueba que los Seq de un archivo sean
// estrictamente crecientes.
//
// Usa recursion y pattern matching.
let rec validarOrden eventos =
    match eventos with
    | []
    | [_] ->
        []

    | actual :: siguiente :: resto ->

        let erroresRestantes =
            validarOrden (siguiente :: resto)

        if actual.Seq < siguiente.Seq then
            erroresRestantes
        else
            let error =
                sprintf
                    "%s: seq fuera de orden o repetido (%d -> %d)"
                    actual.Origen
                    actual.Seq
                    siguiente.Seq

            error :: erroresRestantes


// ---------------------------------------------------------
// MEZCLA DE LISTAS ORDENADAS
// ---------------------------------------------------------

// Mezcla dos listas ya ordenadas por Seq.
// No modifica ninguna de las listas originales.
let rec mezclarDos listaA listaB =
    match listaA, listaB with

    // Si A esta vacia, queda B.
    | [], resto ->
        resto, []

    // Si B esta vacia, queda A.
    | resto, [] ->
        resto, []

    // Ambas listas tienen elementos.
    | eventoA :: restoA, eventoB :: restoB ->

        if eventoA.Seq < eventoB.Seq then

            let mezclados, errores =
                mezclarDos restoA listaB

            eventoA :: mezclados, errores

        elif eventoB.Seq < eventoA.Seq then

            let mezclados, errores =
                mezclarDos listaA restoB

            eventoB :: mezclados, errores

        else

            // Si ambos tienen el mismo Seq,
            // existe un Seq global duplicado.
            let mezclados, errores =
                mezclarDos restoA restoB

            let error =
                sprintf
                    "Seq duplicado %d entre %s y %s"
                    eventoA.Seq
                    eventoA.Origen
                    eventoB.Origen

            eventoA :: eventoB :: mezclados,
            error :: errores


// Generaliza la mezcla para cualquier cantidad de listas.
// Cada lista representa un archivo de log.
let rec mezclarK listas =
    match listas with

    | [] ->
        [], []

    | [lista] ->
        lista, []

    | primera :: resto ->

        let mezclaResto, erroresResto =
            mezclarK resto

        let mezclaFinal, erroresMezcla =
            mezclarDos primera mezclaResto

        mezclaFinal,
        erroresResto @ erroresMezcla


// ---------------------------------------------------------
// CONVERSION A CSV
// ---------------------------------------------------------

// Convierte un Evento nuevamente a una linea CSV.
// Sigue siendo una funcion pura.
let eventoACsv evento =
    sprintf
        "%d,%d,%s,%s,%s,%s"
        evento.Seq
        evento.Tick
        evento.Entidad
        evento.TipoEvento
        evento.Recurso
        (evento.Toneladas.ToString(CultureInfo.InvariantCulture))


// ---------------------------------------------------------
// ENTRADA / SALIDA
// ---------------------------------------------------------

// Lee fisicamente un archivo.
// La E/S queda separada de la logica pura.
let leerArchivo ruta =
    let lineas =
        File.ReadAllLines(ruta)
        |> Array.toList

    parsearContenido
        (Path.GetFileName(ruta))
        lineas


// Lee un archivo y ademas valida
// que sus Seq esten ordenados.
let cargarYValidar ruta =
    match leerArchivo ruta with

    | Error errores ->
        Error errores

    | Ok eventos ->

        let erroresOrden =
            validarOrden eventos

        if List.isEmpty erroresOrden then
            Ok eventos
        else
            Error erroresOrden


// Guarda el log global consolidado.
// Esta funcion SI realiza E/S.
let guardarConsolidado ruta eventos =
    let cabecera =
        "Seq,Tick,Entidad,Evento,Recurso,Toneladas"

    let lineasEventos =
        eventos
        |> List.map eventoACsv

    let contenido =
        cabecera :: lineasEventos

    File.WriteAllLines(ruta, contenido)


// ---------------------------------------------------------
// PROGRAMA PRINCIPAL
// ---------------------------------------------------------

let argumentos =
    fsi.CommandLineArgs
    |> Array.skip 1
    |> Array.toList


match argumentos with

| [] ->
    printfn "Uso:"
    printfn "dotnet fsi paradigmas/funcional/consolidador.fsx <log1.csv> <log2.csv> ..."

| rutas ->

    // map aplica cargarYValidar a cada archivo recibido.
    let resultados =
        rutas
        |> List.map cargarYValidar


    // Extraemos todos los errores encontrados.
    let erroresLectura =
        resultados
        |> List.collect (function
            | Error errores -> errores
            | Ok _ -> []
        )


    // Extraemos solamente las listas validas de eventos.
    let listasEventos =
        resultados
        |> List.choose (function
            | Ok eventos -> Some eventos
            | Error _ -> None
        )


    if not (List.isEmpty erroresLectura) then

        printfn "Errores encontrados:"

        erroresLectura
        |> List.iter (printfn "  - %s")

    else

        // Mezclamos todas las listas ordenadas.
        let consolidado, erroresMezcla =
            mezclarK listasEventos


        if List.isEmpty erroresMezcla then

            printfn "Consolidacion correcta."
            printfn "Eventos totales: %d" (List.length consolidado)

            printfn "Secuencias consolidadas:"

            consolidado
            |> List.map (fun evento -> evento.Seq)
            |> List.iter (printf "%d ")

            printfn ""


            // Usamos como directorio de salida la carpeta
            // donde se encuentra el primer log.
            let primerArchivo =
                List.head rutas

            let directorioOriginal =
                Path.GetDirectoryName(primerArchivo)

            let directorioSalida =
                if String.IsNullOrWhiteSpace(directorioOriginal) then
                    "."
                else
                    directorioOriginal

            let rutaSalida =
                Path.Combine(
                    directorioSalida,
                    "log_consolidado.csv"
                )


            guardarConsolidado rutaSalida consolidado

            printfn "Log consolidado guardado en:"
            printfn "%s" rutaSalida

        else

            printfn "No se puede consolidar correctamente."
            printfn "Se encontraron Seq duplicados:"

            erroresMezcla
            |> List.iter (printfn "  - %s")