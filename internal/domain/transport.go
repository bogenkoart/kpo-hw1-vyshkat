package domain

import (
	"errors"
	"fmt"
)

// Простота для новичка задаётся по шкале 1–10.
const (
	MinSimplicity = 1
	MaxSimplicity = 10
)

var (
	ErrSimplicityOutOfRange = fmt.Errorf("простота для новичка должна быть от %d до %d", MinSimplicity, MaxSimplicity)
	ErrWearOutOfRange       = errors.New("износ должен быть от 0 до 100 %")
)

// Vehicle — прокатный транспорт.
// Электротранспорт дополнительно реализует EnergyConsumer, обычный велосипед — нет.
type Vehicle interface {
	Inventoried
	// Simplicity — «простота для новичка», 1–10.
	Simplicity() int
	// WearPercent — износ в процентах; его оценивает техосмотр.
	WearPercent() int
}

// VehicleSpec — общие параметры любого транспорта, которые вводит оператор.
type VehicleSpec struct {
	Name        string
	Simplicity  int
	WearPercent int
}

// vehicleBase — общая часть всех видов транспорта.
type vehicleBase struct {
	Asset
	simplicity int
	wear       int
}

func newVehicleBase(s VehicleSpec) (vehicleBase, error) {
	asset, err := newAsset(s.Name)
	if err != nil {
		return vehicleBase{}, err
	}
	if s.Simplicity < MinSimplicity || s.Simplicity > MaxSimplicity {
		return vehicleBase{}, ErrSimplicityOutOfRange
	}
	if s.WearPercent < 0 || s.WearPercent > 100 {
		return vehicleBase{}, ErrWearOutOfRange
	}
	return vehicleBase{Asset: asset, simplicity: s.Simplicity, wear: s.WearPercent}, nil
}

// Simplicity возвращает простоту для новичка.
func (v *vehicleBase) Simplicity() int { return v.simplicity }

// WearPercent возвращает износ.
func (v *vehicleBase) WearPercent() int { return v.wear }

// Bicycle — обычный велосипед, энергию не потребляет.
type Bicycle struct{ vehicleBase }

// NewBicycle создаёт велосипед.
func NewBicycle(s VehicleSpec) (*Bicycle, error) {
	base, err := newVehicleBase(s)
	if err != nil {
		return nil, err
	}
	return &Bicycle{base}, nil
}

// Kind возвращает вид транспорта.
func (*Bicycle) Kind() string { return "Велосипед" }

// Scooter — электросамокат.
type Scooter struct {
	vehicleBase
	energy
}

// NewScooter создаёт электросамокат с суточным расходом energyKWh.
func NewScooter(s VehicleSpec, energyKWh float64) (*Scooter, error) {
	base, err := newVehicleBase(s)
	if err != nil {
		return nil, err
	}
	e, err := newEnergy(energyKWh)
	if err != nil {
		return nil, err
	}
	return &Scooter{vehicleBase: base, energy: e}, nil
}

// Kind возвращает вид транспорта.
func (*Scooter) Kind() string { return "Электросамокат" }

// EBike — электровелосипед.
type EBike struct {
	vehicleBase
	energy
}

// NewEBike создаёт электровелосипед с суточным расходом energyKWh.
func NewEBike(s VehicleSpec, energyKWh float64) (*EBike, error) {
	base, err := newVehicleBase(s)
	if err != nil {
		return nil, err
	}
	e, err := newEnergy(energyKWh)
	if err != nil {
		return nil, err
	}
	return &EBike{vehicleBase: base, energy: e}, nil
}

// Kind возвращает вид транспорта.
func (*EBike) Kind() string { return "Электровелосипед" }

// Проверки на этапе компиляции: типы действительно реализуют нужные интерфейсы.
var (
	_ Vehicle        = (*Bicycle)(nil)
	_ Vehicle        = (*Scooter)(nil)
	_ Vehicle        = (*EBike)(nil)
	_ EnergyConsumer = (*Scooter)(nil)
	_ EnergyConsumer = (*EBike)(nil)
)
