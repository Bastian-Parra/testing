package camion

// Estado representa el estado actual de un camión dentro
// de su ciclo de carga y transporte.
type Estado string

const (
	EstadoViajandoVacio   Estado = "ViajandoVacio"
	EstadoEnColaPala      Estado = "EnColaPala"
	EstadoCargando        Estado = "Cargando"
	EstadoTransportando   Estado = "Transportando"
	EstadoEnColaStockpile Estado = "EnColaStockpile"
	EstadoDescargando     Estado = "Descargando"
	EstadoEsperando       Estado = "Esperando"
	EstadoEnFalla         Estado = "EnFalla"
)

// String permite obtener el nombre textual del estado.
//
// Será útil posteriormente para la visualización y los logs.
func (e Estado) String() string {
	return string(e)
}
