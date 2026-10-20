package di

import (
	"bytes"
	"strings"
	"testing"

	"vyshkat/internal/console"
	"vyshkat/internal/domain"
	"vyshkat/internal/fleet"
	"vyshkat/internal/inspection"
)

// rejectAll — заглушка техосмотра, которая отклоняет всё подряд.
type rejectAll struct{ calls int }

func (r *rejectAll) Inspect(domain.Vehicle) domain.InspectionResult {
	r.calls++
	return domain.Reject("заглушка: техосмотр закрыт")
}

func TestContainer_ResolvesWholeGraph(t *testing.T) {
	c, err := NewContainer(IO{In: strings.NewReader(""), Out: &bytes.Buffer{}})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Invoke(func(app *console.App, insp fleet.Inspector) {
		if app == nil {
			t.Error("App не собран")
		}
		if _, ok := insp.(*inspection.ServiceCenter); !ok {
			t.Errorf("по умолчанию Inspector должен быть ServiceCenter, а не %T", insp)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Подмена техосмотра через DI-контейнер: Decorate заменяет fleet.Inspector,
// и все, кто от него зависит (fleet.Fleet), получают заглушку.
func TestContainer_InspectorCanBeReplaced(t *testing.T) {
	c, err := NewContainer(IO{In: strings.NewReader(""), Out: &bytes.Buffer{}})
	if err != nil {
		t.Fatal(err)
	}
	stub := &rejectAll{}
	if err := c.Decorate(func(fleet.Inspector) fleet.Inspector { return stub }); err != nil {
		t.Fatal(err)
	}

	err = c.Invoke(func(f *fleet.Fleet) {
		// Этот самокат реальный ServiceCenter бы принял.
		v, _ := domain.NewScooter(domain.VehicleSpec{Name: "Ninebot", Simplicity: 7}, 0.4)
		res, err := f.AcceptVehicle(v)
		if err != nil {
			t.Fatal(err)
		}
		if res.Approved || stub.calls != 1 {
			t.Errorf("заглушка не подставилась: %+v, вызовов %d", res, stub.calls)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestContainer_RunsDemo(t *testing.T) {
	out := &bytes.Buffer{}
	c, err := NewContainer(IO{In: strings.NewReader(""), Out: out})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Invoke(func(app *console.App) error { return app.RunDemo() }); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "VK-T-0001") {
		t.Error("демо не вывело инвентарные номера")
	}
}
