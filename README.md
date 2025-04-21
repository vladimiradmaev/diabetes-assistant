# Diabetes Assistant

A Go-based application to help diabetics manage their condition. It provides tools for analyzing carbohydrates in food using AI, tracking blood sugar readings, and calculating insulin doses.

## Features

- 🍔 **Food Analysis**: Upload food photos to get carbohydrate estimates using AI
- 📊 **Blood Sugar Tracking**: Record and visualize blood sugar readings
- 💉 **Insulin Calculator**: Get insulin dose recommendations based on food and blood sugar
- 🔄 **Data Integration**: Connect with Freestyle Libre systems, LibreView, and Nightscout
- 🧠 **AI Provider Flexibility**: Choose between different AI providers (OpenAI, Google Gemini, Grok)

## Requirements

- Go 1.20 or higher
- MongoDB (local installation or MongoDB Atlas cloud service)
- At least one AI provider API key (OpenAI, Google Gemini, or Grok)

## Getting Started

1. Clone this repository:
   ```
   git clone https://github.com/yourusername/diabetes-assistant.git
   cd diabetes-assistant
   ```

2. Run the setup script:
   ```
   ./setup.sh
   ```
   
   This script will:
   - Check Go installation
   - Check MongoDB installation and provide installation options
   - Guide you through configuring MongoDB (local or Atlas)
   - Help you set up AI provider API keys
   - Download dependencies

3. Start the application:
   ```
   go run cmd/server/main.go
   ```
   
   The server will start on port 8080 by default (configurable in .env).

4. Access the web interface at http://localhost:8080

## MongoDB Setup

This application requires MongoDB. You have two options:

1. **Local MongoDB Installation**: Install MongoDB directly on your machine
2. **MongoDB Atlas**: Use the cloud MongoDB service (free tier available)

The setup script will guide you through both options. For detailed MongoDB setup instructions, see [MONGODB.md](MONGODB.md).

## Configuration

All configuration is done through environment variables, which can be set in the `.env` file:

- `PORT`: Server port (default: 8080)
- `MONGODB_URI`: MongoDB connection string
- `OPENAI_API_KEY`: OpenAI API key
- `GEMINI_API_KEY`: Google Gemini API key
- `GROK_API_KEY`: Grok API key
- `DEFAULT_MODEL`: Default model for OpenAI (default: gpt-4-turbo)

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/health` | GET | Health check |
| `/api/settings/{userId}` | GET | Get user settings |
| `/api/settings` | POST | Save user settings |
| `/api/bloodsugar/{userId}` | GET | Get blood sugar readings |
| `/api/bloodsugar` | POST | Save a blood sugar reading |
| `/api/analyze-food` | POST | Analyze food image |
| `/api/sync-libre` | POST | Sync blood sugar readings |

## AI Provider Selection

The application supports multiple AI providers for food analysis:

1. **Google Gemini**: Default choice if API key is provided
2. **OpenAI**: Used if Gemini is unavailable but OpenAI key is provided
3. **Grok**: Used as fallback if neither Gemini nor OpenAI keys are provided

You can provide API keys for any or all of these services, and the application will automatically select the first available provider in the order listed above.

## Building for Production

To build the application for production:

```
go build -o diabetes-assistant cmd/server/main.go
```

Then run the executable:

```
./diabetes-assistant
```

## Troubleshooting

If you encounter database connection errors, see the [MongoDB Troubleshooting Guide](MONGODB.md#troubleshooting-mongodb-connection-issues).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - See LICENSE file for details. 