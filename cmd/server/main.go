package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/yourusername/diabetes-assistant/internal/config"
	"github.com/yourusername/diabetes-assistant/internal/handlers"
	"github.com/yourusername/diabetes-assistant/internal/services/ai"
	"github.com/yourusername/diabetes-assistant/internal/services/libre"
	"github.com/yourusername/diabetes-assistant/internal/storage"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Ensure uploads directory exists
	uploadsDir := "./uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}

	// Convert to absolute path
	uploadsDir, err = filepath.Abs(uploadsDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	// Initialize AI service
	aiService, err := ai.NewService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize AI service: %v", err)
	}
	log.Printf("Using AI provider: %s", aiService.GetCurrentProvider())

	// Initialize Libre service
	libreService := libre.NewLibreService()

	// Try to initialize MongoDB storage with helpful error messages
	var dbStorage storage.Storage
	useInMemoryFallback := false

	// Set MongoDB URI - update this with your actual MongoDB URI or use environment variable
	mongoURI := "mongodb://localhost:27017/diabetes_assistant"
	if cfg.MongoURI != "" {
		mongoURI = cfg.MongoURI
	}

	log.Printf("Connecting to MongoDB at %s", getMaskedMongoURI(mongoURI))
	mongoDBStorage, err := storage.NewMongoDBStorage(mongoURI)

	if err != nil {
		log.Printf("MongoDB connection error: %v", err)

		if strings.Contains(err.Error(), "connection refused") {
			log.Printf("\n%s\n", strings.Repeat("-", 80))
			log.Println("ERROR: Could not connect to MongoDB. Please check that:")
			log.Println("1. MongoDB is installed and running")
			log.Println("2. The connection URI is correct")
			log.Println("")
			log.Println("For local MongoDB:")
			log.Println("  - macOS: brew services start mongodb-community")
			log.Println("  - Linux: sudo systemctl start mongod")
			log.Println("  - Windows: Check Service")
			log.Printf("%s\n", strings.Repeat("-", 80))
		}

		// Ask user if they want to use in-memory storage instead
		log.Println("\nWould you like to continue with in-memory storage? (Data will be lost when the app restarts)")
		log.Println("Enter 'y' to continue with in-memory storage, or any other key to exit:")

		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) == "y" {
			log.Println("Using in-memory storage as fallback. Note: all data will be lost when the application restarts.")
			memoryStorage := storage.NewInMemoryStorage()
			dbStorage = memoryStorage
			useInMemoryFallback = true
		} else {
			log.Fatalf("Failed to connect to MongoDB. Application cannot start without database.")
		}
	} else {
		// Use MongoDB storage
		dbStorage = mongoDBStorage
		log.Println("Successfully connected to MongoDB")
		defer mongoDBStorage.Close()
	}

	// Create API handler
	apiHandler := handlers.NewAPIHandler(dbStorage, aiService, libreService, uploadsDir)

	// Create router
	router := mux.NewRouter()

	// Add middleware for debugging
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Request: %s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})

	// Add CORS middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// API routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/health", apiHandler.HealthCheck).Methods("GET")
	api.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"pong"}`))
	}).Methods("GET")
	api.HandleFunc("/test-post", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"post successful"}`))
	}).Methods("POST", "OPTIONS")
	api.HandleFunc("/settings/{userId}", apiHandler.GetUserSettings).Methods("GET")
	api.HandleFunc("/settings/{userId}", apiHandler.SaveUserSettings).Methods("POST")
	api.HandleFunc("/bloodsugar/{userId}", apiHandler.GetBloodSugarReadings).Methods("GET")
	api.HandleFunc("/bloodsugar", apiHandler.SaveBloodSugar).Methods("POST")
	api.HandleFunc("/bloodsugar", apiHandler.DeleteBloodSugar).Methods("DELETE")
	api.HandleFunc("/analyze-food", apiHandler.AnalyzeFood).Methods("POST")
	api.HandleFunc("/sync-libre", apiHandler.SyncLibre).Methods("POST")

	// Serve static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web")))

	// Create server
	port := "8080" // Use port 8080
	server := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%s", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a separate goroutine
	go func() {
		log.Printf("Server listening on port %s", port)
		if useInMemoryFallback {
			log.Printf("\n%s\n", strings.Repeat("*", 80))
			log.Printf("NOTICE: Running with in-memory storage. All data will be lost when the server stops.")
			log.Printf("For persistent storage, please configure MongoDB or MongoDB Atlas.")
			log.Printf("%s\n\n", strings.Repeat("*", 80))
		}

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down server
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server gracefully stopped")
}

// getMaskedMongoURI masks sensitive information in MongoDB URI for logging
func getMaskedMongoURI(uri string) string {
	// If it's a MongoDB Atlas URI (contains username and password)
	if strings.Contains(uri, "@") {
		parts := strings.Split(uri, "@")
		if len(parts) >= 2 {
			credentials := strings.Split(parts[0], "://")
			if len(credentials) >= 2 {
				// Mask the credentials part but keep protocol
				return credentials[0] + "://" + "******:******@" + parts[1]
			}
		}
	}
	return uri
}
