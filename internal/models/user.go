package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a diabetes app user
type User struct {
	ID                 primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID             string              `json:"userId" bson:"userId"`
	Settings           Settings            `json:"settings" bson:"settings"`
	BloodSugarReadings []BloodSugarReading `json:"bloodSugarReadings" bson:"bloodSugarReadings"`
}

// BloodSugarReading represents a blood sugar reading
type BloodSugarReading struct {
	Value     float64   `json:"value" bson:"value"`                       // Blood sugar value in mmol/L
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`               // When the reading was taken
	Source    string    `json:"source,omitempty" bson:"source,omitempty"` // Optional source of reading
}

// FoodAnalysis represents the result of food analysis
type FoodAnalysis struct {
	Dish              string  `json:"dish"`
	Carbs             float64 `json:"carbs"`
	Confidence        string  `json:"confidence"`
	MealInsulin       float64 `json:"mealInsulin"`
	CorrectionInsulin float64 `json:"correctionInsulin"`
	TotalInsulin      float64 `json:"totalInsulin"`
	PeriodCoefficient float64 `json:"periodCoefficient"`
}
