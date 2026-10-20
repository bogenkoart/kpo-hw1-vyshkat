package domain

import "strings"

// InspectionResult — акт техосмотра: принять устройство в парк или нет и почему.
type InspectionResult struct {
	Approved bool
	Reasons  []string
}

// Approve — положительное заключение.
func Approve() InspectionResult { return InspectionResult{Approved: true} }

// Reject — отказ с причинами.
func Reject(reasons ...string) InspectionResult {
	return InspectionResult{Approved: false, Reasons: reasons}
}

// String возвращает заключение в человекочитаемом виде.
func (r InspectionResult) String() string {
	if r.Approved {
		return "принят"
	}
	return "отклонён: " + strings.Join(r.Reasons, "; ")
}
