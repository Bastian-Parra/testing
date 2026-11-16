open System
open System.IO
open System.Globalization


// ---------------------------------------------------------
// TIPOS
// ---------------------------------------------------------

// Representa una fila del log consolidado.
type Evento = {
    Seq: int
    Tick: int
    Entidad: string
    TipoEvento: string
    Recurso: string
    Toneladas: float
}


// Resultado de las metricas calculadas.
type Metricas = {
    ToneladasTransportadas: float
    UtilizacionPalas: (string * float) list
    TiempoPromedioColaPalas: float option
    TiempoPromedioColaStockpiles: float option
    CamionesAtrapados: string list
}


// ---------------------------------------------------------
// FUNCIONES AUXILIARES
// ---------------------------------------------------------

// Normaliza los nombres de eventos para facilitar comparaciones.
let normalizar (texto: string) =
    texto
        .Trim()
        .ToLowerInvariant()
        .Replace(" ", "_")


// Comprueba si un texto comienza con cierto prefijo.
let empiezaCon (prefijo: string) (texto: string) =
    texto.StartsWith(
        prefijo,
        StringComparison.OrdinalIgnoreCase
    )


// Comprueba si el tipo de evento pertenece a una lista.
let esUnoDe nombres evento =
    nombres
    |> List.contains (normalizar evento.TipoEvento)


// ---------------------------------------------------------
// PARSEO DEL CSV
// ---------------------------------------------------------

let parsearLinea numeroLinea (linea: string) =
    let partes =
        linea.Split(',')

    if partes.Length <> 6 then

        Error(
            sprintf
                "Linea %d: se esperaban 6 columnas y se encontraron %d"
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
            }

        else

            Error(
                sprintf
                    "Linea %d: existen valores numericos invalidos"
                    numeroLinea
            )


let parsearContenido lineas =
    match lineas with

    | [] ->
        Error [ "El archivo esta vacio." ]

    | _cabecera :: datos ->

        let resultados =
            datos
            |> List.filter (fun linea ->
                not (String.IsNullOrWhiteSpace(linea))
            )
            |> List.mapi (fun indice linea ->
                parsearLinea (indice + 2) linea
            )

        let errores =
            resultados
            |> List.choose (function
                | Error mensaje ->
                    Some mensaje

                | Ok _ ->
                    None
            )

        let eventos =
            resultados
            |> List.choose (function
                | Ok evento ->
                    Some evento

                | Error _ ->
                    None
            )

        if List.isEmpty errores then
            Ok eventos
        else
            Error errores


// ---------------------------------------------------------
// METRICA 1
// TONELADAS TRANSPORTADAS
// ---------------------------------------------------------

// Se consideran toneladas efectivamente transportadas
// cuando el camion completa una descarga en un stockpile.
let toneladasTransportadas eventos =

    let descargas =
        eventos
        |> List.filter (fun evento ->

            empiezaCon "C-" evento.Entidad
            &&
            empiezaCon "Stockpile" evento.Recurso
            &&
            evento.Toneladas > 0.0
            &&
            esUnoDe
                [
                    "descarga"
                    "fin_descarga"
                    "descarga_completa"
                ]
                evento
        )

    // En algunos logs una descarga terminada puede
    // aparecer como liberacion del stockpile.
    let eventosValidos =
        if not (List.isEmpty descargas) then

            descargas

        else

            eventos
            |> List.filter (fun evento ->

                empiezaCon "C-" evento.Entidad
                &&
                empiezaCon "Stockpile" evento.Recurso
                &&
                evento.Toneladas > 0.0
                &&
                normalizar evento.TipoEvento = "libera"
            )

    eventosValidos
    |> List.fold
        (fun acumulado evento ->
            acumulado + evento.Toneladas
        )
        0.0


// ---------------------------------------------------------
// FUNCIONES DE INTERVALOS
// ---------------------------------------------------------

