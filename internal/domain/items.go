package domain

import "errors"

// ErrNonPositiveCapacity — у станции/шкафа должно быть хотя бы одно место.
var ErrNonPositiveCapacity = errors.New("количество мест должно быть больше нуля")

// Item — вещь на балансе (не транспорт): шлемы, док-станции, зарядные шкафы.
// Отдельный тип нужен, чтобы вещь нельзя было случайно «принять» как транспорт.
type Item interface {
	Inventoried
	isItem()
}

// item — общая часть всех вещей.
type item struct{ Asset }

func (*item) isItem() {}

func newItem(name string) (item, error) {
	a, err := newAsset(name)
	return item{a}, err
}

// Helmet — шлем, энергию не потребляет.
type Helmet struct {
	item
	size string
}

// NewHelmet создаёт шлем указанного размера.
func NewHelmet(name, size string) (*Helmet, error) {
	it, err := newItem(name)
	if err != nil {
		return nil, err
	}
	return &Helmet{item: it, size: size}, nil
}

// Kind возвращает вид вещи.
func (*Helmet) Kind() string { return "Шлем" }

// Size возвращает размер шлема.
func (h *Helmet) Size() string { return h.size }

// DockStation — док-станция для парковки и подзарядки самокатов.
type DockStation struct {
	item
	energy
	slots int
}

// NewDockStation создаёт док-станцию.
func NewDockStation(name string, slots int, energyKWh float64) (*DockStation, error) {
	if slots <= 0 {
		return nil, ErrNonPositiveCapacity
	}
	it, err := newItem(name)
	if err != nil {
		return nil, err
	}
	e, err := newEnergy(energyKWh)
	if err != nil {
		return nil, err
	}
	return &DockStation{item: it, energy: e, slots: slots}, nil
}

// Kind возвращает вид вещи.
func (*DockStation) Kind() string { return "Док-станция" }

// Slots возвращает число парковочных мест.
func (d *DockStation) Slots() int { return d.slots }

// ChargingCabinet — зарядный шкаф для съёмных аккумуляторов.
type ChargingCabinet struct {
	item
	energy
	cells int
}

// NewChargingCabinet создаёт зарядный шкаф.
func NewChargingCabinet(name string, cells int, energyKWh float64) (*ChargingCabinet, error) {
	if cells <= 0 {
		return nil, ErrNonPositiveCapacity
	}
	it, err := newItem(name)
	if err != nil {
		return nil, err
	}
	e, err := newEnergy(energyKWh)
	if err != nil {
		return nil, err
	}
	return &ChargingCabinet{item: it, energy: e, cells: cells}, nil
}

// Kind возвращает вид вещи.
func (*ChargingCabinet) Kind() string { return "Зарядный шкаф" }

// Cells возвращает число ячеек.
func (c *ChargingCabinet) Cells() int { return c.cells }

var (
	_ Item           = (*Helmet)(nil)
	_ Item           = (*DockStation)(nil)
	_ Item           = (*ChargingCabinet)(nil)
	_ EnergyConsumer = (*DockStation)(nil)
	_ EnergyConsumer = (*ChargingCabinet)(nil)
)
