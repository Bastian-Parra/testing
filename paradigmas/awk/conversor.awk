BEGIN {
 FS = ","
}

# ignoramos la primera linea
NR > 1 {
 seq = $1
 tick = $2
 entidad = $3
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

 # --- Eventos agregados por tu compañero para el sistema de fallas ---
 if ($3 ~ /^C-/) {
  camion = $3
  recurso = $5

  if (evento == "temporal") {
   printf("falla_temporal(%s, %s, '%s', '%s').\n", seq, tick, camion, recurso)
  }
  else if (evento == "permanente") {
   printf("falla_permanente(%s, %s, '%s', '%s').\n", seq, tick, camion, recurso)
  }
  else if (evento == "panic_permanente") {
   printf("panic_recuperado(%s, %s, '%s', '%s').\n", seq, tick, camion, recurso)
  }
  else if (evento == "reparando") {
   printf("reparacion(%s, %s, '%s', '%s').\n", seq, tick, camion, recurso)
  }
  else if (evento == "entra") {
   printf("entra_tramo(%s, %s, '%s', '%s').\n", seq, tick, camion, recurso)
  }
  else if (evento == "sale") {
   printf("sale_tramo(%s, %s, '%s', '%s').\n", seq, tick, camion, recurso)
  }
  else if (evento == "avanza") {
   printf("avanza_tramo(%s, %s, '%s', '%s').\n", seq, tick, camion, recurso)
  }
  else if (evento == "carga" || evento == "descarga" || evento == "sin_ruta" || evento == "termina_por_error") {
   printf("evento_camion(%s, %s, '%s', '%s', '%s').\n", seq, tick, camion, evento, recurso)
  }
 }
}