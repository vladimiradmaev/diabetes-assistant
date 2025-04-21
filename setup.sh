#!/bin/bash

# Colors for terminal output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color
BOLD='\033[1m'

echo -e "${BOLD}Diabetes Assistant Setup${NC}"
echo -e "============================\n"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed or not in the PATH.${NC}"
    echo -e "Please install Go from https://golang.org/dl/"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
MIN_GO_VERSION="1.20"
if [[ $(echo -e "$GO_VERSION\n$MIN_GO_VERSION" | sort -V | head -n1) != $MIN_GO_VERSION ]]; then
    echo -e "${YELLOW}Warning: Go version $GO_VERSION detected. We recommend Go $MIN_GO_VERSION or higher.${NC}"
    read -p "Do you want to continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo -e "${YELLOW}Creating .env file...${NC}"
    cat > .env << EOL
# Server configuration
PORT=8080

# Database configuration
MONGODB_URI=mongodb://localhost:27017/diabetes_assistant

# AI provider API keys
OPENAI_API_KEY=
GEMINI_API_KEY=
GROK_API_KEY=

# Default model for OpenAI
DEFAULT_MODEL=gpt-4-turbo
EOL
    echo -e "${GREEN}Created .env file.${NC}"
fi

# Check for MongoDB installation
echo -e "\n${BOLD}Checking MongoDB installation${NC}"
MONGO_LOCAL="false"

if command -v mongod &> /dev/null; then
    echo -e "${GREEN}MongoDB is installed.${NC}"
    # Check if MongoDB is running
    if pgrep mongod &> /dev/null; then
        echo -e "${GREEN}MongoDB is running.${NC}"
        MONGO_LOCAL="true"
    else
        echo -e "${YELLOW}MongoDB is installed but not running.${NC}"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            echo -e "Start MongoDB with: ${BOLD}brew services start mongodb-community${NC}"
        elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
            echo -e "Start MongoDB with: ${BOLD}sudo systemctl start mongod${NC}"
        else
            echo -e "Please start MongoDB manually."
        fi
        
        read -p "Do you want to use a local MongoDB installation? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${YELLOW}Please start MongoDB manually, then continue the setup.${NC}"
            MONGO_LOCAL="true"
        fi
    fi
else
    echo -e "${YELLOW}MongoDB is not installed.${NC}"
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo -e "You can install MongoDB using Homebrew with:"
        echo -e "${BOLD}brew tap mongodb/brew${NC}"
        echo -e "${BOLD}brew install mongodb-community${NC}"
        echo -e "And then start it with: ${BOLD}brew services start mongodb-community${NC}"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo -e "You can install MongoDB on Ubuntu with:"
        echo -e "${BOLD}sudo apt-get install -y mongodb${NC}"
        echo -e "And then start it with: ${BOLD}sudo systemctl start mongodb${NC}"
    else
        echo -e "Please visit https://docs.mongodb.com/manual/installation/ for installation instructions."
    fi
    
    read -p "Do you want to install MongoDB now? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        if [[ "$OSTYPE" == "darwin"* ]]; then
            echo -e "Installing MongoDB via Homebrew..."
            brew tap mongodb/brew && brew install mongodb-community
            brew services start mongodb-community
            if [ $? -eq 0 ]; then
                echo -e "${GREEN}MongoDB installed and started successfully.${NC}"
                MONGO_LOCAL="true"
            else
                echo -e "${RED}Failed to install or start MongoDB.${NC}"
            fi
        elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
            echo -e "${YELLOW}Please install MongoDB manually using apt or your distribution's package manager.${NC}"
        fi
    fi
fi

# Configure MongoDB URI
echo -e "\n${BOLD}MongoDB Configuration${NC}"
if [ "$MONGO_LOCAL" = "true" ]; then
    echo -e "Using local MongoDB installation."
    read -p "Enter MongoDB URI (default: mongodb://localhost:27017/diabetes_assistant): " MONGODB_URI
    MONGODB_URI=${MONGODB_URI:-mongodb://localhost:27017/diabetes_assistant}
else
    echo -e "You can use MongoDB Atlas (cloud MongoDB) as an alternative to local installation."
    echo -e "Sign up at https://www.mongodb.com/cloud/atlas (there's a free tier available)."
    echo -e "Create a cluster, obtain your connection string, and add your IP to the allowlist."
    echo -e "\nEnter your MongoDB Atlas connection string or provide a different MongoDB URI."
    echo -e "Format example: mongodb+srv://username:password@clustername.mongodb.net/diabetes_assistant"
    read -p "MongoDB URI: " MONGODB_URI
    
    if [ -z "$MONGODB_URI" ]; then
        echo -e "${RED}No MongoDB URI provided. Using default, but the application may not work without a valid connection.${NC}"
        MONGODB_URI="mongodb://localhost:27017/diabetes_assistant"
    fi
fi

sed -i.bak "s|MONGODB_URI=.*|MONGODB_URI=$MONGODB_URI|g" .env

# AI provider configuration
echo -e "\n${BOLD}AI Provider Configuration${NC}"
echo -e "You need at least one AI provider API key to analyze food images.\n"

# OpenAI
read -p "Enter OpenAI API Key (leave empty to skip): " OPENAI_API_KEY
sed -i.bak "s|OPENAI_API_KEY=.*|OPENAI_API_KEY=$OPENAI_API_KEY|g" .env

# Gemini (Google)
read -p "Enter Google Gemini API Key (leave empty to skip): " GEMINI_API_KEY
sed -i.bak "s|GEMINI_API_KEY=.*|GEMINI_API_KEY=$GEMINI_API_KEY|g" .env

# Grok (X/Twitter)
read -p "Enter Grok API Key (leave empty to skip): " GROK_API_KEY
sed -i.bak "s|GROK_API_KEY=.*|GROK_API_KEY=$GROK_API_KEY|g" .env

# Check if at least one API key is provided
if [ -z "$OPENAI_API_KEY" ] && [ -z "$GEMINI_API_KEY" ] && [ -z "$GROK_API_KEY" ]; then
    echo -e "${RED}Warning: No AI provider API key provided.${NC}"
    echo -e "The food analysis feature will not work without at least one API key."
    read -p "Do you want to continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Create uploads directory
mkdir -p uploads

# Download dependencies
echo -e "\n${BOLD}Downloading dependencies...${NC}"
go mod tidy

# Clean up backup files
rm -f .env.bak

echo -e "\n${GREEN}Setup completed successfully!${NC}"
echo -e "You can now run the application using:"
echo -e "${BOLD}go run cmd/server/main.go${NC}"
echo -e "or build it with:"
echo -e "${BOLD}go build -o diabetes-assistant cmd/server/main.go${NC}\n"

# Make the script executable
chmod +x setup.sh 