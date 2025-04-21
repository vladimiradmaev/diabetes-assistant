package ai

import (
	"errors"
	"fmt"
	"log"

	"github.com/yourusername/diabetes-assistant/internal/config"
)

// FoodAnalysisResult represents the result of food analysis
type FoodAnalysisResult struct {
	Name       string  `json:"name"`
	Carbs      float64 `json:"carbs"`
	Confidence string  `json:"confidence"`
	Reasoning  string  `json:"reasoning"` // Explanation of how carbs were estimated
}

// Provider represents the interface that all AI providers must implement
type Provider interface {
	// AnalyzeFood analyzes a food image and returns estimated carbohydrates
	// The second parameter (originally for ingredient images) is deprecated and not used
	AnalyzeFood(foodImagePath, unusedDescriptionParam string, foodWeight float64) (*FoodAnalysisResult, error)
}

// Service is the main AI service that delegates to the appropriate provider
type Service struct {
	config       *config.Config
	provider     Provider
	providerName string // Stores which provider is being used
}

// NewService creates a new AI service
func NewService(cfg *config.Config) (*Service, error) {
	var provider Provider
	var providerName string
	var err error

	// Try to use OpenAI if the API key is available
	if cfg.OpenAIToken != "" {
		log.Println("Using OpenAI provider for AI analysis")
		provider, err = NewOpenAIProvider(cfg.OpenAIToken)
		providerName = "openai"
		if err != nil {
			log.Printf("Failed to initialize OpenAI provider: %v", err)
		}
	}

	// If OpenAI failed or not available, try Gemini
	if provider == nil && cfg.GeminiToken != "" {
		log.Println("Using Gemini provider for AI analysis")
		provider, err = NewGeminiProvider(cfg.GeminiToken)
		providerName = "gemini"
		if err != nil {
			log.Printf("Failed to initialize Gemini provider: %v", err)
		}
	}

	// If Gemini failed or not available, try Grok
	if provider == nil && cfg.GrokToken != "" {
		log.Println("Using Grok provider for AI analysis")
		provider, err = NewGrokProvider(cfg.GrokToken)
		providerName = "grok"
		if err != nil {
			log.Printf("Failed to initialize Grok provider: %v", err)
		}
	}

	// If all providers failed, fall back to mock
	if provider == nil {
		log.Println("No API keys provided or all providers failed to initialize. Using mock provider.")
		provider = &mockProvider{}
		providerName = "mock"
	}

	return &Service{
		config:       cfg,
		provider:     provider,
		providerName: providerName,
	}, nil
}

// AnalyzeFood analyzes a food image and returns the estimated carbohydrates
func (s *Service) AnalyzeFood(foodImagePath, ingredientImagePath string, foodWeight float64) (*FoodAnalysisResult, error) {
	// Note: ingredientImagePath parameter is deprecated and not used anymore, but kept for backward compatibility
	if s.provider == nil {
		return nil, errors.New("AI provider not initialized")
	}

	result, err := s.provider.AnalyzeFood(foodImagePath, ingredientImagePath, foodWeight)
	if err != nil {
		return nil, err
	}

	// Return the food analysis result
	return result, nil
}

// ChangeProvider changes the AI provider
func (s *Service) ChangeProvider(providerName string, key string) error {
	var provider Provider
	var err error

	switch providerName {
	case "openai":
		provider, err = NewOpenAIProvider(key)
	case "gemini":
		provider, err = NewGeminiProvider(key)
	case "grok":
		provider, err = NewGrokProvider(key)
	default:
		return fmt.Errorf("unsupported AI provider: %s", providerName)
	}

	if err != nil {
		return err
	}

	s.provider = provider
	s.providerName = providerName
	return nil
}

// GetCurrentProvider returns the name of the current AI provider
func (s *Service) GetCurrentProvider() string {
	return s.providerName
}

// mockProvider is a simple mock implementation of the Provider interface
type mockProvider struct{}

// AnalyzeFood implements the Provider interface for the mock provider
func (p *mockProvider) AnalyzeFood(foodImagePath, unusedDescriptionParam string, foodWeight float64) (*FoodAnalysisResult, error) {
	// Prepare response
	result := &FoodAnalysisResult{
		Name:       "Пицца",
		Carbs:      45.0,
		Confidence: "high",
		Reasoning:  "Это тестовый анализ для демонстрационных целей. Типичная пицца (среднего размера) содержит примерно 45г углеводов на кусок, в основном из-за теста.",
	}

	// Adjust carbs based on weight if provided
	if foodWeight > 0 {
		// Assuming standard slice is ~100g
		standardSliceWeight := 100.0
		weightRatio := foodWeight / standardSliceWeight
		result.Carbs = result.Carbs * weightRatio
		result.Reasoning = fmt.Sprintf("Это тестовый анализ для демонстрационных целей. Для указанного веса %.1f г пиццы (стандартный кусок ~100г) содержит примерно %.1fг углеводов.",
			foodWeight, result.Carbs)
	}

	return result, nil
}
