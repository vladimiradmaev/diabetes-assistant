package insulin

import (
	"math"
	"time"

	"github.com/yourusername/diabetes-assistant/internal/models"
)

// CalculateMealInsulin calculates the insulin dose for a meal based on carbohydrates
// and the user's insulin-to-carb ratio
func CalculateMealInsulin(carbGrams, carbRatio, timeCoefficient float64) float64 {
	// Basic calculation: carbs / carb ratio
	baseInsulin := carbGrams / carbRatio

	// Apply the time-specific coefficient
	return baseInsulin * timeCoefficient
}

// CalculateCorrectionInsulin calculates the correction insulin for high blood sugar
func CalculateCorrectionInsulin(currentBG, targetBG, sensitivityFactor float64) float64 {
	bgDifference := currentBG - targetBG
	return bgDifference / sensitivityFactor
}

// CalculateTotalInsulin calculates the total insulin dose (meal + correction)
func CalculateTotalInsulin(mealInsulin, correctionInsulin, insulinOnBoard float64) float64 {
	totalDose := mealInsulin + correctionInsulin - insulinOnBoard
	// Don't return negative insulin doses
	return math.Max(0, totalDose)
}

// CalculateSensitivityFactor calculates the insulin sensitivity factor based on total daily insulin
// This uses the "1800 rule" for mmol/L
func CalculateSensitivityFactor(totalDailyInsulin float64) float64 {
	// 1800 rule for mmol/L (100 rule for mg/dL divided by 18)
	return 100 / totalDailyInsulin / 18
}

// AdjustInsulinCoefficients adjusts insulin coefficients based on blood sugar readings
func AdjustInsulinCoefficients(readings []models.BloodSugarReading, targetBG float64, currentCoefficients models.BaseInsulinCoefficients) models.BaseInsulinCoefficients {
	if len(readings) < 5 {
		// Not enough data to make adjustments
		return currentCoefficients
	}

	// Group readings by time of day
	morningReadings := []float64{}
	afternoonReadings := []float64{}
	eveningReadings := []float64{}
	nightReadings := []float64{}

	for _, reading := range readings {
		hour := reading.Timestamp.Hour()

		if hour >= 6 && hour < 12 {
			morningReadings = append(morningReadings, reading.Value)
		} else if hour >= 12 && hour < 18 {
			afternoonReadings = append(afternoonReadings, reading.Value)
		} else if hour >= 18 && hour < 22 {
			eveningReadings = append(eveningReadings, reading.Value)
		} else {
			nightReadings = append(nightReadings, reading.Value)
		}
	}

	// Calculate average blood sugar for each time of day
	calculatedCoefficients := currentCoefficients

	if len(morningReadings) >= 3 {
		avgMorningBG := calculateAverage(morningReadings)
		calculatedCoefficients.Morning = adjustCoefficient(avgMorningBG, targetBG, currentCoefficients.Morning)
	}

	if len(afternoonReadings) >= 3 {
		avgAfternoonBG := calculateAverage(afternoonReadings)
		calculatedCoefficients.Afternoon = adjustCoefficient(avgAfternoonBG, targetBG, currentCoefficients.Afternoon)
	}

	if len(eveningReadings) >= 3 {
		avgEveningBG := calculateAverage(eveningReadings)
		calculatedCoefficients.Evening = adjustCoefficient(avgEveningBG, targetBG, currentCoefficients.Evening)
	}

	if len(nightReadings) >= 3 {
		avgNightBG := calculateAverage(nightReadings)
		calculatedCoefficients.Night = adjustCoefficient(avgNightBG, targetBG, currentCoefficients.Night)
	}

	return calculatedCoefficients
}

// GetTimeBasedCoefficient gets the appropriate coefficient based on the current time
func GetTimeBasedCoefficient(currentTime time.Time, coefficients models.BaseInsulinCoefficients) float64 {
	hour := currentTime.Hour()

	if hour >= 6 && hour < 12 {
		return coefficients.Morning
	} else if hour >= 12 && hour < 18 {
		return coefficients.Afternoon
	} else if hour >= 18 && hour < 22 {
		return coefficients.Evening
	} else {
		return coefficients.Night
	}
}

// Helper function to adjust an individual coefficient
func adjustCoefficient(avgBG, targetBG, currentCoefficient float64) float64 {
	if math.Abs(avgBG-targetBG) < 1.0 {
		// Blood sugar is close enough to target, no adjustment needed
		return currentCoefficient
	}

	// Calculate adjustment factor
	adjustmentFactor := targetBG / avgBG

	// Limit adjustment to 20% at a time to avoid drastic changes
	limitedAdjustment := math.Max(0.8, math.Min(1.2, adjustmentFactor))

	// Apply adjustment to current coefficient
	return currentCoefficient * limitedAdjustment
}

// Helper function to calculate average of float64 slice
func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}

	return sum / float64(len(values))
}
