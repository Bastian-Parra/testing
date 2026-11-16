% =========================================================
% INVARIANTES 4, 5 Y 6 - CERTAMEN III
% =========================================================
%
% 4. Una pala o stockpile atiende a un solo camion a la vez.
% 5. Un camion en falla no se mueve hasta su reparacion.
%    Si la falla es permanente, no se mueve nuevamente.
% 6. Las cargas registradas por camion y pala deben coincidir.
%
% Los hechos son generados desde los logs por la Parte 1.
% =========================================================


:- use_module(library(lists)).


% ---------------------------------------------------------
% PREDICADOS QUE PUEDEN VENIR DESDE LOS HECHOS
% ---------------------------------------------------------

:- multifile adquiere/4.
:- multifile libera/4.
:- multifile carga/6.
:- multifile falla/5.
:- multifile repara/4.
:- multifile viaja/4.


% ---------------------------------------------------------
% FUNCIONES AUXILIARES
% ---------------------------------------------------------

% Determina si un recurso corresponde a una pala o stockpile.
recurso_atencion(Recurso) :-
    atom(Recurso),
    (
        sub_atom(Recurso, 0, 4, _, 'Pala')
        ;
        sub_atom(Recurso, 0, 9, _, 'Stockpile')
    ).


% Busca la primera liberacion posterior a una adquisicion.
liberacion_posterior(
    Camion,
    Recurso,
    SeqAdquiere,
    SeqLibera
) :-
    findall(
        Seq,
        (
            libera(Seq, _, Camion, Recurso),
            Seq > SeqAdquiere
        ),
        Liberaciones
    ),

    Liberaciones \= [],
    min_list(Liberaciones, SeqLibera).


% Consideramos movimiento una adquisicion,
% una liberacion o un evento explicito de viaje.
movimiento(
    Seq,
    Tick,
    Camion,
    Recurso
) :-
    adquiere(
        Seq,
        Tick,
        Camion,
        Recurso
    ).


movimiento(
    Seq,
    Tick,
    Camion,
    Recurso
) :-
    libera(
        Seq,
        Tick,
        Camion,
        Recurso
    ).


movimiento(
    Seq,
    Tick,
    Camion,
    Recurso
) :-
    viaja(
        Seq,
        Tick,
        Camion,
        Recurso
    ).


% =========================================================
% INVARIANTE 4
% UNA PALA O STOCKPILE NO PUEDE ATENDER
% DOS CAMIONES AL MISMO TIEMPO
% =========================================================

violacion_atencion_exclusiva(
    Recurso,
    Camion1,
    Camion2,
    SeqInicio1,
    SeqInicio2
) :-

    % Primero obtenemos un recurso desde un evento real.
    adquiere(
        SeqInicio1,
        _,
        Camion1,
        Recurso
    ),

    % Ahora que Recurso ya tiene un valor,
    % comprobamos que sea una pala o stockpile.
    recurso_atencion(Recurso),

    % Buscamos cuando Camion1 libera ese recurso.
    liberacion_posterior(
        Camion1,
        Recurso,
        SeqInicio1,
        SeqFin1
    ),

    % Buscamos otro camion que adquiere el mismo recurso.
    adquiere(
        SeqInicio2,
        _,
        Camion2,
        Recurso
    ),

    Camion1 \= Camion2,

    % El segundo camion entra despues del primero,
    % pero antes de que el primero libere el recurso.
    SeqInicio1 < SeqInicio2,
    SeqInicio2 < SeqFin1.


% =========================================================
% INVARIANTE 5
% FALLAS TEMPORALES Y PERMANENTES
% =========================================================


% ---------------------------------------------------------
% FALLA TEMPORAL
%
% Si un camion falla temporalmente, no puede registrar
% movimiento antes de una reparacion.
% ---------------------------------------------------------