// Busca el primer evento posterior que cumple una condicion.
let buscarCierre eventos inicio condicion =

    eventos
    |> List.filter (fun evento ->
        evento.Seq > inicio.Seq
        &&
        condicion evento
    )
    |> List.sortBy (fun evento ->
        evento.Seq
    )
    |> List.tryHead


// Duracion total observada en el log.
let duracionTurno eventos =
    match eventos with

    | [] ->
        0

    | _ ->

        let tickInicial =
            eventos
            |> List.map (fun evento ->
                evento.Tick
            )
            |> List.min

        let tickFinal =
            eventos
            |> List.map (fun evento ->
                evento.Tick
            )
            |> List.max

        max 0 (tickFinal - tickInicial)


// ---------------------------------------------------------
// METRICA 2
// UTILIZACION DE CADA PALA
// ---------------------------------------------------------

// Calcula cuanto tiempo una pala estuvo atendiendo camiones.
let duracionUsoPala pala eventos =

    let adquisiciones =
        eventos
        |> List.filter (fun evento ->

            evento.Entidad = pala
            &&
            normalizar evento.TipoEvento = "adquiere"
        )

    adquisiciones
    |> List.choose (fun inicio ->

        buscarCierre
            eventos
            inicio
            (fun cierre ->

                cierre.Entidad = pala
                &&
                cierre.Recurso = inicio.Recurso
                &&
                normalizar cierre.TipoEvento = "libera"
            )

        |> Option.map (fun cierre ->
            max 0 (cierre.Tick - inicio.Tick)
        )
    )
    |> List.sum


let utilizacionPalas eventos =

    let duracion =
        duracionTurno eventos

    let palas =
        eventos
        |> List.map (fun evento ->
            evento.Entidad
        )
        |> List.filter (empiezaCon "Pala")
        |> List.distinct

    palas
    |> List.map (fun pala ->

        let tiempoUtilizado =
            duracionUsoPala pala eventos

        let porcentaje =
            if duracion = 0 then
                0.0
            else
                (
                    float tiempoUtilizado
                    /
                    float duracion
                )
                * 100.0

        pala, porcentaje
    )


// ---------------------------------------------------------
// METRICAS 3 Y 4
// TIEMPO PROMEDIO EN COLAS
// ---------------------------------------------------------

type TipoCola =
    | ColaPala
    | ColaStockpile


let eventosEntradaCola tipo =
    match tipo with

    | ColaPala ->
        [
            "entra_cola_pala"
            "entrada_cola_pala"
            "espera_pala"
        ]

    | ColaStockpile ->
        [
            "entra_cola_stockpile"
            "entrada_cola_stockpile"
            "espera_stockpile"
        ]


let eventosSalidaCola tipo =
    match tipo with

    | ColaPala ->
        [
            "sale_cola_pala"
            "salida_cola_pala"
            "inicio_carga"
        ]

    | ColaStockpile ->
        [
            "sale_cola_stockpile"
            "salida_cola_stockpile"
            "inicio_descarga"
        ]


// Obtiene todas las duraciones de espera encontradas.
let duracionesCola tipo eventos =

    let nombresEntrada =
        eventosEntradaCola tipo

    let nombresSalida =
        eventosSalidaCola tipo

    let entradas =
        eventos
        |> List.filter (fun evento ->

            empiezaCon "C-" evento.Entidad
            &&
            esUnoDe nombresEntrada evento
        )

    entradas
    |> List.choose (fun inicio ->

        buscarCierre
            eventos
            inicio
            (fun cierre ->

                cierre.Entidad = inicio.Entidad
                &&
                cierre.Recurso = inicio.Recurso
                &&
                esUnoDe nombresSalida cierre
            )

        |> Option.map (fun cierre ->
            max 0 (cierre.Tick - inicio.Tick)
        )
    )


let promedio valores =
    match valores with

    | [] ->
        None

    | _ ->
        valores
        |> List.averageBy float
        |> Some


