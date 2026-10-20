package inspection

import (
	"testing"

	"vyshkat/internal/domain"
)

func spec(wear int) domain.VehicleSpec {
	return domain.VehicleSpec{Name: "Тест", Simplicity: 6, WearPercent: wear}
}

func TestServiceCenter_Inspect(t *testing.T) {
	sc := NewServiceCenter(Config{MaxWearPercent: 60, MaxDailyKWh: 1.5})

	mustScooter := func(wear int, kwh float64) domain.Vehicle {
		v, err := domain.NewScooter(spec(wear), kwh)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	bike, _ := domain.NewBicycle(spec(10))
	wornBike, _ := domain.NewBicycle(spec(61))

	tests := []struct {
		name        string
		v           domain.Vehicle
		wantOK      bool
		wantReasons int
	}{
		{"исправный самокат", mustScooter(10, 0.5), true, 0},
		{"износ ровно на границе", mustScooter(60, 0.5), true, 0},
		{"износ выше нормы", mustScooter(61, 0.5), false, 1},
		{"расход выше нормы", mustScooter(10, 1.6), false, 1},
		{"не указан расход", mustScooter(10, 0), false, 1},
		{"износ и расход сразу", mustScooter(90, 5), false, 2},
		{"велосипед без энергии принимается", bike, true, 0},
		{"изношенный велосипед", wornBike, false, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sc.Inspect(tt.v)
			if got.Approved != tt.wantOK || len(got.Reasons) != tt.wantReasons {
				t.Errorf("Inspect() = %+v, ожидали approved=%v, причин=%d", got, tt.wantOK, tt.wantReasons)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.MaxWearPercent <= 0 || cfg.MaxDailyKWh <= 0 {
		t.Errorf("нормативы по умолчанию должны быть положительными: %+v", cfg)
	}
}
