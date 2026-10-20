package fleet

import (
	"errors"
	"math"
	"testing"

	"vyshkat/internal/domain"
)

// fakeInspector — подмена техосмотра: возвращает заранее заданное решение
// и запоминает, какие устройства ему показывали.
type fakeInspector struct {
	result    domain.InspectionResult
	inspected []domain.Vehicle
}

func (f *fakeInspector) Inspect(v domain.Vehicle) domain.InspectionResult {
	f.inspected = append(f.inspected, v)
	return f.result
}

func approving() *fakeInspector { return &fakeInspector{result: domain.Approve()} }

func scooter(t *testing.T, simplicity int, kwh float64) *domain.Scooter {
	t.Helper()
	v, err := domain.NewScooter(domain.VehicleSpec{Name: "Самокат", Simplicity: simplicity}, kwh)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func bicycle(t *testing.T, simplicity int) *domain.Bicycle {
	t.Helper()
	v, err := domain.NewBicycle(domain.VehicleSpec{Name: "Велосипед", Simplicity: simplicity})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestAcceptVehicle_ApprovedByInspection_GetsNumber(t *testing.T) {
	insp := approving()
	f := New(insp, MinSimplicityRule(DefaultBeginnerThreshold))
	v := scooter(t, 7, 0.4)

	res, err := f.AcceptVehicle(v)

	if err != nil || !res.Approved {
		t.Fatalf("ожидали приёмку, получили %+v, %v", res, err)
	}
	if len(insp.inspected) != 1 || insp.inspected[0] != v {
		t.Error("транспорт должен пройти через техосмотр ровно один раз")
	}
	if v.InventoryNumber() != "VK-T-0001" {
		t.Errorf("инв. номер = %q", v.InventoryNumber())
	}
	if got := f.Report().Vehicles; got != 1 {
		t.Errorf("в парке %d единиц, ожидали 1", got)
	}
}

func TestAcceptVehicle_RejectedByInspection_NotAdded(t *testing.T) {
	insp := &fakeInspector{result: domain.Reject("сломан тормоз")}
	f := New(insp, MinSimplicityRule(DefaultBeginnerThreshold))
	v := scooter(t, 7, 0.4)

	res, err := f.AcceptVehicle(v)

	if err != nil || res.Approved {
		t.Fatalf("ожидали отказ, получили %+v, %v", res, err)
	}
	if v.InventoryNumber() != "" {
		t.Error("отклонённый транспорт не должен получать инвентарный номер")
	}
	if len(f.Inventory()) != 0 {
		t.Error("отклонённый транспорт не должен попадать в парк")
	}
	rej := f.Rejections()
	if len(rej) != 1 || rej[0].Vehicle != v || rej[0].Result.Reasons[0] != "сломан тормоз" {
		t.Errorf("журнал отказов: %+v", rej)
	}
}

func TestAcceptVehicle_Errors(t *testing.T) {
	f := New(approving(), MinSimplicityRule(6))
	if _, err := f.AcceptVehicle(nil); !errors.Is(err, ErrNilAsset) {
		t.Errorf("nil: %v", err)
	}
	v := scooter(t, 7, 0.4)
	_, _ = f.AcceptVehicle(v)
	if _, err := f.AcceptVehicle(v); !errors.Is(err, ErrAlreadyOnBalance) {
		t.Errorf("повтор: %v", err)
	}
}

func TestAddItem(t *testing.T) {
	f := New(approving(), MinSimplicityRule(6))
	h, _ := domain.NewHelmet("Шлем", "M")

	n, err := f.AddItem(h)
	if err != nil || n != "VK-I-0001" || h.InventoryNumber() != n {
		t.Fatalf("AddItem: %q %v", n, err)
	}
	if _, err := f.AddItem(h); !errors.Is(err, ErrAlreadyOnBalance) {
		t.Errorf("повтор: %v", err)
	}
	if _, err := f.AddItem(nil); !errors.Is(err, ErrNilAsset) {
		t.Errorf("nil: %v", err)
	}
}

func TestBeginnerVehicles_ThresholdSix(t *testing.T) {
	f := New(approving(), MinSimplicityRule(DefaultBeginnerThreshold))
	hard := scooter(t, 5, 0.4) // ниже порога
	edge := scooter(t, 6, 0.4) // ровно порог — подходит
	easy := bicycle(t, 9)
	for _, v := range []domain.Vehicle{hard, edge, easy} {
		if _, err := f.AcceptVehicle(v); err != nil {
			t.Fatal(err)
		}
	}

	got := f.BeginnerVehicles()

	if len(got) != 2 || got[0] != edge || got[1] != easy {
		t.Errorf("для новичков: %v", got)
	}
}

func TestBeginnerVehicles_CustomRuleInjected(t *testing.T) {
	onlyBicycles := func(v domain.Vehicle) bool { _, ok := v.(*domain.Bicycle); return ok }
	f := New(approving(), onlyBicycles)
	_, _ = f.AcceptVehicle(scooter(t, 10, 0.4))
	b := bicycle(t, 1)
	_, _ = f.AcceptVehicle(b)

	got := f.BeginnerVehicles()
	if len(got) != 1 || got[0] != b {
		t.Errorf("правило не подменилось: %v", got)
	}
}

func TestTotalEnergyAndReport(t *testing.T) {
	f := New(approving(), MinSimplicityRule(6))
	_, _ = f.AcceptVehicle(scooter(t, 7, 0.45))
	_, _ = f.AcceptVehicle(bicycle(t, 9)) // не потребляет
	helmet, _ := domain.NewHelmet("Шлем", "M")
	dock, _ := domain.NewDockStation("Док", 4, 2.4)
	_, _ = f.AddItem(helmet)
	_, _ = f.AddItem(dock)

	if got := f.TotalDailyKWh(); math.Abs(got-2.85) > 1e-9 {
		t.Errorf("энергия = %v, ожидали 2.85", got)
	}
	r := f.Report()
	if r.Vehicles != 2 || r.Items != 2 || r.Total() != 4 || r.Rejected != 0 {
		t.Errorf("отчёт: %+v", r)
	}
	if r.ByKind["Электросамокат"] != 1 || r.ByKind["Шлем"] != 1 {
		t.Errorf("разбивка по видам: %v", r.ByKind)
	}
}

func TestInventory_VehiclesThenItems(t *testing.T) {
	f := New(approving(), MinSimplicityRule(6))
	h, _ := domain.NewHelmet("Шлем", "M")
	_, _ = f.AddItem(h)
	s := scooter(t, 7, 0.4)
	_, _ = f.AcceptVehicle(s)

	inv := f.Inventory()
	if len(inv) != 2 || inv[0] != s || inv[1] != h {
		t.Errorf("ведомость: %v", inv)
	}
}
