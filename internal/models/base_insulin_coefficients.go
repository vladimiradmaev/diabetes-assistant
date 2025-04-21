package models

import (
	"strconv"
	"strings"
)

// BaseInsulinCoefficients represents the time-based insulin coefficients
// This is kept for backward compatibility with existing code
type BaseInsulinCoefficients struct {
	Morning   float64 `json:"morning" bson:"morning"`
	Afternoon float64 `json:"afternoon" bson:"afternoon"`
	Evening   float64 `json:"evening" bson:"evening"`
	Night     float64 `json:"night" bson:"night"`
}

// ConvertToInsulinPeriods converts the legacy BaseInsulinCoefficients format to an array of InsulinPeriod
func (b *BaseInsulinCoefficients) ConvertToInsulinPeriods() []InsulinPeriod {
	return []InsulinPeriod{
		{StartTime: "00:00", Hours: 6, Coefficient: b.Night},
		{StartTime: "06:00", Hours: 6, Coefficient: b.Morning},
		{StartTime: "12:00", Hours: 6, Coefficient: b.Afternoon},
		{StartTime: "18:00", Hours: 6, Coefficient: b.Evening},
	}
}

// InsulinPeriodsToBaseCoefficients converts an array of InsulinPeriod to the legacy BaseInsulinCoefficients format
func InsulinPeriodsToBaseCoefficients(periods []InsulinPeriod) BaseInsulinCoefficients {
	result := BaseInsulinCoefficients{
		Morning:   1.0,
		Afternoon: 1.0,
		Evening:   1.0,
		Night:     1.0,
	}

	for _, period := range periods {
		startHour, _ := strconv.Atoi(strings.Split(period.StartTime, ":")[0])
		endHour := startHour + int(period.Hours)

		if startHour >= 0 && startHour < 6 && endHour <= 6 {
			result.Night = period.Coefficient
		} else if startHour >= 6 && startHour < 12 && endHour <= 12 {
			result.Morning = period.Coefficient
		} else if startHour >= 12 && startHour < 18 && endHour <= 18 {
			result.Afternoon = period.Coefficient
		} else if startHour >= 18 && startHour < 24 && endHour <= 24 {
			result.Evening = period.Coefficient
		}
	}

	return result
}

// DefaultBaseCoefficients returns default base coefficients
func DefaultBaseCoefficients() []InsulinPeriod {
	return []InsulinPeriod{
		{
			StartTime:   "00:00",
			Coefficient: 1.0,
			Hours:       6,
		},
		{
			StartTime:   "06:00",
			Coefficient: 1.2,
			Hours:       6,
		},
		{
			StartTime:   "12:00",
			Coefficient: 1.0,
			Hours:       6,
		},
		{
			StartTime:   "18:00",
			Coefficient: 0.8,
			Hours:       6,
		},
	}
}
