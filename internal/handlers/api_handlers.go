package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/diabetes-assistant/internal/models"
	"github.com/yourusername/diabetes-assistant/internal/services/ai"
	"github.com/yourusername/diabetes-assistant/internal/services/insulin"
	"github.com/yourusername/diabetes-assistant/internal/services/libre"
	"github.com/yourusername/diabetes-assistant/internal/storage"
)

// APIHandler handles API requests
type APIHandler struct {
	storage    storage.Storage
	ai         *ai.Service
	libre      *libre.LibreService
	uploadsDir string
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(storage storage.Storage, aiService *ai.Service, libreService *libre.LibreService, uploadsDir string) *APIHandler {
	return &APIHandler{
		storage:    storage,
		ai:         aiService,
		libre:      libreService,
		uploadsDir: uploadsDir,
	}
}

// HealthCheck handles health check requests
func (h *APIHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GetUserSettings handles GET /api/settings/:userId
func (h *APIHandler) GetUserSettings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]

	user, err := h.storage.GetUser(userId)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Error fetching user: %v", err))
		return
	}

	if user == nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"settings": user.Settings})
}

// SaveUserSettings handles saving user settings
func (h *APIHandler) SaveUserSettings(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get user ID from URL path
	vars := mux.Vars(r)
	userID := vars["userId"]
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Parse JSON body
	var settings models.Settings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		log.Printf("Error parsing settings: %v", err)
		http.Error(w, "Invalid settings data", http.StatusBadRequest)
		return
	}

	// Set user ID and update timestamp
	settings.UserID = userID
	settings.UpdatedAt = time.Now()

	// Validate required fields
	if settings.TargetMin == 0 || settings.TargetMax == 0 || settings.IOBDuration == 0 {
		http.Error(w, "Target blood sugar range and IOB duration are required", http.StatusBadRequest)
		return
	}

	// Validate periods
	if len(settings.InsulinPeriods) == 0 || len(settings.SensitivityPeriods) == 0 || len(settings.CarbRatioPeriods) == 0 {
		http.Error(w, "At least one period is required for each type", http.StatusBadRequest)
		return
	}

	// Create context
	ctx := context.Background()

	// Save settings
	if err := h.storage.SaveUserSettings(ctx, &settings); err != nil {
		log.Printf("Error saving settings: %v", err)
		http.Error(w, "Error saving settings", http.StatusInternalServerError)
		return
	}

	// Return success response
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Settings saved successfully",
	})
}

