package storage

import (
	"context"
	"time"

	"github.com/yourusername/diabetes-assistant/internal/models"
)

// Storage defines the interface for data storage operations
type Storage interface {
	// User settings
	GetUserSettings(ctx context.Context, userID string) (*models.Settings, error)
	SaveUserSettings(ctx context.Context, settings *models.Settings) error

	// Blood sugar readings
	SaveBloodSugarReading(ctx context.Context, reading *models.BloodSugarReading) error
	GetBloodSugarReadings(ctx context.Context, userID string) ([]*models.BloodSugarReading, error)
	DeleteBloodSugarReading(ctx context.Context, userID string, timestamp string) error

	// User operations
	GetUser(userID string) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	UpdateUserSettings(userID string, settings models.Settings) error

	// Blood sugar readings operations
	AddBloodSugarReading(userID string, reading models.BloodSugarReading) error
	GetRecentBloodSugarReadings(userID string, limit int, startDate time.Time) ([]models.BloodSugarReading, error)

	// Close connection if needed
	Close() error
}
