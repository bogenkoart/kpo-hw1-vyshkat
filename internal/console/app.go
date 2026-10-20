// Package console — консольный интерфейс учёта парка «ВышКат».
package console

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"vyshkat/internal/domain"
	"vyshkat/internal/fleet"
)

// App — меню и вывод. Бизнес-логики здесь нет: всё делегируется в fleet.Fleet.
type App struct {
	p     *prompter
	out   io.Writer
	fleet *fleet.Fleet
}

// New создаёт консольное приложение. in/out передаются снаружи, чтобы приложение можно было тестировать.
func New(in io.Reader, out io.Writer, f *fleet.Fleet) *App {
	return &App{p: &prompter{in: bufio.NewScanner(in), out: out}, out: out, fleet: f}
}

const menu = `
==== ВышКат · «Доедем до дедлайна» ====
1. Принять транспорт (с техосмотром)
2. Поставить вещь на баланс
3. Отчёт по парку (число единиц и энергия)
4. Транспорт для новичков
5. Инвентарная ведомость (все номера)
6. Журнал отказов техосмотра
7. Запустить демо-сценарий
0. Выход`

// Run — главный цикл меню. Завершается по «0» или по концу ввода.
func (a *App) Run() error {
	for {
		fmt.Fprintln(a.out, menu)
		choice, err := a.p.line("> ")
		if errors.Is(err, ErrInputClosed) || choice == "0" {
			fmt.Fprintln(a.out, "До встречи на дедлайне!")
			return nil
		}

		switch choice {
		case "1":
			err = a.acceptVehicle()
		case "2":
			err = a.addItem()
		case "3":
			a.PrintReport()
		case "4":
			a.PrintBeginners()
		case "5":
			a.PrintInventory()
		case "6":
			a.PrintRejections()
		case "7":
			err = a.RunDemo()
		default:
			fmt.Fprintln(a.out, "  ! неизвестный пункт меню")
		}

		if errors.Is(err, ErrInputClosed) {
			fmt.Fprintln(a.out, "До встречи на дедлайне!")
			return nil
		}
		if err != nil {
			fmt.Fprintln(a.out, "  ! ошибка:", err)
		}
	}
}

func (a *App) acceptVehicle() error {
	fmt.Fprintln(a.out, "Вид транспорта:")
	for i, t := range vehicleTypes {
		fmt.Fprintf(a.out, "  %d. %s\n", i+1, t.title)
	}
	idx, err := a.p.integer("> ", 1, len(vehicleTypes))
	if err != nil {
		return err
	}
	t := vehicleTypes[idx-1]

	var spec domain.VehicleSpec
	if spec.Name, err = a.p.text("Наименование (модель): "); err != nil {
		return err
	}
	if spec.Simplicity, err = a.p.integer("Простота для новичка (1–10): ", domain.MinSimplicity, domain.MaxSimplicity); err != nil {
		return err
	}
	if spec.WearPercent, err = a.p.integer("Износ, % (0–100): ", 0, 100); err != nil {
		return err
	}
	var kwh float64
	if t.electric {
		if kwh, err = a.p.number("Суточный расход, кВт·ч: "); err != nil {
			return err
		}
	}

	v, err := t.build(spec, kwh)
	if err != nil {
		return err
	}
	return a.accept(v)
}

func (a *App) accept(v domain.Vehicle) error {
	result, err := a.fleet.AcceptVehicle(v)
	if err != nil {
		return err
	}
	if result.Approved {
		fmt.Fprintf(a.out, "  ✔ %s «%s» прошёл техосмотр, инв. № %s\n", v.Kind(), v.Name(), v.InventoryNumber())
	} else {
		fmt.Fprintf(a.out, "  ✘ %s «%s» %s\n", v.Kind(), v.Name(), result)
	}
	return nil
}

func (a *App) addItem() error {
	fmt.Fprintln(a.out, "Вид вещи:")
	for i, t := range itemTypes {
		fmt.Fprintf(a.out, "  %d. %s\n", i+1, t.title)
	}
	idx, err := a.p.integer("> ", 1, len(itemTypes))
	if err != nil {
		return err
	}
	name, err := a.p.text("Наименование: ")
	if err != nil {
		return err
	}
	it, err := itemTypes[idx-1].read(a.p, name)
	if err != nil {
		return err
	}
	return a.add(it)
}

func (a *App) add(it domain.Item) error {
	number, err := a.fleet.AddItem(it)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.out, "  ✔ %s «%s» на балансе, инв. № %s\n", it.Kind(), it.Name(), number)
	return nil
}

// PrintReport выводит число единиц по видам и суммарное суточное потребление.
func (a *App) PrintReport() {
	r := a.fleet.Report()
	fmt.Fprintln(a.out, "--- Отчёт по парку ---")
	fmt.Fprintf(a.out, "Транспорт: %d, вещи: %d, всего на балансе: %d\n", r.Vehicles, r.Items, r.Total())

	kinds := make([]string, 0, len(r.ByKind))
	for k := range r.ByKind {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	for _, k := range kinds {
		fmt.Fprintf(a.out, "  %-18s %d\n", k, r.ByKind[k])
	}
	fmt.Fprintf(a.out, "Суммарное потребление: %.2f кВт·ч/сут\n", r.TotalKWh)
	fmt.Fprintf(a.out, "Отклонено техосмотром: %d\n", r.Rejected)
}

// PrintBeginners выводит транспорт, который можно выдавать новичкам.
func (a *App) PrintBeginners() {
	list := a.fleet.BeginnerVehicles()
	fmt.Fprintln(a.out, "--- Транспорт для новичков ---")
	if len(list) == 0 {
		fmt.Fprintln(a.out, "(пока нет)")
		return
	}
	w := tabwriter.NewWriter(a.out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Инв. №\tВид\tНаименование\tПростота")
	for _, v := range list {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", v.InventoryNumber(), v.Kind(), v.Name(), v.Simplicity())
	}
	w.Flush()
}

// PrintInventory выводит наименования и инвентарные номера всей техники и вещей.
func (a *App) PrintInventory() {
	list := a.fleet.Inventory()
	fmt.Fprintln(a.out, "--- Инвентарная ведомость ---")
	if len(list) == 0 {
		fmt.Fprintln(a.out, "(на балансе ничего нет)")
		return
	}
	w := tabwriter.NewWriter(a.out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Инв. №\tВид\tНаименование\tкВт·ч/сут")
	for _, item := range list {
		kwh := "—"
		if c, ok := item.(domain.EnergyConsumer); ok {
			kwh = fmt.Sprintf("%.2f", c.DailyEnergyKWh())
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", item.InventoryNumber(), item.Kind(), item.Name(), kwh)
	}
	w.Flush()
}

// PrintRejections выводит устройства, не прошедшие техосмотр.
func (a *App) PrintRejections() {
	list := a.fleet.Rejections()
	fmt.Fprintln(a.out, "--- Журнал отказов техосмотра ---")
	if len(list) == 0 {
		fmt.Fprintln(a.out, "(отказов нет)")
		return
	}
	for _, r := range list {
		fmt.Fprintf(a.out, "  %s «%s»: %s\n", r.Vehicle.Kind(), r.Vehicle.Name(), r.Result)
	}
}
