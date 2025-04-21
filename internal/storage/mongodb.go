package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/yourusername/diabetes-assistant/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBStorage implements the Storage interface for MongoDB
type MongoDBStorage struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

// Check that MongoDBStorage implements the Storage interface
var _ Storage = (*MongoDBStorage)(nil)

// NewMongoDBStorage creates a new MongoDB storage instance
func NewMongoDBStorage(uri string) (*MongoDBStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Ping the database to check connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	database := client.Database("diabetes-assistant")
	collection := database.Collection("users")

	return &MongoDBStorage{
		client:     client,
		database:   database,
		collection: collection,
	}, nil
}

// Close closes the MongoDB connection
func (s *MongoDBStorage) Close() error {
	return s.client.Disconnect(context.Background())
}

// GetUser retrieves a user by ID
func (s *MongoDBStorage) GetUser(userID string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := s.collection.FindOne(ctx, bson.M{"userId": userID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a new user
func (s *MongoDBStorage) CreateUser(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.collection.InsertOne(ctx, user)
	return err
}

// UpdateUser updates an existing user
func (s *MongoDBStorage) UpdateUser(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{"userId": user.UserID},
		bson.M{"$set": user},
	)
	return err
}

// UpdateUserSettings updates user settings
func (s *MongoDBStorage) UpdateUserSettings(userID string, settings models.Settings) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	settings.UpdatedAt = time.Now()
	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{"userId": userID},
		bson.M{"$set": settings},
	)
	return err
}

// AddBloodSugarReading adds a new blood sugar reading
func (s *MongoDBStorage) AddBloodSugarReading(userID string, reading models.BloodSugarReading) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{"userId": userID},
		bson.M{"$push": bson.M{"bloodSugarReadings": reading}},
		options.Update().SetUpsert(true),
	)
	return err
}

// GetRecentBloodSugarReadings gets recent blood sugar readings
func (s *MongoDBStorage) GetRecentBloodSugarReadings(userID string, limit int, startDate time.Time) ([]models.BloodSugarReading, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user struct {
		BloodSugarReadings []models.BloodSugarReading `bson:"bloodSugarReadings"`
	}
	err := s.collection.FindOne(ctx, bson.M{"userId": userID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	// Filter readings by date and limit
	var filteredReadings []models.BloodSugarReading
	for _, reading := range user.BloodSugarReadings {
		if reading.Timestamp.After(startDate) {
			filteredReadings = append(filteredReadings, reading)
		}
	}

	if limit > 0 && len(filteredReadings) > limit {
		filteredReadings = filteredReadings[:limit]
	}

	return filteredReadings, nil
}

// GetUserSettings retrieves user settings from the database
func (s *MongoDBStorage) GetUserSettings(ctx context.Context, userID string) (*models.Settings, error) {
	var settings models.Settings
	err := s.collection.FindOne(ctx, bson.M{"userId": userID}).Decode(&settings)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &settings, nil
}

// ensureValidSettingsMongo ensures all required fields are set and valid
func ensureValidSettingsMongo(settings *models.Settings) {
	if settings.TargetMin == 0 {
		settings.TargetMin = 4.0
	}
	if settings.TargetMax == 0 {
		settings.TargetMax = 8.0
	}
	if settings.IOBDuration == 0 {
		settings.IOBDuration = 4.0
	}
	if len(settings.CarbRatioPeriods) == 0 {
		settings.CarbRatioPeriods = []models.CarbRatioPeriod{
			{
				StartTime: "00:00",
				Ratio:     1.0,
				Hours:     24,
			},
		}
	}
	if len(settings.SensitivityPeriods) == 0 {
		settings.SensitivityPeriods = []models.SensitivityPeriod{
			{
				StartTime:   "00:00",
				Sensitivity: 2.0,
				Hours:       24,
			},
		}
	}
	if len(settings.InsulinPeriods) == 0 {
		settings.InsulinPeriods = []models.InsulinPeriod{
			{
				StartTime:   "00:00",
				Coefficient: 1.0,
				Hours:       24,
			},
		}
	}
}

// SaveUserSettings saves user settings to the database
func (s *MongoDBStorage) SaveUserSettings(ctx context.Context, settings *models.Settings) error {
	if settings.UserID == "" {
		return errors.New("user ID is required")
	}

	// Ensure settings are valid
	ensureValidSettingsMongo(settings)
	settings.UpdatedAt = time.Now()

	// Create or update user document
	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{"userId": settings.UserID},
		bson.M{"$set": settings},
		options.Update().SetUpsert(true),
	)
	return err
}

// SaveBloodSugarReading saves a blood sugar reading to the database
func (s *MongoDBStorage) SaveBloodSugarReading(ctx context.Context, reading *models.BloodSugarReading) error {
	// Since reading doesn't have UserID, we need to get it from the context
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return errors.New("userID not found in context")
	}

	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{"userId": userID},
		bson.M{"$push": bson.M{"bloodSugarReadings": reading}},
		options.Update().SetUpsert(true),
	)
	return err
}

// GetBloodSugarReadings retrieves all blood sugar readings for a user
func (s *MongoDBStorage) GetBloodSugarReadings(ctx context.Context, userID string) ([]*models.BloodSugarReading, error) {
	var user struct {
		BloodSugarReadings []*models.BloodSugarReading `bson:"bloodSugarReadings"`
	}
	err := s.collection.FindOne(ctx, bson.M{"userId": userID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return user.BloodSugarReadings, nil
}

// DeleteBloodSugarReading deletes a specific blood sugar reading
func (s *MongoDBStorage) DeleteBloodSugarReading(ctx context.Context, userID string, timestamp string) error {
	// Convert string timestamp to time.Time
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return fmt.Errorf("invalid timestamp format: %v", err)
	}

	// First check if user exists
	user, err := s.GetUser(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Update the user document to remove the reading with matching timestamp
	result, err := s.collection.UpdateOne(
		ctx,
		bson.M{"userId": userID},
		bson.M{"$pull": bson.M{"bloodSugarReadings": bson.M{"timestamp": t}}},
	)
	if err != nil {
		return err
	}

	// Check if any document was modified
	if result.ModifiedCount == 0 {
		return errors.New("no reading found with the specified timestamp")
	}

	return nil
}
