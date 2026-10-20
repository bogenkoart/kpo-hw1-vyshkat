// Команда vyshkat — консольное приложение учёта парка «ВышКат».
//
//	go run ./cmd/vyshkat         — интерактивное меню
//	go run ./cmd/vyshkat -demo   — прогнать демо-сценарий и выйти
package main

import (
	"flag"
	"fmt"
	"os"

	"vyshkat/internal/console"
	"vyshkat/internal/di"
)

func main() {
	demo := flag.Bool("demo", false, "прогнать демо-сценарий и выйти")
	flag.Parse()

	if err := run(*demo); err != nil {
		fmt.Fprintln(os.Stderr, "ошибка:", err)
		os.Exit(1)
	}
}

func run(demo bool) error {
	container, err := di.NewContainer(di.IO{In: os.Stdin, Out: os.Stdout})
	if err != nil {
		return err
	}
	return container.Invoke(func(app *console.App) error {
		if demo {
			return app.RunDemo()
		}
		return app.Run()
	})
}
