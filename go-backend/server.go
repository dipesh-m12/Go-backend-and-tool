package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
		// In a real application, you would typically save the user to a database here
		log.Printf("Received new user: %+v", newUser) // Log received user data
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
		"status":         "healthy",
		"request_number": currentCount,
		"timestamp":      time.Now().Format(time.RFC3339),
	}

	// Encode and handle potential encoding errors
	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error encoding response: %v", err)
		return
	}
}

// isPrime is a helper function for simulating CPU-bound work
func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// HeavyEndpointHandler simulates a CPU-intensive or blocking operation
func HeavyEndpointHandler(w http.ResponseWriter, r *http.Request) {
	// Get parameters for simulating work
	cpuWorkStr := r.URL.Query().Get("cpu_work_ms") // CPU work duration in milliseconds
	ioDelayStr := r.URL.Query().Get("io_delay_ms") // I/O delay duration in milliseconds

	cpuWorkMs := 0
	if cpuWorkStr != "" {
		if val, err := strconv.Atoi(cpuWorkStr); err == nil && val > 0 {
			cpuWorkMs = val
		}
	}

	ioDelayMs := 0
	if ioDelayStr != "" {
		if val, err := strconv.Atoi(ioDelayStr); err == nil && val > 0 {
			ioDelayMs = val
		}
	}

	log.Printf("Heavy endpoint received request. CPU Work: %dms, I/O Delay: %dms", cpuWorkMs, ioDelayMs)
	startTime := time.Now()

	// Simulate CPU-intensive work (e.g., finding primes up to a certain limit)
	if cpuWorkMs > 0 {
		log.Printf("Simulating %dms of CPU-intensive work...", cpuWorkMs)
		cpuWorkStartTime := time.Now()
		// Adjust the loop count to roughly match the desired CPU work time.
		// This is a crude approximation; actual CPU time depends on hardware.
		targetIterations := 10000000 // Base iterations for very light work
		factor := float64(cpuWorkMs) / 100.0 // Scale factor based on 100ms baseline

		current := 0
		for i := 0; float64(i) < float64(targetIterations)*factor && time.Since(cpuWorkStartTime) < time.Duration(cpuWorkMs)*time.Millisecond; i++ {
			// Perform some non-trivial computation in the loop
			current++
			if current % 1000 == 0 {
				isPrime(current) // Call a function to keep CPU busy
			}
		}
		log.Printf("Finished CPU work after %s", time.Since(cpuWorkStartTime))
	}

	// Simulate I/O blocking operation (e.g., database call, external API request)
	if ioDelayMs > 0 {
		log.Printf("Simulating %dms of I/O delay...", ioDelayMs)
		time.Sleep(time.Duration(ioDelayMs) * time.Millisecond)
		log.Printf("Finished I/O delay.")
	}

	duration := time.Since(startTime)
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message":     "Heavy processing completed",
		"cpu_work_ms": cpuWorkMs,
		"io_delay_ms": ioDelayMs,
		"total_time":  fmt.Sprintf("%.2fms", float64(duration.Microseconds())/1000.0),
		"timestamp":   time.Now().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(response)
	log.Printf("Heavy endpoint request processed in %s", duration)
}

func main() {
	// Define routes
	http.HandleFunc("/api/user", UserHandler)
	http.HandleFunc("/health", HealthCheckHandler)
	http.HandleFunc("/heavy", HeavyEndpointHandler) // New heavy endpoint

	// Start server
	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
