% =========================================================
% CONSULTAS - CERTAMEN III
% =========================================================
%
% Consulta 1:
% Camiones afectados por una falla permanente.
%
% Consulta 2:
% Camiones y recursos involucrados en un deadlock.
%
% Consulta 3:
% Toneladas cargadas por cada pala.
% =========================================================


:- use_module(library(lists)).

:- multifile adquiere/4.
:- multifile libera/4.
:- multifile carga/6.
:- multifile falla/5.
:- multifile espera/5.


% =========================================================
% CONSULTA 1
% CAMIONES ATRAPADOS POR UNA FALLA PERMANENTE
% =========================================================

camion_atrapado(
    CamionAtrapado,
    CamionEnFalla,
    Recurso,
    Tick
) :-

    falla(
        _,
        TickFalla,
        CamionEnFalla,
        Recurso,
        permanente
    ),

    espera(
        _,
        Tick,
        CamionAtrapado,
        Recurso,
        _
    ),

    CamionAtrapado \= CamionEnFalla,
    Tick >= TickFalla.


% =========================================================
% AUXILIAR PARA CONSULTA DE DEADLOCK
% =========================================================

% Un camion ocupa un recurso en cierto seq si
% lo adquirio previamente y todavia no lo libero.
ocupa_recurso(
    Camion,
    Recurso,
    SeqConsulta
) :-

    adquiere(
        SeqAdquiere,
        _,
        Camion,
        Recurso
    ),

    SeqAdquiere < SeqConsulta,

    \+ (
        libera(
            SeqLibera,
            _,
            Camion,
            Recurso
        ),

        SeqLibera > SeqAdquiere,
        SeqLibera < SeqConsulta
    ).


% =========================================================
% CONSULTA 2
% DEADLOCK ENTRE DOS CAMIONES
% =========================================================

deadlock(
    Camion1,
    Camion2,
    Recurso1,
    Recurso2,
    Tick
) :-

    espera(
        SeqEspera1,
        Tick1,
        Camion1,
        Recurso2,
        _
    ),

    espera(
        SeqEspera2,
        Tick2,
        Camion2,
        Recurso1,
        _
    ),

    % Evita reportar dos veces el mismo deadlock:
    % C-05/C-06 y luego C-06/C-05.
    Camion1 @< Camion2,

    ocupa_recurso(
        Camion1,
        Recurso1,
        SeqEspera1
    ),

    ocupa_recurso(
        Camion2,
        Recurso2,
        SeqEspera2
    ),

    Tick is max(Tick1, Tick2).


% =========================================================
% CONSULTA 3
% TONELADAS CARGADAS POR PALA
% =========================================================

toneladas_por_pala(
    Pala,
    Total
) :-

    % Primero obtenemos todas las palas que tienen
    % al menos una carga registrada por ellas mismas.
    setof(
        P,
        Seq^Tick^Camion^Toneladas^
        carga(
            P,
            Seq,
            Tick,
            Camion,
            P,
            Toneladas
        ),
        Palas
    ),

    % Entregamos una pala a la vez.
    member(
        Pala,
        Palas
    ),

    % Buscamos todas las toneladas registradas
    % especificamente por esa pala.
    findall(
        Toneladas,
        carga(
            Pala,
            _,
            _,
            _,
            Pala,
            Toneladas
        ),
        Cargas
    ),

    sum_list(
        Cargas,
        Total
    ).


% =========================================================
% REPORTE DE CONSULTAS
% =========================================================

mostrar_consultas :-

    writeln('--- Camiones atrapados por falla permanente ---'),

    forall(
        camion_atrapado(
            Camion,
            CamionEnFalla,
            Recurso,
            Tick
        ),

        format(
            '~w atrapado por ~w en ~w desde tick ~w~n',
            [
                Camion,
                CamionEnFalla,
                Recurso,
                Tick
            ]
        )
    ),

    nl,

    writeln('--- Deadlocks encontrados ---'),

    forall(
        deadlock(
            Camion1,
            Camion2,
            Recurso1,
            Recurso2,
            Tick
        ),

        format(
            '~w y ~w involucrados entre ~w y ~w desde tick ~w~n',
            [
                Camion1,
                Camion2,
                Recurso1,
                Recurso2,
                Tick
            ]
        )
    ),

    nl,

    writeln('--- Toneladas cargadas por pala ---'),

    forall(
        toneladas_por_pala(
            Pala,
            Total
        ),

        format(
            '~w: ~w toneladas~n',
            [
                Pala,
                Total
            ]
        )
    ).