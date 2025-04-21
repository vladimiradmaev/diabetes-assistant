package config

import (
	"os"
)

// Config holds the application configuration
type Config struct {
	Port         string
	MongoURI     string
	GeminiToken  string
	OpenAIToken  string
	GrokToken    string
	DefaultModel string
}

// LoadConfig loads the application configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		Port:         getEnvWithDefault("PORT", "8080"),
		MongoURI:     getEnvWithDefault("MONGODB_URI", "mongodb://localhost:27017/diabetes_assistant"),
		GeminiToken:  os.Getenv("GEMINI_API_KEY"),
		OpenAIToken:  os.Getenv("OPENAI_API_KEY"),
		GrokToken:    os.Getenv("GROK_API_KEY"),
		DefaultModel: getEnvWithDefault("DEFAULT_MODEL", "gpt-3.5-turbo"),
	}

	return config, nil
}

// getEnvWithDefault returns the value of an environment variable or a default value
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
