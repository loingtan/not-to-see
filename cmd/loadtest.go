package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// LoadTestConfig holds configuration for load testing
type LoadTestConfig struct {
	BaseURL         string
	NumStudents     int
	NumSections     int
	ConcurrentUsers int
	RequestsPerUser int
	TestDurationSec int
	SectionCapacity int
}

// RegistrationRequest represents the API request
type RegistrationRequest struct {
	StudentID  uuid.UUID   `json:"student_id"`
	SectionIDs []uuid.UUID `json:"section_ids"`
}

// LoadTestResult holds the results of load testing
type LoadTestResult struct {
	TotalRequests     int
	SuccessfulReqs    int
	FailedReqs        int
	WaitlistedReqs    int
	AvgResponseTimeMs float64
	MaxResponseTimeMs int64
	MinResponseTimeMs int64
	ThroughputRPS     float64
	ErrorsByType      map[string]int
}

// LoadTester handles course registration load testing
type LoadTester struct {
	config    LoadTestConfig
	client    *http.Client
	students  []uuid.UUID
	sections  []uuid.UUID
	results   LoadTestResult
	mutex     sync.Mutex
	startTime time.Time
}

// NewLoadTester creates a new load tester
func NewLoadTester(config LoadTestConfig) *LoadTester {
	return &LoadTester{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		students: make([]uuid.UUID, config.NumStudents),
		sections: make([]uuid.UUID, config.NumSections),
		results: LoadTestResult{
			ErrorsByType: make(map[string]int),
		},
	}
}

// Initialize sets up test data
func (lt *LoadTester) Initialize() {
	fmt.Println("Initializing load test data...")

	// Generate student IDs
	for i := 0; i < lt.config.NumStudents; i++ {
		lt.students[i] = uuid.New()
	}

	// Generate section IDs - simulate sections with limited capacity
	for i := 0; i < lt.config.NumSections; i++ {
		lt.sections[i] = uuid.New()
	}

	fmt.Printf("Generated %d students and %d sections\n", len(lt.students), len(lt.sections))
}

// RunLoadTest executes the load test
func (lt *LoadTester) RunLoadTest() {
	fmt.Printf("Starting load test with %d concurrent users...\n", lt.config.ConcurrentUsers)

	lt.startTime = time.Now()
	var wg sync.WaitGroup

	// Create semaphore to limit concurrent requests
	semaphore := make(chan struct{}, lt.config.ConcurrentUsers)

	// Calculate total requests to distribute across users
	totalRequests := lt.config.ConcurrentUsers * lt.config.RequestsPerUser

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)

		go func(requestID int) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			lt.simulateUserRegistration(requestID)
		}(i)

		// Add small delay between request starts to simulate realistic user behavior
		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()

	// Calculate final metrics
	lt.calculateMetrics()
	lt.printResults()
}

// simulateUserRegistration simulates a single user's registration attempt
func (lt *LoadTester) simulateUserRegistration(requestID int) {
	startTime := time.Now()

	// Select random student and sections
	studentID := lt.students[requestID%len(lt.students)]

	// Simulate trying to register for 1-3 sections (common scenario)
	numSections := 1 + (requestID % 3)
	sectionIDs := make([]uuid.UUID, numSections)

	for i := 0; i < numSections; i++ {
		sectionIDs[i] = lt.sections[(requestID+i)%len(lt.sections)]
	}

	// Create registration request
	reqBody := RegistrationRequest{
		StudentID:  studentID,
		SectionIDs: sectionIDs,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		lt.recordError("json_marshal", startTime)
		return
	}

	// Make HTTP request
	url := fmt.Sprintf("%s/api/register", lt.config.BaseURL)
	resp, err := lt.client.Post(url, "application/json", bytes.NewBuffer(jsonData))

	responseTime := time.Since(startTime)

	if err != nil {
		lt.recordError("http_request", startTime)
		return
	}
	defer resp.Body.Close()

	// Record response metrics
	lt.recordResponse(resp.StatusCode, responseTime)
}

