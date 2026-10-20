package console

import (
	"fmt"

	"vyshkat/internal/domain"
)

// must используется только для заведомо корректных демо-данных:
// паника здесь означает ошибку программиста, а не пользователя.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// RunDemo прогоняет все сценарии приложения на заранее заданных данных:
// успешная приёмка, отказы техосмотра, некорректные данные, вещи, отчёты.
func (a *App) RunDemo() error {
	step := func(title string) { fmt.Fprintf(a.out, "\n### %s\n", title) }

	step("1. Приёмка транспорта: исправная техника проходит техосмотр")
	scooter := must(domain.NewScooter(domain.VehicleSpec{Name: "Ninebot Max G30", Simplicity: 7, WearPercent: 10}, 0.45))
	bike := must(domain.NewBicycle(domain.VehicleSpec{Name: "Stels Navigator 300", Simplicity: 9, WearPercent: 25}))
	ebike := must(domain.NewEBike(domain.VehicleSpec{Name: "Volteco Bad Dual", Simplicity: 4, WearPercent: 5}, 0.9))
	for _, v := range []domain.Vehicle{scooter, bike, ebike} {
		if err := a.accept(v); err != nil {
			return err
		}
	}

	step("2. Отказ техосмотра: сильный износ и подозрительный расход энергии")
	worn := must(domain.NewScooter(domain.VehicleSpec{Name: "Kugoo S3 (б/у)", Simplicity: 8, WearPercent: 85}, 0.4))
	hungry := must(domain.NewEBike(domain.VehicleSpec{Name: "NoName E-Bike", Simplicity: 6, WearPercent: 20}, 3.2))
	for _, v := range []domain.Vehicle{worn, hungry} {
		if err := a.accept(v); err != nil {
			return err
		}
	}

	step("3. Некорректные данные отсекаются ещё при создании объекта")
	if _, err := domain.NewScooter(domain.VehicleSpec{Name: "Самокат-мечта", Simplicity: 11}, 0.3); err != nil {
		fmt.Fprintln(a.out, "  ! не создан:", err)
	}

	step("4. Повторная приёмка уже стоящего на балансе транспорта запрещена")
	if err := a.accept(scooter); err != nil {
		fmt.Fprintln(a.out, "  ! не принят:", err)
	}

	step("5. Вещи на балансе: шлем, док-станция, зарядный шкаф")
	helmet := must(domain.NewHelmet("Urban Pro", "M"))
	dock := must(domain.NewDockStation("Док-станция у корпуса на Покровке", 6, 2.4))
	cabinet := must(domain.NewChargingCabinet("Зарядный шкаф АКБ", 12, 4.8))
	for _, it := range []domain.Item{helmet, dock, cabinet} {
		if err := a.add(it); err != nil {
			return err
		}
	}

	step("6. Отчёт по парку")
	a.PrintReport()

	step("7. Транспорт для новичков (простота ≥ 6)")
	a.PrintBeginners()

	step("8. Инвентарная ведомость: вся техника и вещи")
	a.PrintInventory()

	step("9. Журнал отказов техосмотра")
	a.PrintRejections()
	return nil
}
