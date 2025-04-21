package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/yourusername/diabetes-assistant/internal/models"
)

// InMemoryStorage implements Storage interface with an in-memory map
type InMemoryStorage struct {
	users map[string]*models.User
	mu    sync.RWMutex
}

// NewInMemoryStorage creates a new in-memory storage
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		users: make(map[string]*models.User),
	}
}

// GetUser retrieves a user by ID
func (s *InMemoryStorage) GetUser(userID string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[userID]
	if !exists {
		return nil, nil // User not found, but not an error
	}
	return user, nil
}

// CreateUser creates a new user
func (s *InMemoryStorage) CreateUser(user *models.User) error {
	if user.UserID == "" {
		return errors.New("user ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user already exists
	if _, exists := s.users[user.UserID]; exists {
		return errors.New("user already exists")
	}

	// Create a deep copy to avoid reference issues
	s.users[user.UserID] = &models.User{
		UserID:             user.UserID,
		Settings:           user.Settings,
		BloodSugarReadings: make([]models.BloodSugarReading, 0),
	}

	return nil
}

// UpdateUser updates an existing user
func (s *InMemoryStorage) UpdateUser(user *models.User) error {
	if user.UserID == "" {
		return errors.New("user ID is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user exists
	if _, exists := s.users[user.UserID]; !exists {
		return errors.New("user not found")
	}

	// Store the user (make a copy to avoid reference issues)
	userCopy := *user
	s.users[user.UserID] = &userCopy

	return nil
}

// UpdateUserSettings updates a user's settings
func (s *InMemoryStorage) UpdateUserSettings(userID string, settings models.Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	// Ensure settings are valid
	settingsCopy := settings
	ensureValidSettingsMemory(&settingsCopy)

	user.Settings = settingsCopy
	return nil
}

// AddBloodSugarReading adds a blood sugar reading to a user
func (s *InMemoryStorage) AddBloodSugarReading(userID string, reading models.BloodSugarReading) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	// Add reading at the beginning of the slice (newest first)
	user.BloodSugarReadings = append([]models.BloodSugarReading{reading}, user.BloodSugarReadings...)
	return nil
}

// GetRecentBloodSugarReadings gets recent blood sugar readings for a user
func (s *InMemoryStorage) GetRecentBloodSugarReadings(userID string, limit int, startDate time.Time) ([]models.BloodSugarReading, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}

	// Filter readings by date
	var filteredReadings []models.BloodSugarReading
	for _, reading := range user.BloodSugarReadings {
		if reading.Timestamp.After(startDate) || reading.Timestamp.Equal(startDate) {
			filteredReadings = append(filteredReadings, reading)
		}
	}

	// Apply limit if needed
	if limit > 0 && len(filteredReadings) > limit {
		filteredReadings = filteredReadings[:limit]
	}

	return filteredReadings, nil
}

// SaveUserSettings saves user settings, creating a user if they don't exist
func (s *InMemoryStorage) SaveUserSettings(ctx context.Context, settings *models.Settings) error {
	if settings.UserID == "" {
		return errors.New("user ID is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user exists
	user, exists := s.users[settings.UserID]
	if !exists {
		// Create new user with these settings
		newUser := &models.User{
			UserID:             settings.UserID,
			Settings:           *settings,
			BloodSugarReadings: make([]models.BloodSugarReading, 0),
		}
		s.users[settings.UserID] = newUser
		return nil
	}

	// Update existing user's settings
	user.Settings = *settings
	return nil
}

// GetUserSettings returns the user's settings
func (s *InMemoryStorage) GetUserSettings(ctx context.Context, userID string) (*models.Settings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Return a copy of the settings to prevent modification
	settings := user.Settings
	return &settings, nil
}

// Helper function to ensure settings are valid
func ensureValidSettingsMemory(settings *models.Settings) {
	// Make sure periods are properly initialized
	if len(settings.InsulinPeriods) == 0 {
		// If no periods, create default ones
		defaultSettings := models.CreateDefaultSettings(settings.UserID)
		settings.InsulinPeriods = defaultSettings.InsulinPeriods
	}

	if len(settings.SensitivityPeriods) == 0 {
		defaultSettings := models.CreateDefaultSettings(settings.UserID)
		settings.SensitivityPeriods = defaultSettings.SensitivityPeriods
	}

	if len(settings.CarbRatioPeriods) == 0 {
		defaultSettings := models.CreateDefaultSettings(settings.UserID)
		settings.CarbRatioPeriods = defaultSettings.CarbRatioPeriods
	}

	// Make sure UpdatedAt is set
	if settings.UpdatedAt.IsZero() {
		settings.UpdatedAt = time.Now()
	}

	// Set default target range if not specified
	if settings.TargetMin == 0 && settings.TargetMax == 0 {
		settings.TargetMin = 4.0
		settings.TargetMax = 8.0
	}

	// Set default IOB duration if not specified
	if settings.IOBDuration == 0 {
		settings.IOBDuration = 4.0
	}
}

// DeleteBloodSugarReading deletes a blood sugar reading for a user by timestamp
func (s *InMemoryStorage) DeleteBloodSugarReading(ctx context.Context, userID string, timestamp string) error {
	// Convert string timestamp to time.Time
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return fmt.Errorf("invalid timestamp format: %v", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	// Find and remove the reading with matching timestamp
	found := false
	for i, reading := range user.BloodSugarReadings {
		if reading.Timestamp.Equal(t) {
			// Remove the reading
			user.BloodSugarReadings = append(user.BloodSugarReadings[:i], user.BloodSugarReadings[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return errors.New("no reading found with the specified timestamp")
	}

	return nil
}

// Close is a no-op for in-memory storage
func (s *InMemoryStorage) Close() error {
	return nil
}

// GetBloodSugarReadings retrieves all blood sugar readings for a user
func (s *InMemoryStorage) GetBloodSugarReadings(ctx context.Context, userID string) ([]*models.BloodSugarReading, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}

	// Convert to pointers
	readings := make([]*models.BloodSugarReading, len(user.BloodSugarReadings))
	for i := range user.BloodSugarReadings {
		readings[i] = &user.BloodSugarReadings[i]
	}

	return readings, nil
}

// SaveBloodSugarReading saves a blood sugar reading
func (s *InMemoryStorage) SaveBloodSugarReading(ctx context.Context, reading *models.BloodSugarReading) error {
	// Since reading doesn't have UserID, we need to get it from the context
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return errors.New("userID not found in context")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	// Add reading at the beginning of the slice (newest first)
	user.BloodSugarReadings = append([]models.BloodSugarReading{*reading}, user.BloodSugarReadings...)
	return nil
}