violacion_falla_temporal(
    Camion,
    SeqFalla,
    SeqMovimiento
) :-

    falla(
        SeqFalla,
        _,
        Camion,
        _,
        temporal
    ),

    movimiento(
        SeqMovimiento,
        _,
        Camion,
        _
    ),

    SeqMovimiento > SeqFalla,

    % No debe existir reparacion entre la falla y movimiento.
    \+ (
        repara(
            SeqReparacion,
            _,
            Camion,
            _
        ),

        SeqReparacion > SeqFalla,
        SeqReparacion < SeqMovimiento
    ).


% ---------------------------------------------------------
% FALLA PERMANENTE
%
% Despues de una falla permanente no puede existir
% ningun movimiento posterior.
% ---------------------------------------------------------

violacion_falla_permanente(
    Camion,
    SeqFalla,
    SeqMovimiento
) :-

    falla(
        SeqFalla,
        _,
        Camion,
        _,
        permanente
    ),

    movimiento(
        SeqMovimiento,
        _,
        Camion,
        _
    ),

    SeqMovimiento > SeqFalla.


% =========================================================
% INVARIANTE 6
% CONSISTENCIA ENTRE LOG DEL CAMION Y LOG DE LA PALA
% =========================================================


% Una carga registrada desde el punto de vista del camion.
carga_camion(
    Seq,
    Camion,
    Pala,
    Toneladas
) :-

    carga(
        Entidad,
        Seq,
        _,
        Camion,
        Pala,
        Toneladas
    ),

    Entidad == Camion.


% Una carga registrada desde el punto de vista de la pala.
carga_pala(
    Seq,
    Camion,
    Pala,
    Toneladas
) :-

    carga(
        Entidad,
        Seq,
        _,
        Camion,
        Pala,
        Toneladas
    ),

    Entidad == Pala.


% El camion tiene una carga que no aparece en la pala.
violacion_carga_solo_camion(
    Camion,
    Pala,
    Toneladas,
    SeqCamion
) :-

    carga_camion(
        SeqCamion,
        Camion,
        Pala,
        Toneladas
    ),

    \+ carga_pala(
        _,
        Camion,
        Pala,
        Toneladas
    ).


% La pala tiene una carga que no aparece en el camion.
violacion_carga_solo_pala(
    Camion,
    Pala,
    Toneladas,
    SeqPala
) :-

    carga_pala(
        SeqPala,
        Camion,
        Pala,
        Toneladas
    ),

    \+ carga_camion(
        _,
        Camion,
        Pala,
        Toneladas
    ).


% =========================================================
% REPORTE GENERAL
% =========================================================

violacion(
    atencion_exclusiva(
        Recurso,
        Camion1,
        Camion2,
        Seq1,
        Seq2
    )
) :-

    violacion_atencion_exclusiva(
        Recurso,
        Camion1,
        Camion2,
        Seq1,
        Seq2
    ).


violacion(
    falla_temporal(
        Camion,
        SeqFalla,
        SeqMovimiento
    )
) :-

    violacion_falla_temporal(
        Camion,
        SeqFalla,
        SeqMovimiento
    ).


violacion(
    falla_permanente(
        Camion,
        SeqFalla,
        SeqMovimiento
    )
) :-

    violacion_falla_permanente(
        Camion,
        SeqFalla,
        SeqMovimiento
    ).


violacion(
    inconsistencia_carga_camion(
        Camion,
        Pala,
        Toneladas,
        Seq
    )
) :-

    violacion_carga_solo_camion(
        Camion,
        Pala,
        Toneladas,
        Seq
    ).


violacion(
    inconsistencia_carga_pala(
        Camion,
        Pala,
        Toneladas,
        Seq
    )
) :-

    violacion_carga_solo_pala(
        Camion,
        Pala,
        Toneladas,
        Seq
    ).


% Muestra todas las violaciones encontradas.
mostrar_violaciones :-

    findall(
        Violacion,
        violacion(Violacion),
        ListaOriginal
    ),

    sort(
        ListaOriginal,
        Violaciones
    ),

    (
        Violaciones = []
        ->
        writeln('No se encontraron violaciones.')

        ;

        writeln('Violaciones encontradas:'),

        forall(
            member(Violacion, Violaciones),
            writeln(Violacion)
        )
    ).