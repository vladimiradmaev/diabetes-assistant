package ai

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/yourusername/diabetes-assistant/internal/config"
)

// Provider defines which AI service to use
type Provider string

const (
	ProviderOpenAI Provider = "openai"
	ProviderGemini Provider = "gemini"
	ProviderGrok   Provider = "grok"
)

// FoodAnalysis represents the result of food analysis
type FoodAnalysis struct {
	Name       string  // Name of the dish
	Carbs      float64 // Carbohydrates in grams
	Confidence float64 // Confidence level (0-1)
}

// Service interface defines methods that all AI providers must implement
type Service interface {
	ChatCompletion(ctx context.Context, messages []Message) (string, error)
	AnalyzeFood(foodImagePath, ingredientImagePath string) (*FoodAnalysis, error)
	GetCurrentProvider() Provider
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// NewService creates an AI service based on available API keys
func NewService(cfg *config.Config) (Service, error) {
	// Try to create services in order of preference: Gemini, OpenAI, Grok
	var activeProvider Provider
	var service Service

	if cfg.GeminiToken != "" {
		geminiService, err := NewGeminiService(cfg.GeminiToken)
		if err != nil {
			log.Printf("Failed to initialize Gemini: %v", err)
		} else {
			log.Println("Using Gemini AI provider")
			service = geminiService
			activeProvider = ProviderGemini
		}
	}

	if service == nil && cfg.OpenAIToken != "" {
		openaiService, err := NewOpenAIService(cfg.OpenAIToken, cfg.DefaultModel)
		if err != nil {
			log.Printf("Failed to initialize OpenAI: %v", err)
		} else {
			log.Println("Using OpenAI provider")
			service = openaiService
			activeProvider = ProviderOpenAI
		}
	}

	if service == nil && cfg.GrokToken != "" {
		grokService, err := NewGrokService(cfg.GrokToken)
		if err != nil {
			log.Printf("Failed to initialize Grok: %v", err)
		} else {
			log.Println("Using Grok AI provider")
			service = grokService
			activeProvider = ProviderGrok
		}
	}

	if service == nil {
		return nil, errors.New("no AI service could be initialized - missing API keys")
	}

	return &aiService{
		provider: activeProvider,
		impl:     service,
	}, nil
}

// Wrapper service that keeps track of the active provider
type aiService struct {
	provider Provider
	impl     Service
}

func (s *aiService) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	return s.impl.ChatCompletion(ctx, messages)
}

func (s *aiService) AnalyzeFood(foodImagePath, ingredientImagePath string) (*FoodAnalysis, error) {
	return s.impl.AnalyzeFood(foodImagePath, ingredientImagePath)
}

func (s *aiService) GetCurrentProvider() Provider {
	return s.provider
}

// NewGeminiService creates a new Gemini service client
func NewGeminiService(apiKey string) (Service, error) {
	// For now we'll return a placeholder implementation
	// TODO: Implement actual Gemini API integration
	return &geminiService{
		apiKey: apiKey,
		model:  "gemini-1.5-flash", // Updated from gemini-pro-vision to gemini-1.5-flash
	}, nil
}

// geminiService implements the Service interface for Google's Gemini API
type geminiService struct {
	apiKey string
	model  string
}

func (s *geminiService) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	// TODO: Implement actual Gemini API call
	return fmt.Sprintf("Placeholder response from Gemini (model: %s, messages: %d)", s.model, len(messages)), nil
}

func (s *geminiService) AnalyzeFood(foodImagePath, ingredientImagePath string) (*FoodAnalysis, error) {
	// TODO: Implement actual Gemini API call for food analysis with image input
	// Using the updated model that supports multimodal inputs
	log.Printf("Analyzing food using Gemini model: %s", s.model)

	// This is a placeholder implementation that returns mock data
	return &FoodAnalysis{
		Name:       "Placeholder dish from Gemini",
		Carbs:      45.0,
		Confidence: 0.85,
	}, nil
}

func (s *geminiService) GetCurrentProvider() Provider {
	return ProviderGemini
}

// NewOpenAIService creates a new OpenAI service client
func NewOpenAIService(apiKey, model string) (Service, error) {
	// For now we'll return a placeholder implementation
	// TODO: Implement actual OpenAI API integration
	return &openAIService{
		apiKey: apiKey,
		model:  model,
	}, nil
}

// openAIService implements the Service interface for OpenAI API
type openAIService struct {
	apiKey string
	model  string
}

func (s *openAIService) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	// TODO: Implement actual OpenAI API call
	return fmt.Sprintf("Placeholder response from OpenAI using model %s (messages: %d)", s.model, len(messages)), nil
}

func (s *openAIService) AnalyzeFood(foodImagePath, ingredientImagePath string) (*FoodAnalysis, error) {
	// TODO: Implement actual OpenAI API call for food analysis with image input
	// This is a placeholder implementation that returns mock data
	return &FoodAnalysis{
		Name:       "Placeholder dish from OpenAI",
		Carbs:      50.0,
		Confidence: 0.92,
	}, nil
}

func (s *openAIService) GetCurrentProvider() Provider {
	return ProviderOpenAI
}

// NewGrokService creates a new Grok service client
func NewGrokService(apiKey string) (Service, error) {
	// For now we'll return a placeholder implementation
	// TODO: Implement actual Grok API integration
	return &grokService{apiKey: apiKey}, nil
}

// grokService implements the Service interface for Grok API
type grokService struct {
	apiKey string
}

func (s *grokService) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	// TODO: Implement actual Grok API call
	return fmt.Sprintf("Placeholder response from Grok (messages: %d)", len(messages)), nil
}

func (s *grokService) AnalyzeFood(foodImagePath, ingredientImagePath string) (*FoodAnalysis, error) {
	// TODO: Implement actual Grok API call for food analysis with image input
	// This is a placeholder implementation that returns mock data
	return &FoodAnalysis{
		Name:       "Placeholder dish from Grok",
		Carbs:      55.0,
		Confidence: 0.88,
	}, nil
}

func (s *grokService) GetCurrentProvider() Provider {
	return ProviderGrok
}
