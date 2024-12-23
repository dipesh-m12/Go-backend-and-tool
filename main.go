package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
    // Command line flags for configuration
    url := flag.String("url", "", "Target URL to test")
    numRequests := flag.Int("n", 100, "Number of requests to send")
    concurrency := flag.Int("c", 10, "Number of concurrent workers")
    method := flag.String("method", "GET", "HTTP method to use")
    timeout := flag.Int("timeout", 10, "Timeout in seconds")
    flag.Parse()

    if *url == "" {
        log.Fatal("URL is required. Use -url flag")
    }

    // Create a custom HTTP client with timeout
    client := &http.Client{
        Timeout: time.Duration(*timeout) * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        *concurrency,
            MaxIdleConnsPerHost: *concurrency,
            DisableKeepAlives:   true,
        },
    }

    // Create a channel to feed work to workers
    jobs := make(chan int, *numRequests)
    var wg sync.WaitGroup

    // Start time for metrics
    startTime := time.Now()

    // Create workers
    for w := 1; w <= *concurrency; w++ {
        wg.Add(1)
        go worker(w, jobs, client, *url, *method, &wg)
    }

    // Send requests to workers
    for i := 1; i <= *numRequests; i++ {
        jobs <- i
    }
    close(jobs)

    // Wait for all workers to complete
    wg.Wait()

    // Calculate metrics
    duration := time.Since(startTime)
    requestsPerSecond := float64(*numRequests) / duration.Seconds()

    fmt.Printf("\nLoad Test Results:\n")
    fmt.Printf("Total Requests: %d\n", *numRequests)
    fmt.Printf("Concurrency Level: %d\n", *concurrency)
    fmt.Printf("Time taken: %.2f seconds\n", duration.Seconds())
    fmt.Printf("Requests per second: %.2f\n", requestsPerSecond)
}

func worker(id int, jobs <-chan int, client *http.Client, url string, method string, wg *sync.WaitGroup) {
    defer wg.Done()
    
    // Create a WaitGroup for the requests within this worker
    var requestWg sync.WaitGroup

    for range jobs {
        requestWg.Add(1)
        // Create request
        req, err := http.NewRequest(method, url, nil)
        if err != nil {
            log.Printf("Worker %d: Error creating request: %v\n", id, err)
            requestWg.Done()
            continue
        }

        // Set some basic headers
        req.Header.Set("User-Agent", "LoadTester/1.0")
        req.Header.Set("Accept", "*/*")

        // Send request and wait for response
        go func() {
            defer requestWg.Done()
            resp, err := client.Do(req)
            if err != nil {
                log.Printf("Worker %d: Request error: %v\n", id, err)
                return
            }
            defer resp.Body.Close()
            log.Printf("Worker %d: Request completed with status %d\n", id, resp.StatusCode)
        }()
    }

    // Wait for all requests in this worker to complete
    requestWg.Wait()
}