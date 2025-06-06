package ai

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// OpenAIProvider implements the Provider interface for OpenAI's API
type OpenAIProvider struct {
	apiKey string
}

type openAIImageAnalysisRequest struct {
	Model     string               `json:"model"`
	Messages  []openAIMessageInput `json:"messages"`
	MaxTokens int                  `json:"max_tokens"`
}

type openAIMessageInput struct {
	Role    string        `json:"role"`
	Content []interface{} `json:"content"`
}

type openAITextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type openAIImageContent struct {
	Type     string             `json:"type"`
	ImageURL openAIImageURLData `json:"image_url"`
}

type openAIImageURLData struct {
	URL string `json:"url"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey string) (*OpenAIProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	return &OpenAIProvider{
		apiKey: apiKey,
	}, nil
}

// AnalyzeFood analyzes a food image and returns the estimated carbohydrates
func (p *OpenAIProvider) AnalyzeFood(foodImagePath, unusedDescriptionParam string, foodWeight float64) (*FoodAnalysisResult, error) {
	// Description parameter is no longer used, only photo and weight
	// Read food image file
	foodImg, err := os.ReadFile(foodImagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read food image: %w", err)
	}

	// Convert image to base64
	foodImgBase64 := base64.StdEncoding.EncodeToString(foodImg)

	// Create content items for the OpenAI request
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

	contentItems := []interface{}{
		openAITextContent{
			Type: "text",
			Text: promptText,
		},
		openAIImageContent{
			Type: "image_url",
			ImageURL: openAIImageURLData{
				URL: "data:image/jpeg;base64," + foodImgBase64,
			},
		},
	}

	// Create the request payload
	payload := openAIImageAnalysisRequest{
		Model: "gpt-4-vision-preview",
		Messages: []openAIMessageInput{
			{
				Role:    "user",
				Content: contentItems,
			},
		},
		MaxTokens: 1024, // Increased token limit to allow for detailed reasoning
	}

	// Marshal the payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(payloadJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to OpenAI: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the response
	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if openAIResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	// Check if we got a response
	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Extract the content
	content := openAIResp.Choices[0].Message.Content

	// Parse the JSON response
	var result struct {
		Name       string  `json:"name"`
		Carbs      float64 `json:"carbs"`
		Confidence string  `json:"confidence"`
		Reasoning  string  `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// Try to extract JSON from a text response
		jsonStr, extractErr := extractJSONFromText(content)
		if extractErr != nil {
			return nil, fmt.Errorf("failed to parse response: %w (response was: %s)", err, content)
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
