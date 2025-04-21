package models

import "time"

type TargetBloodSugarRange struct {
	Min float64 `json:"min" bson:"min"`
	Max float64 `json:"max" bson:"max"`
}

type InsulinPeriod struct {
	StartTime   string  `json:"startTime" bson:"startTime"`
	Coefficient float64 `json:"coefficient" bson:"coefficient"`
	Hours       float64 `json:"hours" bson:"hours"`
}

type SensitivityPeriod struct {
	StartTime   string  `json:"startTime" bson:"startTime"`
	Sensitivity float64 `json:"sensitivity" bson:"sensitivity"`
	Hours       float64 `json:"hours" bson:"hours"`
}

type CarbRatioPeriod struct {
	StartTime string  `json:"startTime" bson:"startTime"`
	Ratio     float64 `json:"ratio" bson:"ratio"`
	Hours     float64 `json:"hours" bson:"hours"`
}

// Settings represents user-specific settings for the diabetes assistant
type Settings struct {
	ID string `json:"id" bson:"_id,omitempty"`
	// User ID
	UserID string `json:"userId" bson:"userId"`
	// Target blood sugar range
	TargetMin float64 `json:"targetMin" bson:"targetMin"`
	TargetMax float64 `json:"targetMax" bson:"targetMax"`
	// Insulin on board duration
	IOBDuration float64 `json:"iobDuration" bson:"iobDuration"`
	// Insulin periods
	InsulinPeriods []InsulinPeriod `json:"insulinPeriods" bson:"insulinPeriods"`
	// Sensitivity periods
	SensitivityPeriods []SensitivityPeriod `json:"sensitivityPeriods" bson:"sensitivityPeriods"`
	// Carb ratio periods
	CarbRatioPeriods []CarbRatioPeriod `json:"carbRatioPeriods" bson:"carbRatioPeriods"`
	// Timestamp when settings were last updated
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

// CreateDefaultSettings creates a new settings object with default values
func CreateDefaultSettings(userID string) *Settings {
	return &Settings{
		UserID:      userID,
		TargetMin:   4.0,
		TargetMax:   8.0,
		IOBDuration: 4.0,
		InsulinPeriods: []InsulinPeriod{
			{StartTime: "00:00", Coefficient: 1.0, Hours: 24},
		},
		SensitivityPeriods: []SensitivityPeriod{
			{StartTime: "00:00", Sensitivity: 2.0, Hours: 24},
		},
		CarbRatioPeriods: []CarbRatioPeriod{
			{StartTime: "00:00", Ratio: 1.0, Hours: 24},
		},
		UpdatedAt: time.Now(),
	}
}