// recordResponse records the response metrics
func (lt *LoadTester) recordResponse(statusCode int, responseTime time.Duration) {
	lt.mutex.Lock()
	defer lt.mutex.Unlock()

	lt.results.TotalRequests++
	responseTimeMs := responseTime.Milliseconds()

	// Update response time metrics
	if lt.results.MaxResponseTimeMs < responseTimeMs {
		lt.results.MaxResponseTimeMs = responseTimeMs
	}

	if lt.results.MinResponseTimeMs == 0 || lt.results.MinResponseTimeMs > responseTimeMs {
		lt.results.MinResponseTimeMs = responseTimeMs
	}

	// Calculate running average
	currentAvg := lt.results.AvgResponseTimeMs
	currentCount := float64(lt.results.TotalRequests)
	lt.results.AvgResponseTimeMs = (currentAvg*(currentCount-1) + float64(responseTimeMs)) / currentCount

	// Categorize responses
	switch {
	case statusCode >= 200 && statusCode < 300:
		lt.results.SuccessfulReqs++
	case statusCode == 409: // Conflict - likely waitlisted
		lt.results.WaitlistedReqs++
	default:
		lt.results.FailedReqs++
		lt.results.ErrorsByType[fmt.Sprintf("http_%d", statusCode)]++
	}
}

// recordError records an error that occurred during testing
func (lt *LoadTester) recordError(errorType string, startTime time.Time) {
	lt.mutex.Lock()
	defer lt.mutex.Unlock()

	lt.results.TotalRequests++
	lt.results.FailedReqs++
	lt.results.ErrorsByType[errorType]++
}

// calculateMetrics calculates final test metrics
func (lt *LoadTester) calculateMetrics() {
	totalDuration := time.Since(lt.startTime)
	lt.results.ThroughputRPS = float64(lt.results.TotalRequests) / totalDuration.Seconds()
}

// printResults displays the load test results
func (lt *LoadTester) printResults() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("Test Configuration:\n")
	fmt.Printf("  - Concurrent Users: %d\n", lt.config.ConcurrentUsers)
	fmt.Printf("  - Requests per User: %d\n", lt.config.RequestsPerUser)
	fmt.Printf("  - Total Students: %d\n", lt.config.NumStudents)
	fmt.Printf("  - Total Sections: %d\n", lt.config.NumSections)
	fmt.Printf("  - Section Capacity: %d seats each\n", lt.config.SectionCapacity)

	fmt.Printf("\nOverall Performance:\n")
	fmt.Printf("  - Total Requests: %d\n", lt.results.TotalRequests)
	fmt.Printf("  - Successful: %d (%.2f%%)\n",
		lt.results.SuccessfulReqs,
		float64(lt.results.SuccessfulReqs)/float64(lt.results.TotalRequests)*100)
	fmt.Printf("  - Waitlisted: %d (%.2f%%)\n",
		lt.results.WaitlistedReqs,
		float64(lt.results.WaitlistedReqs)/float64(lt.results.TotalRequests)*100)
	fmt.Printf("  - Failed: %d (%.2f%%)\n",
		lt.results.FailedReqs,
		float64(lt.results.FailedReqs)/float64(lt.results.TotalRequests)*100)

	fmt.Printf("\nResponse Time Metrics:\n")
	fmt.Printf("  - Average: %.2f ms\n", lt.results.AvgResponseTimeMs)
	fmt.Printf("  - Minimum: %d ms\n", lt.results.MinResponseTimeMs)
	fmt.Printf("  - Maximum: %d ms\n", lt.results.MaxResponseTimeMs)

	fmt.Printf("\nThroughput:\n")
	fmt.Printf("  - Requests per Second: %.2f\n", lt.results.ThroughputRPS)

	if len(lt.results.ErrorsByType) > 0 {
		fmt.Printf("\nError Breakdown:\n")
		for errorType, count := range lt.results.ErrorsByType {
			fmt.Printf("  - %s: %d\n", errorType, count)
		}
	}

	// Performance analysis
	fmt.Printf("\nPerformance Analysis:\n")
	lt.analyzePerformance()
}

