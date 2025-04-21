package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiProvider implements the Provider interface for Google's Gemini API
type GeminiProvider struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(apiKey string) (*GeminiProvider, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Use gemini-1.5-pro as it offers the best quality/performance while being free for reasonable usage
	model := client.GenerativeModel("gemini-1.5-pro")

	// Configure the model with appropriate settings for medical analysis
	model.SetTemperature(0.2) // Lower temperature for more deterministic, accurate responses
	model.SetTopK(40)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(1024) // Allow enough tokens for detailed analysis

	return &GeminiProvider{
		client: client,
		model:  model,
	}, nil
}

// AnalyzeFood analyzes a food image and returns the estimated carbohydrates
func (p *GeminiProvider) AnalyzeFood(foodImagePath, unusedDescriptionParam string, foodWeight float64) (*FoodAnalysisResult, error) {
	// Description parameter is no longer used, only photo and weight
	ctx := context.Background()

	// Read food image file
	foodImgData, err := os.ReadFile(foodImagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read food image: %w", err)
	}

	// Create the base prompt
	promptText := `You are a certified diabetes educator specializing in nutrition analysis. 
You will analyze the food in the image to estimate its carbohydrate content accurately for diabetes management.

TASK:
1. Identify the food items in the image
2. Estimate total carbohydrates (in grams) based on standard nutritional databases
3. Assess your confidence in this estimation (low, medium, high)
4. Provide the information in a specific JSON format

REQUIREMENTS:
- Be medically precise in your carbohydrate estimation
- Include both visible ingredients and likely hidden ingredients that contain carbs
- Consider portion sizes carefully
- Account for various cooking methods that might affect carbohydrate content
- If the image contains nutritional information or packaging, prioritize that data
- IMPORTANT: Provide all text responses in Russian language for Russian users
- Food names should be in Russian
- Reasoning/descriptions should be in Russian`

	// Add weight information if provided
	if foodWeight > 0 {
		promptText += fmt.Sprintf(`

IMPORTANT WEIGHT INFORMATION:
- The user has specified that the food weighs %.1f grams
- Adjust your carbohydrate calculation based on this exact weight
- Make sure to mention the weight in your reasoning`, foodWeight)
	}

	promptText += `

RESPONSE FORMAT:
Respond ONLY with valid JSON matching this exact structure:
{
  "name": "Complete name of the dish in Russian",
  "carbs": number, 
  "confidence": "low|medium|high",
  "reasoning": "Brief explanation of how you estimated the carbs in Russian"
}

This information will be used for insulin dosing, so accuracy is critically important for patient safety.`

	prompt := genai.Text(promptText)

	// Create image part
	img := genai.ImageData("image/jpeg", foodImgData)
	parts := []genai.Part{prompt, img}

	// Generate content
	log.Printf("Sending request to Gemini for food analysis with model: gemini-1.5-pro")
	resp, err := p.model.GenerateContent(ctx, parts...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Extract the response text
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	responseText, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return nil, fmt.Errorf("unexpected response type from Gemini")
	}

	log.Printf("Received Gemini response: %s", string(responseText)[:min(100, len(string(responseText)))]+"...")

	// Parse the JSON response
	var result struct {
		Name       string  `json:"name"`
		Carbs      float64 `json:"carbs"`
		Confidence string  `json:"confidence"`
		Reasoning  string  `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(string(responseText)), &result); err != nil {
		// Try to extract JSON from a text response
		jsonStr, extractErr := extractJSONFromText(string(responseText))
		if extractErr != nil {
			return nil, fmt.Errorf("failed to parse response: %w (response was: %s)", err, truncateString(string(responseText), 200))
		}

		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			return nil, fmt.Errorf("failed to parse extracted JSON: %w", err)
		}
	}

	// Convert to the expected return format
	return &FoodAnalysisResult{
		Name:       result.Name,
		Carbs:      result.Carbs,
		Confidence: result.Confidence,
		Reasoning:  result.Reasoning,
	}, nil
}

// truncateString truncates a string to the specified length and adds "..." if truncated
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
