BEGIN {
 FS = ","
}

# ignoramos la primera linea

NR > 1 {
 seq = $1
 tick = $2
 camion = $3
 evento = $4
 recurso = $5
 toneladas = $6
 if (entidad ~ /^Pala/) {
  camion = recurso
  pala = entidad
  if (evento == "Camion a la pala") {
   printf("adquiere(%s, %s, '%s', '%s').\n", seq, tick, camion, recurso)
  } 
  else if (evento == "Inicio de carga") {
   printf("libera(%s, %s, '%s', '%s').\n", seq, tick, camion, recurso)
  }
  else if (evento == "Fin de carga") {
   printf("carga('%s', %s, %s, '%s', '%s', %s).\n", recurso, seq, tick, camion, recurso, toneladas)
  }
 }

 else if (entidad ~ /^C-/) {
  camion = entidad
  recurso = "C-" 
  if (evento == "Inicio de viaje vacío") {
   printf("viaja(%s, %s, '%s', '%s').\n", seq, tick, camion, recurso)
  }
 }
}