// analyzePerformance provides performance insights
func (lt *LoadTester) analyzePerformance() {
	successRate := float64(lt.results.SuccessfulReqs) / float64(lt.results.TotalRequests) * 100

	if lt.results.AvgResponseTimeMs > 1000 {
		fmt.Printf("  ⚠️  High average response time (>1s) indicates potential bottlenecks\n")
	} else if lt.results.AvgResponseTimeMs > 500 {
		fmt.Printf("  ⚠️  Moderate response time, monitor under higher load\n")
	} else {
		fmt.Printf("  ✅ Good response time performance\n")
	}

	if successRate < 50 {
		fmt.Printf("  ❌ Low success rate indicates system overload or issues\n")
	} else if successRate < 80 {
		fmt.Printf("  ⚠️  Moderate success rate, consider capacity planning\n")
	} else {
		fmt.Printf("  ✅ Good success rate\n")
	}

	if lt.results.ThroughputRPS < 10 {
		fmt.Printf("  ❌ Low throughput, system may not handle production load\n")
	} else if lt.results.ThroughputRPS < 50 {
		fmt.Printf("  ⚠️  Moderate throughput, monitor scaling requirements\n")
	} else {
		fmt.Printf("  ✅ Good throughput performance\n")
	}

	// Calculate contention metrics
	totalSeats := lt.config.NumSections * lt.config.SectionCapacity
	totalDemand := lt.results.TotalRequests
	contentionRatio := float64(totalDemand) / float64(totalSeats)

	fmt.Printf("\nContention Analysis:\n")
	fmt.Printf("  - Total Available Seats: %d\n", totalSeats)
	fmt.Printf("  - Total Registration Attempts: %d\n", totalDemand)
	fmt.Printf("  - Contention Ratio: %.2f:1\n", contentionRatio)

	if contentionRatio > 5 {
		fmt.Printf("  ❌ Very high contention - expect many waitlists\n")
	} else if contentionRatio > 2 {
		fmt.Printf("  ⚠️  High contention - some waitlisting expected\n")
	} else {
		fmt.Printf("  ✅ Reasonable contention level\n")
	}
}

// RunConcurrencyStressTest tests system under extreme concurrent load
func (lt *LoadTester) RunConcurrencyStressTest() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("CONCURRENCY STRESS TEST")
	fmt.Println(strings.Repeat("=", 80))

	// Test with increasingly higher concurrency
	concurrencyLevels := []int{10, 50, 100, 200, 500}

	for _, concurrency := range concurrencyLevels {
		fmt.Printf("\nTesting with %d concurrent users...\n", concurrency)

		originalConfig := lt.config
		lt.config.ConcurrentUsers = concurrency
		lt.config.RequestsPerUser = 5 // Keep requests per user consistent

		// Reset results
		lt.results = LoadTestResult{
			ErrorsByType: make(map[string]int),
		}

		lt.RunLoadTest()

		// Brief pause between tests
		time.Sleep(2 * time.Second)

		// Restore original config
		lt.config = originalConfig
	}
}

func main() {
	config := LoadTestConfig{
		BaseURL:         "http://localhost:8080",
		NumStudents:     1000,
		NumSections:     50,
		ConcurrentUsers: 100,
		RequestsPerUser: 10,
		TestDurationSec: 60,
		SectionCapacity: 30, // 30 seats per section = 1500 total seats
	}

	loadTester := NewLoadTester(config)
	loadTester.Initialize()

	fmt.Println("Course Registration System Load Test")
	fmt.Println("===================================")

	// Run standard load test
	loadTester.RunLoadTest()

	// Uncomment to run stress test
	// loadTester.RunConcurrencyStressTest()
}