// SaveBloodSugar handles POST /api/bloodsugar
func (h *APIHandler) SaveBloodSugar(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string  `json:"userId"`
		Value  float64 `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	if req.UserID == "" {
		respondError(w, http.StatusBadRequest, "Missing user ID")
		return
	}

	if req.Value <= 0 {
		respondError(w, http.StatusBadRequest, "Invalid blood sugar value")
		return
	}

	user, err := h.storage.GetUser(req.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Error fetching user: %v", err))
		return
	}

	if user == nil {
		// Create new user with default settings
		defaultSettings := models.CreateDefaultSettings(req.UserID)
		user = &models.User{
			UserID:             req.UserID,
			Settings:           *defaultSettings,
			BloodSugarReadings: []models.BloodSugarReading{}, // Initialize as empty array
		}
		if err := h.storage.CreateUser(user); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating user: %v", err))
			return
		}
	}

	// Create reading
	reading := models.BloodSugarReading{
		Value:     req.Value,
		Timestamp: time.Now(),
	}

	// Save reading
	if err := h.storage.AddBloodSugarReading(req.UserID, reading); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Error saving reading: %v", err))
		return
	}

	// Determine status
	status := "Normal range"
	if req.Value < 3.9 {
		status = "Low blood sugar (hypoglycemia)"
	} else if req.Value > 10.0 {
		status = "High blood sugar (hyperglycemia)"
	} else if req.Value > 7.0 {
		status = "Slightly elevated"
	}

	// Analyze readings and potentially adjust coefficients
	coefficientsAdjusted := false

	// Get recent readings (last week)
	oneWeekAgo := time.Now().AddDate(0, 0, -7)
	recentReadings, err := h.storage.GetRecentBloodSugarReadings(req.UserID, 0, oneWeekAgo)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Error fetching readings: %v", err))
		return
	}

	if len(recentReadings) >= 5 && user.Settings.TargetMin > 0 {
		// Analyze readings for potential adjustments
		// First convert the new-style insulin periods to the legacy format for the calculator
		legacyCoefficients := models.InsulinPeriodsToBaseCoefficients(user.Settings.InsulinPeriods)

		// Use the insulin calculator to adjust coefficients (returns legacy format)
		adjustedLegacyCoefficients := insulin.AdjustInsulinCoefficients(
			recentReadings,
			user.Settings.TargetMin,
			legacyCoefficients,
		)

		// Convert back to the new format
		adjustedPeriods := adjustedLegacyCoefficients.ConvertToInsulinPeriods()

		// Only update if coefficients actually changed
		if !areInsulinPeriodsEqual(user.Settings.InsulinPeriods, adjustedPeriods) {
			user.Settings.InsulinPeriods = adjustedPeriods
			if err := h.storage.UpdateUserSettings(req.UserID, user.Settings); err != nil {
				respondError(w, http.StatusInternalServerError, fmt.Sprintf("Error updating settings: %v", err))
				return
			}

			coefficientsAdjusted = true
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"reading": map[string]interface{}{
			"value":     req.Value,
			"status":    status,
			"timestamp": reading.Timestamp,
		},
		"coefficientsAdjusted": coefficientsAdjusted,
		"targetLevel":          user.Settings.TargetMin,
	})
}

// Helper to compare insulin periods
func areInsulinPeriodsEqual(a, b []models.InsulinPeriod) bool {
	if len(a) != len(b) {
		return false
	}

	// This is a simple comparison that assumes the periods are in the same order
	for i := range a {
		if a[i].StartTime != b[i].StartTime ||
			a[i].Hours != b[i].Hours ||
			a[i].Coefficient != b[i].Coefficient {
			return false
		}
	}

	return true
}

// GetBloodSugarReadings handles GET /api/bloodsugar/:userId
func (h *APIHandler) GetBloodSugarReadings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	startDateStr := r.URL.Query().Get("startDate")
	// Parse endDateStr if we need to filter by end date in the future
	// endDateStr := r.URL.Query().Get("endDate")

	limit := 0
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid limit parameter")
			return
		}
	}

	// Default to 1 week ago if no start date provided
	startDate := time.Now().AddDate(0, 0, -7)
	if startDateStr != "" {
		var err error
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid startDate parameter")
			return
		}
	}

	readings, err := h.storage.GetRecentBloodSugarReadings(userId, limit, startDate)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Error fetching readings: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"readings": readings})
}

// AnalyzeFood handles POST /api/analyze-food
func (h *APIHandler) AnalyzeFood(w http.ResponseWriter, r *http.Request) {
	// Log request
	fmt.Printf("AnalyzeFood: Received %s request with content type: %s\n", r.Method, r.Header.Get("Content-Type"))

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Set CORS headers for all responses
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	var userId string
	var photoProvided bool
	var foodPhotoPath string
	var foodWeight float64 // Weight in grams

	// Parse the multipart form first (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		fmt.Printf("AnalyzeFood: Not a multipart form: %v\n", err)
		// Not a multipart form - try to parse as JSON
		var req struct {
			UserID     string  `json:"userId"`
			FoodWeight float64 `json:"foodWeight,omitempty"` // Optional weight in grams
		}

		bodyBytes, _ := io.ReadAll(r.Body)
		if len(bodyBytes) > 0 {
			fmt.Printf("AnalyzeFood: Raw request body: %s\n", string(bodyBytes))
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				fmt.Printf("AnalyzeFood: Error decoding JSON: %v\n", err)
				respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %v", err))
				return
			}

			userId = req.UserID
			foodWeight = req.FoodWeight

			// Photo is required
			respondError(w, http.StatusBadRequest, "Food photo is required for analysis")
			return
		} else {
			fmt.Println("AnalyzeFood: Empty request body")
			respondError(w, http.StatusBadRequest, "Empty request body")
			return
		}
	} else {
		// Successfully parsed multipart form
		userId = r.FormValue("userId")

		// Parse weight if provided
		weightStr := r.FormValue("foodWeight")
		if weightStr != "" {
			var err error
			foodWeight, err = strconv.ParseFloat(weightStr, 64)
			if err != nil {
				fmt.Printf("AnalyzeFood: Invalid food weight: %v\n", err)
				// Not a critical error, continue with weight=0
				foodWeight = 0
			}
		}

		fmt.Printf("AnalyzeFood: Multipart form values - userId: %s, foodWeight: %.1f\n",
			userId, foodWeight)

		// Check for photo (required)
		foodPhoto, foodPhotoHeader, err := r.FormFile("foodPhoto")
		if err == nil && foodPhoto != nil {
			defer foodPhoto.Close()
			photoProvided = true
			fmt.Printf("AnalyzeFood: Photo provided - filename: %s, size: %d\n", foodPhotoHeader.Filename, foodPhotoHeader.Size)

			// Save photo to disk
			foodPhotoFileName := fmt.Sprintf("food_%s%s", uuid.New().String(), filepath.Ext(foodPhotoHeader.Filename))
			foodPhotoPath = filepath.Join(h.uploadsDir, foodPhotoFileName)

			foodPhotoFile, err := os.Create(foodPhotoPath)
			if err != nil {
				fmt.Printf("AnalyzeFood: Error creating file: %v\n", err)
				respondError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating file: %v", err))
				return
			}
			defer foodPhotoFile.Close()

			if _, err := io.Copy(foodPhotoFile, foodPhoto); err != nil {
				fmt.Printf("AnalyzeFood: Error saving file: %v\n", err)
				respondError(w, http.StatusInternalServerError, fmt.Sprintf("Error saving file: %v", err))
				return
			}
			fmt.Printf("AnalyzeFood: Photo saved to %s\n", foodPhotoPath)
		} else {
			fmt.Printf("AnalyzeFood: No photo in request or error: %v\n", err)
			respondError(w, http.StatusBadRequest, "Food photo is required for analysis")
			return
		}
	}

	// Get user settings for insulin calculations
	user, err := h.storage.GetUser(userId)
	if err != nil {
		fmt.Printf("AnalyzeFood: Error fetching user: %v\n", err)
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Error fetching user: %v", err))
		return
	}

	// Use default settings if user not found
	var userSettings *models.Settings
	if user != nil {
		// Make a copy of the settings
		settingsCopy := user.Settings
		userSettings = &settingsCopy
	} else {
		// Create default settings
		userSettings = models.CreateDefaultSettings(userId)
	}

	// Analyze food using AI service - we're passing empty string as the description parameter
	var foodAnalysisResult *ai.FoodAnalysisResult
	if photoProvided && foodPhotoPath != "" {
		// Analyze with photo and optional weight
		foodAnalysisResult, err = h.ai.AnalyzeFood(foodPhotoPath, "", foodWeight)
		if err != nil {
			fmt.Printf("AnalyzeFood: AI analysis error: %v\n", err)
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("Error analyzing food: %v", err))
			return
		}
	} else {
		// Photo is required
		respondError(w, http.StatusBadRequest, "Food photo is required for analysis")
		return
	}

	// Calculate insulin dose based on carbs and user settings
	mealInsulin := foodAnalysisResult.Carbs / userSettings.CarbRatioPeriods[0].Ratio

	// Get current time to determine time-based coefficient
	hour := time.Now().Hour()
	periodCoefficient := 1.0

	// Apply time-based coefficient
	if userSettings.InsulinPeriods != nil && len(userSettings.InsulinPeriods) > 0 {
		// Use the new array format
		for _, period := range userSettings.InsulinPeriods {
			startHour, _ := strconv.Atoi(strings.Split(period.StartTime, ":")[0])
			if hour >= startHour && hour < startHour+int(period.Hours) {
				periodCoefficient = period.Coefficient
				break
			}
		}
	}

	// Apply coefficient
	mealInsulin *= periodCoefficient

	// Add correction insulin if needed
	lastReading, _ := h.storage.GetRecentBloodSugarReadings(userId, 1, time.Now().AddDate(0, 0, -7))

	correctionInsulin := 0.0
	if len(lastReading) > 0 && userSettings.TargetMin > 0 && userSettings.SensitivityPeriods[0].Sensitivity > 0 {
		bloodSugarDiff := lastReading[0].Value - userSettings.TargetMin
		if bloodSugarDiff > 0 {
			correctionInsulin = bloodSugarDiff / userSettings.SensitivityPeriods[0].Sensitivity
		}
	}

	// Calculate total insulin
	totalInsulin := mealInsulin + correctionInsulin

	// Send the results
	response := map[string]interface{}{
		"success":       true,
		"detectedFood":  foodAnalysisResult.Name,
		"carbs":         foodAnalysisResult.Carbs,
		"insulinDose":   totalInsulin,
		"reasoning":     foodAnalysisResult.Reasoning,
		"photoProvided": photoProvided,
		"analysis": map[string]interface{}{
			"dish":              foodAnalysisResult.Name,
			"carbs":             foodAnalysisResult.Carbs,
			"confidence":        foodAnalysisResult.Confidence,
			"reasoning":         foodAnalysisResult.Reasoning,
			"mealInsulin":       mealInsulin,
			"correctionInsulin": correctionInsulin,
			"totalInsulin":      totalInsulin,
			"periodCoefficient": periodCoefficient,
		},
	}

	fmt.Printf("AnalyzeFood: Sending response: %+v\n", response)
	respondJSON(w, http.StatusOK, response)
}

// SyncLibre handles POST /api/sync-libre
func (h *APIHandler) SyncLibre(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"userId"`
		Method string `json:"method"`
		Value  string `json:"value,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	if req.UserID == "" || req.Method == "" {
		respondError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	user, err := h.storage.GetUser(req.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Error fetching user: %v", err))
		return
	}

	if user == nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	// Only manual entry is supported for now
	if req.Method != "manual" {
		respondError(w, http.StatusBadRequest, "Only manual entry is supported")
		return
	}

	if req.Value == "" {
		respondError(w, http.StatusBadRequest, "Blood sugar value is required for manual entry")
		return
	}

	value, err := strconv.ParseFloat(req.Value, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid blood sugar value")
		return
	}

	reading := models.BloodSugarReading{
		Value:     value,
		Timestamp: time.Now(),
	}

	if err := h.storage.AddBloodSugarReading(req.UserID, reading); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Error saving reading: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"reading": reading,
	})
}

// DeleteBloodSugar handles DELETE /api/bloodsugar
func (h *APIHandler) DeleteBloodSugar(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req struct {
		UserID    string `json:"userId"`
		Timestamp string `json:"timestamp"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Validate required fields
	if req.UserID == "" {
		respondError(w, http.StatusBadRequest, "Missing user ID")
		return
	}

	if req.Timestamp == "" {
		respondError(w, http.StatusBadRequest, "Missing or invalid timestamp")
		return
	}

	// Create context
	ctx := context.Background()

	// Delete the reading
	err := h.storage.DeleteBloodSugarReading(ctx, req.UserID, req.Timestamp)
	if err != nil {
		if err.Error() == "user not found" {
			respondError(w, http.StatusNotFound, "User not found")
			return
		}
		if err.Error() == "no reading found with the specified timestamp" {
			respondError(w, http.StatusNotFound, "Blood sugar reading not found")
			return
		}
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Error deleting reading: %v", err))
		return
	}

	// Return success
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Blood sugar reading successfully deleted",
	})
}

// Helper function to respond with JSON
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

// Helper function to respond with error
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
