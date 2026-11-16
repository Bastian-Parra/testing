% ==========================================
% Hechos de prueba para las invariantes
% ==========================================
%
% Este archivo representa un log alterado.
% Contiene errores introducidos a proposito
% para comprobar que Prolog los detecta.
% ==========================================


% ----------------------------------------------------------
% VIOLACION 1
% Dos camiones ocupan Pala-01 al mismo tiempo.
%
% C-01 entra primero.
% C-02 entra antes de que C-01 libere la pala.
% ----------------------------------------------------------

:- discontiguous adquiere/4.
:- discontiguous libera/4.
:- discontiguous carga/6.
:- discontiguous falla/5.
:- discontiguous repara/4.
:- discontiguous espera/5.
:- discontiguous viaja/4.

adquiere(1, 10, 'C-01', 'Pala-01').

adquiere(2, 11, 'C-02', 'Pala-01').

libera(3, 14, 'C-02', 'Pala-01').

libera(4, 15, 'C-01', 'Pala-01').


% ----------------------------------------------------------
% VIOLACION 2
% C-03 sufre una falla permanente
% y posteriormente registra movimiento.
% ----------------------------------------------------------

falla(
    5,
    20,
    'C-03',
    'Tramo-Norte',
    permanente
).

viaja(
    6,
    21,
    'C-03',
    'Tramo-Sur'
).


% ----------------------------------------------------------
% VIOLACION 3
% El camion registra una carga,
% pero no existe la carga equivalente
% registrada por la pala.
% ----------------------------------------------------------

carga(
    'C-04',
    7,
    25,
    'C-04',
    'Pala-02',
    50
).


% ----------------------------------------------------------
% CASO CORRECTO
% Esta carga aparece tanto en camion como en pala.
% No deberia generar una violacion.
% ----------------------------------------------------------

carga(
    'C-05',
    8,
    30,
    'C-05',
    'Pala-03',
    50
).

carga(
    'Pala-03',
    9,
    30,
    'C-05',
    'Pala-03',
    50
).

% ----------------------------------------------------------
% DATOS PARA PROBAR LAS CONSULTAS
% ----------------------------------------------------------


% C-07 queda detenido detras de C-03,
% que tiene una falla permanente en Tramo-Norte.

espera(
    10,
    22,
    'C-07',
    'Tramo-Norte',
    prolongada
).


% ----------------------------------------------------------
% DEADLOCK DE PRUEBA
%
% C-05 ocupa Tramo-A y espera Tramo-B.
% C-06 ocupa Tramo-B y espera Tramo-A.
% ----------------------------------------------------------

adquiere(
    11,
    30,
    'C-05',
    'Tramo-A'
).

adquiere(
    12,
    30,
    'C-06',
    'Tramo-B'
).

espera(
    13,
    31,
    'C-05',
    'Tramo-B',
    prolongada
).

espera(
    14,
    31,
    'C-06',
    'Tramo-A',
    prolongada
).