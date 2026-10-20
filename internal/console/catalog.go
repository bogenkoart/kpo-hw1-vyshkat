package console

import (
	"math"

	"vyshkat/internal/domain"
)

// Каталоги видов техники и вещей для меню.
// Новый вид транспорта = новый тип в domain + одна строка здесь; остальной код не меняется.

type vehicleType struct {
	title    string
	electric bool
	build    func(spec domain.VehicleSpec, kwh float64) (domain.Vehicle, error)
}

var vehicleTypes = []vehicleType{
	{"Электросамокат", true, func(s domain.VehicleSpec, kwh float64) (domain.Vehicle, error) {
		return domain.NewScooter(s, kwh)
	}},
	{"Велосипед", false, func(s domain.VehicleSpec, _ float64) (domain.Vehicle, error) {
		return domain.NewBicycle(s)
	}},
	{"Электровелосипед", true, func(s domain.VehicleSpec, kwh float64) (domain.Vehicle, error) {
		return domain.NewEBike(s, kwh)
	}},
}

type itemType struct {
	title string
	read  func(p *prompter, name string) (domain.Item, error)
}

var itemTypes = []itemType{
	{"Шлем", func(p *prompter, name string) (domain.Item, error) {
		size, err := p.text("Размер (S/M/L): ")
		if err != nil {
			return nil, err
		}
		return domain.NewHelmet(name, size)
	}},
	{"Док-станция", func(p *prompter, name string) (domain.Item, error) {
		slots, err := p.integer("Число мест: ", 1, math.MaxInt32)
		if err != nil {
			return nil, err
		}
		kwh, err := p.number("Суточный расход, кВт·ч: ")
		if err != nil {
			return nil, err
		}
		return domain.NewDockStation(name, slots, kwh)
	}},
	{"Зарядный шкаф", func(p *prompter, name string) (domain.Item, error) {
		cells, err := p.integer("Число ячеек: ", 1, math.MaxInt32)
		if err != nil {
			return nil, err
		}
		kwh, err := p.number("Суточный расход, кВт·ч: ")
		if err != nil {
			return nil, err
		}
		return domain.NewChargingCabinet(name, cells, kwh)
	}},
}
