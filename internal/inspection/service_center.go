// Package inspection — сервис техосмотра (ServiceCenter).
package inspection

import (
	"fmt"

	"vyshkat/internal/domain"
)

// Config — нормативы техосмотра.
type Config struct {
	// MaxWearPercent — максимально допустимый износ.
	MaxWearPercent int
	// MaxDailyKWh — максимально допустимый суточный расход: больше — подозрение на неисправную батарею.
	MaxDailyKWh float64
}

// DefaultConfig — нормативы по умолчанию.
func DefaultConfig() Config {
	return Config{MaxWearPercent: 60, MaxDailyKWh: 1.5}
}

// ServiceCenter решает, принимать ли устройство в активный парк.
type ServiceCenter struct {
	cfg Config
}

// NewServiceCenter создаёт сервис техосмотра с заданными нормативами.
func NewServiceCenter(cfg Config) *ServiceCenter {
	return &ServiceCenter{cfg: cfg}
}

// Inspect проверяет устройство и возвращает акт техосмотра со всеми найденными замечаниями.
func (s *ServiceCenter) Inspect(v domain.Vehicle) domain.InspectionResult {
	var reasons []string

	if v.WearPercent() > s.cfg.MaxWearPercent {
		reasons = append(reasons,
			fmt.Sprintf("износ %d%% превышает норму %d%%", v.WearPercent(), s.cfg.MaxWearPercent))
	}

	if c, ok := v.(domain.EnergyConsumer); ok {
		kwh := c.DailyEnergyKWh()
		switch {
		case kwh == 0:
			reasons = append(reasons, "не указан расход энергии у электротранспорта")
		case kwh > s.cfg.MaxDailyKWh:
			reasons = append(reasons,
				fmt.Sprintf("расход %.2f кВт·ч/сут выше нормы %.2f — проверьте батарею", kwh, s.cfg.MaxDailyKWh))
		}
	}

	if len(reasons) > 0 {
		return domain.Reject(reasons...)
	}
	return domain.Approve()
}