let tiempoPromedioCola tipo eventos =
    duracionesCola tipo eventos
    |> promedio


// ---------------------------------------------------------
// METRICA 5
// CAMIONES ATRAPADOS POR FALLAS PERMANENTES
// ---------------------------------------------------------

let camionesAtrapados eventos =

    let fallasPermanentes =
        eventos
        |> List.filter (fun evento ->

            empiezaCon "C-" evento.Entidad
            &&
            esUnoDe
                [
                    "falla_permanente"
                    "falla-permanente"
                ]
                evento
        )

    let esperas =
        eventos
        |> List.filter (fun evento ->

            empiezaCon "C-" evento.Entidad
            &&
            esUnoDe
                [
                    "espera"
                    "espera_prolongada"
                ]
                evento
        )

    fallasPermanentes
    |> List.collect (fun falla ->

        esperas
        |> List.choose (fun espera ->

            if
                espera.Seq > falla.Seq
                &&
                espera.Entidad <> falla.Entidad
                &&
                espera.Recurso = falla.Recurso
            then
                Some espera.Entidad
            else
                None
        )
    )
    |> List.distinct


// ---------------------------------------------------------
// CALCULO DE TODAS LAS METRICAS
// ---------------------------------------------------------

// Esta funcion es pura.
// Recibe eventos y devuelve las metricas.
let calcularMetricas eventos =

    {
        ToneladasTransportadas =
            toneladasTransportadas eventos

        UtilizacionPalas =
            utilizacionPalas eventos

        TiempoPromedioColaPalas =
            tiempoPromedioCola ColaPala eventos

        TiempoPromedioColaStockpiles =
            tiempoPromedioCola ColaStockpile eventos

        CamionesAtrapados =
            camionesAtrapados eventos
    }


// ---------------------------------------------------------
// ENTRADA / SALIDA
// ---------------------------------------------------------

let leerArchivo ruta =

    File.ReadAllLines(ruta)
    |> Array.toList
    |> parsearContenido


let mostrarPromedio nombre resultado =

    match resultado with

    | Some valor ->
        printfn "%s: %.2f ticks" nombre valor

    | None ->
        printfn "%s: sin datos suficientes" nombre


let mostrarMetricas metricas =

    printfn ""
    printfn "===== METRICAS ====="

    printfn
        "Toneladas transportadas: %.2f"
        metricas.ToneladasTransportadas

    printfn ""
    printfn "Utilizacion de palas:"

    match metricas.UtilizacionPalas with

    | [] ->
        printfn "  Sin datos de palas."

    | palas ->

        palas
        |> List.iter (fun (pala, porcentaje) ->

            printfn
                "  %s: %.2f %%"
                pala
                porcentaje
        )

    printfn ""

    mostrarPromedio
        "Tiempo promedio cola en palas"
        metricas.TiempoPromedioColaPalas

    mostrarPromedio
        "Tiempo promedio cola en stockpiles"
        metricas.TiempoPromedioColaStockpiles

    printfn ""

    printfn
        "Camiones atrapados por fallas permanentes: %d"
        (List.length metricas.CamionesAtrapados)

    metricas.CamionesAtrapados
    |> List.iter (fun camion ->
        printfn "  - %s" camion
    )


// ---------------------------------------------------------
// PROGRAMA PRINCIPAL
// ---------------------------------------------------------

let argumentos =
    fsi.CommandLineArgs
    |> Array.skip 1
    |> Array.toList


match argumentos with

| [ruta] ->

    match leerArchivo ruta with

    | Error errores ->

        printfn "Errores encontrados:"

        errores
        |> List.iter (fun error ->
            printfn "  - %s" error
        )

    | Ok eventos ->

        let metricas =
            calcularMetricas eventos

        mostrarMetricas metricas


| _ ->

    printfn "Uso:"
    printfn "dotnet fsi paradigmas/funcional/metricas.fsx <log_consolidado.csv>"