// Package di — единственное место, где собирается граф зависимостей (composition root).
package di

import (
	"io"

	"go.uber.org/dig"

	"vyshkat/internal/console"
	"vyshkat/internal/fleet"
	"vyshkat/internal/inspection"
)

// IO — потоки ввода/вывода консольного приложения.
type IO struct {
	In  io.Reader
	Out io.Writer
}

// NewContainer регистрирует все зависимости приложения в DI-контейнере dig.
// Тесты могут подменить любую из них через container.Decorate.
func NewContainer(streams IO) (*dig.Container, error) {
	c := dig.New()

	providers := []any{
		func() IO { return streams },
		inspection.DefaultConfig,
		inspection.NewServiceCenter,
		// Связываем абстракцию fleet.Inspector с реализацией *inspection.ServiceCenter.
		func(sc *inspection.ServiceCenter) fleet.Inspector { return sc },
		func() fleet.BeginnerRule { return fleet.MinSimplicityRule(fleet.DefaultBeginnerThreshold) },
		fleet.New,
		func(s IO, f *fleet.Fleet) *console.App { return console.New(s.In, s.Out, f) },
	}
	for _, p := range providers {
		if err := c.Provide(p); err != nil {
			return nil, err
		}
	}
	return c, nil
}
