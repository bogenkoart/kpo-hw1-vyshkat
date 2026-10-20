package domain

import (
	"errors"
	"testing"
)

func validSpec() VehicleSpec {
	return VehicleSpec{Name: "Тест", Simplicity: 6, WearPercent: 10}
}

func TestNewVehicle_Validation(t *testing.T) {
	tests := []struct {
		name string
		spec VehicleSpec
		want error
	}{
		{"пустое имя", VehicleSpec{Name: "  ", Simplicity: 5}, ErrEmptyName},
		{"простота 0", VehicleSpec{Name: "x", Simplicity: 0}, ErrSimplicityOutOfRange},
		{"простота 11", VehicleSpec{Name: "x", Simplicity: 11}, ErrSimplicityOutOfRange},
		{"отрицательный износ", VehicleSpec{Name: "x", Simplicity: 5, WearPercent: -1}, ErrWearOutOfRange},
		{"износ больше 100", VehicleSpec{Name: "x", Simplicity: 5, WearPercent: 101}, ErrWearOutOfRange},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewBicycle(tt.spec); !errors.Is(err, tt.want) {
				t.Errorf("NewBicycle: ошибка %v, ожидали %v", err, tt.want)
			}
			if _, err := NewScooter(tt.spec, 0.5); !errors.Is(err, tt.want) {
				t.Errorf("NewScooter: ошибка %v, ожидали %v", err, tt.want)
			}
			if _, err := NewEBike(tt.spec, 0.5); !errors.Is(err, tt.want) {
				t.Errorf("NewEBike: ошибка %v, ожидали %v", err, tt.want)
			}
		})
	}
}

func TestElectricVehicles_RejectNegativeEnergy(t *testing.T) {
	if _, err := NewScooter(validSpec(), -0.1); !errors.Is(err, ErrNegativeEnergy) {
		t.Errorf("самокат: %v", err)
	}
	if _, err := NewEBike(validSpec(), -0.1); !errors.Is(err, ErrNegativeEnergy) {
		t.Errorf("электровелосипед: %v", err)
	}
}

func TestVehicles_KindsAndEnergy(t *testing.T) {
	bike, _ := NewBicycle(validSpec())
	scooter, _ := NewScooter(validSpec(), 0.4)
	ebike, _ := NewEBike(validSpec(), 0.8)

	tests := []struct {
		v        Vehicle
		kind     string
		consumer bool
		kwh      float64
	}{
		{bike, "Велосипед", false, 0},
		{scooter, "Электросамокат", true, 0.4},
		{ebike, "Электровелосипед", true, 0.8},
	}
	for _, tt := range tests {
		if tt.v.Kind() != tt.kind {
			t.Errorf("Kind() = %q, ожидали %q", tt.v.Kind(), tt.kind)
		}
		if tt.v.Simplicity() != 6 || tt.v.WearPercent() != 10 || tt.v.Name() != "Тест" {
			t.Errorf("%s: неверные базовые поля", tt.kind)
		}
		c, ok := tt.v.(EnergyConsumer)
		if ok != tt.consumer {
			t.Fatalf("%s: EnergyConsumer = %v, ожидали %v", tt.kind, ok, tt.consumer)
		}
		if ok && c.DailyEnergyKWh() != tt.kwh {
			t.Errorf("%s: расход %v, ожидали %v", tt.kind, c.DailyEnergyKWh(), tt.kwh)
		}
	}
}

func TestAssignInventoryNumber(t *testing.T) {
	h, _ := NewHelmet("Шлем", "L")
	if h.InventoryNumber() != "" {
		t.Fatal("у нового объекта не должно быть номера")
	}
	if err := h.AssignInventoryNumber(" "); !errors.Is(err, ErrEmptyNumber) {
		t.Errorf("пустой номер: %v", err)
	}
	if err := h.AssignInventoryNumber("VK-1"); err != nil {
		t.Fatalf("первое присвоение: %v", err)
	}
	if err := h.AssignInventoryNumber("VK-2"); !errors.Is(err, ErrNumberAlreadyIssued) {
		t.Errorf("повторное присвоение: %v", err)
	}
	if h.InventoryNumber() != "VK-1" {
		t.Errorf("номер изменился: %s", h.InventoryNumber())
	}
}

func TestItems(t *testing.T) {
	helmet, err := NewHelmet("Urban", "M")
	if err != nil || helmet.Size() != "M" || helmet.Kind() != "Шлем" {
		t.Fatalf("шлем: %v %+v", err, helmet)
	}
	if _, ok := any(helmet).(EnergyConsumer); ok {
		t.Error("шлем не должен потреблять энергию")
	}

	dock, err := NewDockStation("Док", 6, 2.4)
	if err != nil || dock.Slots() != 6 || dock.DailyEnergyKWh() != 2.4 || dock.Kind() != "Док-станция" {
		t.Fatalf("док-станция: %v %+v", err, dock)
	}

	cab, err := NewChargingCabinet("Шкаф", 12, 4.8)
	if err != nil || cab.Cells() != 12 || cab.DailyEnergyKWh() != 4.8 || cab.Kind() != "Зарядный шкаф" {
		t.Fatalf("шкаф: %v %+v", err, cab)
	}
}

func TestItems_Validation(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want error
	}{
		{"шлем без имени", second(NewHelmet("", "M")), ErrEmptyName},
		{"док без мест", second(NewDockStation("Док", 0, 1)), ErrNonPositiveCapacity},
		{"док без имени", second(NewDockStation("", 1, 1)), ErrEmptyName},
		{"док с отрицательной энергией", second(NewDockStation("Док", 1, -1)), ErrNegativeEnergy},
		{"шкаф без ячеек", second(NewChargingCabinet("Шкаф", 0, 1)), ErrNonPositiveCapacity},
		{"шкаф без имени", second(NewChargingCabinet("", 1, 1)), ErrEmptyName},
		{"шкаф с отрицательной энергией", second(NewChargingCabinet("Шкаф", 1, -1)), ErrNegativeEnergy},
	}
	for _, c := range cases {
		if !errors.Is(c.err, c.want) {
			t.Errorf("%s: %v, ожидали %v", c.name, c.err, c.want)
		}
	}
}

func second[T any](_ T, err error) error { return err }

func TestInspectionResult_String(t *testing.T) {
	if Approve().String() != "принят" {
		t.Error(Approve().String())
	}
	if got := Reject("a", "b").String(); got != "отклонён: a; b" {
		t.Error(got)
	}
}
