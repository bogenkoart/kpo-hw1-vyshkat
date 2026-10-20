// Package domain описывает предметную область «ВышКат»: что стоит на балансе,
// что потребляет энергию и что можно выдавать в прокат.
package domain

import (
	"errors"
	"strings"
)

// Inventoried — любой объект, который ставится на баланс и получает инвентарный номер.
// Сотрудники этот интерфейс не реализуют: на баланс их не ставят.
type Inventoried interface {
	Name() string
	Kind() string
	InventoryNumber() string
	// AssignInventoryNumber вызывается один раз — при постановке на баланс.
	AssignInventoryNumber(number string) error
}

// EnergyConsumer — всё, что потребляет электроэнергию.
type EnergyConsumer interface {
	DailyEnergyKWh() float64
}

var (
	ErrEmptyName           = errors.New("наименование не может быть пустым")
	ErrEmptyNumber         = errors.New("инвентарный номер не может быть пустым")
	ErrNumberAlreadyIssued = errors.New("инвентарный номер уже присвоен")
	ErrNegativeEnergy      = errors.New("суточный расход энергии не может быть отрицательным")
)

// Asset — общая часть всех объектов на балансе (встраивается в конкретные типы).
type Asset struct {
	name   string
	number string
}

func newAsset(name string) (Asset, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Asset{}, ErrEmptyName
	}
	return Asset{name: name}, nil
}

// Name возвращает наименование объекта.
func (a *Asset) Name() string { return a.name }

// InventoryNumber возвращает инвентарный номер (пусто, пока объект не на балансе).
func (a *Asset) InventoryNumber() string { return a.number }

// AssignInventoryNumber присваивает номер; повторно переназначить номер нельзя.
func (a *Asset) AssignInventoryNumber(number string) error {
	if strings.TrimSpace(number) == "" {
		return ErrEmptyNumber
	}
	if a.number != "" {
		return ErrNumberAlreadyIssued
	}
	a.number = number
	return nil
}

// energy — встраиваемый «кусочек» для всего, что потребляет энергию.
type energy struct {
	dailyKWh float64
}

func newEnergy(kwh float64) (energy, error) {
	if kwh < 0 {
		return energy{}, ErrNegativeEnergy
	}
	return energy{dailyKWh: kwh}, nil
}

// DailyEnergyKWh возвращает суточный расход энергии, кВт·ч.
func (e energy) DailyEnergyKWh() float64 { return e.dailyKWh }
