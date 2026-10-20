// Package fleet — учёт парка «ВышКат»: приёмка транспорта, вещи на балансе, отчёты.
package fleet

import (
	"errors"
	"fmt"

	"vyshkat/internal/domain"
)

// Inspector — то, что нужно парку от техосмотра. Интерфейс объявлен здесь,
// на стороне потребителя (идиома Go), поэтому fleet не зависит от пакета inspection.
type Inspector interface {
	Inspect(v domain.Vehicle) domain.InspectionResult
}

// BeginnerRule — правило «можно выдавать новичку». Вынесено отдельно,
// потому что заказчик прямо говорит, что правило может усложниться.
type BeginnerRule func(v domain.Vehicle) bool

// MinSimplicityRule — базовое правило: простота не ниже threshold.
func MinSimplicityRule(threshold int) BeginnerRule {
	return func(v domain.Vehicle) bool { return v.Simplicity() >= threshold }
}

// DefaultBeginnerThreshold — порог из ТЗ.
const DefaultBeginnerThreshold = 6

var (
	ErrNilAsset         = errors.New("объект не передан")
	ErrAlreadyOnBalance = errors.New("объект уже стоит на балансе")
)

// Rejection — запись об устройстве, не прошедшем техосмотр.
type Rejection struct {
	Vehicle domain.Vehicle
	Result  domain.InspectionResult
}

// Report — сводка по парку.
type Report struct {
	Vehicles int
	Items    int
	TotalKWh float64
	ByKind   map[string]int
	Rejected int
}

// Total — общее число единиц на балансе.
func (r Report) Total() int { return r.Vehicles + r.Items }

// Fleet хранит парк в памяти и выполняет над ним операции учёта.
type Fleet struct {
	inspector Inspector
	beginner  BeginnerRule

	vehicles []domain.Vehicle
	items    []domain.Item
	rejected []Rejection

	vehicleSeq int
	itemSeq    int
}

// New создаёт парк. Зависимости передаются снаружи (их собирает DI-контейнер).
func New(inspector Inspector, beginner BeginnerRule) *Fleet {
	return &Fleet{inspector: inspector, beginner: beginner}
}

// AcceptVehicle проводит техосмотр и, если он пройден, ставит транспорт на баланс.
// Отказ техосмотра — это не ошибка, а нормальный исход: он возвращается в акте.
func (f *Fleet) AcceptVehicle(v domain.Vehicle) (domain.InspectionResult, error) {
	if v == nil {
		return domain.InspectionResult{}, ErrNilAsset
	}
	if v.InventoryNumber() != "" {
		return domain.InspectionResult{}, ErrAlreadyOnBalance
	}

	result := f.inspector.Inspect(v)
	if !result.Approved {
		f.rejected = append(f.rejected, Rejection{Vehicle: v, Result: result})
		return result, nil
	}

	f.vehicleSeq++
	if err := v.AssignInventoryNumber(fmt.Sprintf("VK-T-%04d", f.vehicleSeq)); err != nil {
		return domain.InspectionResult{}, err
	}
	f.vehicles = append(f.vehicles, v)
	return result, nil
}

// AddItem ставит вещь на баланс (техосмотр для вещей не требуется) и возвращает её номер.
func (f *Fleet) AddItem(it domain.Item) (string, error) {
	if it == nil {
		return "", ErrNilAsset
	}
	if it.InventoryNumber() != "" {
		return "", ErrAlreadyOnBalance
	}
	f.itemSeq++
	number := fmt.Sprintf("VK-I-%04d", f.itemSeq)
	if err := it.AssignInventoryNumber(number); err != nil {
		return "", err
	}
	f.items = append(f.items, it)
	return number, nil
}

// Inventory — вся техника и вещи на балансе (сначала транспорт, затем вещи).
func (f *Fleet) Inventory() []domain.Inventoried {
	all := make([]domain.Inventoried, 0, len(f.vehicles)+len(f.items))
	for _, v := range f.vehicles {
		all = append(all, v)
	}
	for _, it := range f.items {
		all = append(all, it)
	}
	return all
}

// TotalDailyKWh — суммарное суточное потребление всего, что потребляет энергию.
func (f *Fleet) TotalDailyKWh() float64 {
	var total float64
	for _, a := range f.Inventory() {
		if c, ok := a.(domain.EnergyConsumer); ok {
			total += c.DailyEnergyKWh()
		}
	}
	return total
}

// BeginnerVehicles — прокатный транспорт, который можно выдавать новичкам.
func (f *Fleet) BeginnerVehicles() []domain.Vehicle {
	var res []domain.Vehicle
	for _, v := range f.vehicles {
		if f.beginner(v) {
			res = append(res, v)
		}
	}
	return res
}

// Rejections — журнал отказов техосмотра.
func (f *Fleet) Rejections() []Rejection {
	return append([]Rejection(nil), f.rejected...)
}

// Report — сводный отчёт: число единиц, разбивка по видам, суммарная энергия.
func (f *Fleet) Report() Report {
	byKind := make(map[string]int)
	for _, a := range f.Inventory() {
		byKind[a.Kind()]++
	}
	return Report{
		Vehicles: len(f.vehicles),
		Items:    len(f.items),
		TotalKWh: f.TotalDailyKWh(),
		ByKind:   byKind,
		Rejected: len(f.rejected),
	}
}
