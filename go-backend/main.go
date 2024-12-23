package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

// User represents a basic user model
type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// UserHandler handles user-related requests
func UserHandler(w http.ResponseWriter, r *http.Request) {
    // Set JSON content type
    w.Header().Set("Content-Type", "application/json")

    switch r.Method {
    case "GET":
        // Sample user data
        user := User{
            ID:    "1",
            Name:  "John Doe",
            Email: "john@example.com",
        }
        json.NewEncoder(w).Encode(user)

    case "POST":
        var newUser User
        if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        // Here you would typically save the user to a database
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(newUser)

    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

// Global counter for requests
var requestCount uint64

// HealthCheckHandler provides a basic health check endpoint with request counting
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
    // Increment request counter atomically
    currentCount := atomic.AddUint64(&requestCount, 1)
    
    // Set response headers
    w.Header().Set("Content-Type", "application/json")
    
    // Log the request with count
    log.Printf("Health check request received (Request #%d)", currentCount)
    
    // Create response object with request count
    response := map[string]interface{}{
        "status": "healthy",
        "request_number": currentCount,
        "timestamp": time.Now().Format(time.RFC3339),
    }
    
    // Encode and handle potential encoding errors
    if err := json.NewEncoder(w).Encode(response); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        log.Printf("Error encoding response: %v", err)
        return
    }
}

func main() {
    // Define routes
    http.HandleFunc("/api/user", UserHandler)
    http.HandleFunc("/health", HealthCheckHandler)

    // Start server
    log.Println("Server starting on port 8080...")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}