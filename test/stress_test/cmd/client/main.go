package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

var companies = []string{"discount_co", "shipping_co", "validator_co", "transform_co", "tax_co"}
var categories = []string{"electronics", "clothing", "food", "books", "toys", "home", "sports"}
var zones = []string{"domestic", "domestic", "domestic", "regional", "international"}
var stateList = []string{"CA", "NY", "TX", "FL", "IL", "OH", "WA", "PA", "NJ", "MA"}
var streets = []string{"123 Main St", "456 Oak Ave", "789 Elm Blvd", "1010 Pine Rd", "2020 Maple Dr"}
var cities = []string{"Springfield", "Portland", "Austin", "Miami", "Chicago", "Columbus"}
var tagPool = []string{
	"sale", "clearance", "vip", "member", "seasonal", "new-arrival",
	"eco-friendly", "organic", "imported", "fragile", "hazmat",
	"express", "overnight", "bulk", "limited-edition", "handmade",
	"premium-quality", "best-seller", "trending", "green-certified",
}

// Fixed pools of names and zips to avoid bloating RAGE's global string intern table.
// RAGE interns all strings <=64 chars; unique names/zips would leak memory.
var itemNames []string
var zipCodes []string

func init() {
	itemNames = make([]string, 200)
	for i := range itemNames {
		itemNames[i] = fmt.Sprintf("Item-%d", i)
	}
	zipCodes = make([]string, 100)
	for i := range zipCodes {
		zipCodes[i] = fmt.Sprintf("%05d", 10000+i*900)
	}
}

type request struct {
	CompanyID string         `json:"company_id"`
	Item      map[string]any `json:"item"`
}

func randomTags(rng *rand.Rand) []string {
	n := rng.Intn(6) + 1 // 1-6 tags
	tags := make([]string, n)
	for i := range n {
		tags[i] = tagPool[rng.Intn(len(tagPool))]
	}
	return tags
}

func randomAddress(rng *rand.Rand) map[string]any {
	return map[string]any{
		"street": streets[rng.Intn(len(streets))],
		"city":   cities[rng.Intn(len(cities))],
		"state":  stateList[rng.Intn(len(stateList))],
		"zip":    zipCodes[rng.Intn(len(zipCodes))],
		"zone":   zones[rng.Intn(len(zones))],
	}
}

func randomItem(rng *rand.Rand) request {
	return request{
		CompanyID: companies[rng.Intn(len(companies))],
		Item: map[string]any{
			"name":     itemNames[rng.Intn(len(itemNames))],
			"price":    float64(rng.Intn(99900)+100) / 100.0,
			"quantity": rng.Intn(100) + 1,
			"category": categories[rng.Intn(len(categories))],
			"weight":   float64(rng.Intn(4990)+10) / 100.0,
			"tags":     randomTags(rng),
			"address":  randomAddress(rng),
		},
	}
}

type companyStats struct {
	mu       sync.Mutex
	latency  []time.Duration
	errors   int64
	requests int64
}

func (cs *companyStats) add(d time.Duration, isErr bool) {
	cs.mu.Lock()
	cs.latency = append(cs.latency, d)
	cs.requests++
	if isErr {
		cs.errors++
	}
	cs.mu.Unlock()
}

func validateResponse(company string, body map[string]any) error {
	// Every response must have "company" matching what we sent
	got, ok := body["company"]
	if !ok {
		return fmt.Errorf("missing 'company' in response")
	}
	if gotStr, ok := got.(string); !ok || gotStr != company {
		return fmt.Errorf("company mismatch: want %s, got %v", company, got)
	}
	// Must have item_name or original_name
	_, hasItemName := body["item_name"]
	_, hasOrigName := body["original_name"]
	if !hasItemName && !hasOrigName {
		return fmt.Errorf("missing 'item_name'/'original_name' in response")
	}
	// Company-specific checks
	switch company {
	case "discount_co":
		if _, ok := body["total"]; !ok {
			return fmt.Errorf("discount_co: missing 'total'")
		}
		if _, ok := body["breakdown"]; !ok {
			return fmt.Errorf("discount_co: missing 'breakdown'")
		}
	case "shipping_co":
		if _, ok := body["shipping"]; !ok {
			return fmt.Errorf("shipping_co: missing 'shipping'")
		}
		if _, ok := body["estimated_days"]; !ok {
			return fmt.Errorf("shipping_co: missing 'estimated_days'")
		}
	case "validator_co":
		if _, ok := body["valid"]; !ok {
			return fmt.Errorf("validator_co: missing 'valid'")
		}
		if _, ok := body["risk_level"]; !ok {
			return fmt.Errorf("validator_co: missing 'risk_level'")
		}
	case "transform_co":
		if _, ok := body["sku"]; !ok {
			return fmt.Errorf("transform_co: missing 'sku'")
		}
		if _, ok := body["tag_categories"]; !ok {
			return fmt.Errorf("transform_co: missing 'tag_categories'")
		}
	case "tax_co":
		if _, ok := body["total_tax"]; !ok {
			return fmt.Errorf("tax_co: missing 'total_tax'")
		}
		if _, ok := body["breakdown"]; !ok {
			return fmt.Errorf("tax_co: missing 'breakdown'")
		}
	}
	return nil
}

func main() {
	addr := flag.String("addr", "http://localhost:8080", "server address")
	totalRequests := flag.Int("requests", 500000, "total number of requests")
	workers := flag.Int("workers", 50, "number of concurrent workers")
	triggerGC := flag.Bool("gc", false, "trigger server GC before starting")
	flag.Parse()

	transport := &http.Transport{
		MaxIdleConns:        *workers + 10,
		MaxIdleConnsPerHost: *workers + 10,
		MaxConnsPerHost:     *workers + 10,
		IdleConnTimeout:     90 * time.Second,
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	if *triggerGC {
		resp, err := client.Post(*addr+"/gc", "application/json", nil)
		if err == nil {
			resp.Body.Close()
			fmt.Println("Server GC triggered.")
		}
	}

	fmt.Printf("Stress test: %d requests, %d workers -> %s\n", *totalRequests, *workers, *addr)

	work := make(chan int, *totalRequests)
	for i := range *totalRequests {
		work <- i
	}
	close(work)

	var (
		errCount       atomic.Int64
		validationErrs atomic.Int64
		mu             sync.Mutex
		allLatency     []time.Duration
	)

	perCompany := make(map[string]*companyStats)
	for _, c := range companies {
		perCompany[c] = &companyStats{}
	}

	start := time.Now()
	var wg sync.WaitGroup

	for w := range *workers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))
			var latencies []time.Duration

			for range work {
				req := randomItem(rng)
				body, _ := json.Marshal(req)

				t0 := time.Now()
				resp, err := client.Post(*addr+"/process", "application/json", bytes.NewReader(body))
				elapsed := time.Since(t0)
				latencies = append(latencies, elapsed)

				isErr := false
				if err != nil {
					errCount.Add(1)
					isErr = true
					perCompany[req.CompanyID].add(elapsed, true)
					continue
				}

				respBody, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					errCount.Add(1)
					isErr = true
				} else {
					// Validate response
					var result map[string]any
					if jsonErr := json.Unmarshal(respBody, &result); jsonErr != nil {
						validationErrs.Add(1)
						isErr = true
					} else if valErr := validateResponse(req.CompanyID, result); valErr != nil {
						validationErrs.Add(1)
						isErr = true
					}
				}
				perCompany[req.CompanyID].add(elapsed, isErr)
			}

			mu.Lock()
			allLatency = append(allLatency, latencies...)
			mu.Unlock()
		}(w)
	}

	wg.Wait()
	totalTime := time.Since(start)

	// Compute stats
	n := len(allLatency)
	errs := errCount.Load()
	valErrs := validationErrs.Load()
	if n == 0 {
		fmt.Println("No requests completed.")
		os.Exit(1)
	}

	sort.Slice(allLatency, func(i, j int) bool { return allLatency[i] < allLatency[j] })

	p50 := allLatency[n*50/100]
	p95 := allLatency[n*95/100]
	p99 := allLatency[n*99/100]
	rps := float64(n) / totalTime.Seconds()

	fmt.Println()
	fmt.Println("=== Overall Results ===")
	fmt.Printf("Total requests:     %d\n", n)
	fmt.Printf("Total time:         %s\n", totalTime.Round(time.Millisecond))
	fmt.Printf("Requests/sec:       %.1f\n", rps)
	fmt.Printf("HTTP errors:        %d (%.2f%%)\n", errs, float64(errs)/float64(n)*100)
	fmt.Printf("Validation errors:  %d (%.2f%%)\n", valErrs, float64(valErrs)/float64(n)*100)
	fmt.Println()
	fmt.Printf("Latency p50:        %s\n", p50.Round(time.Microsecond))
	fmt.Printf("Latency p95:        %s\n", p95.Round(time.Microsecond))
	fmt.Printf("Latency p99:        %s\n", p99.Round(time.Microsecond))
	fmt.Printf("Latency min:        %s\n", allLatency[0].Round(time.Microsecond))
	fmt.Printf("Latency max:        %s\n", allLatency[n-1].Round(time.Microsecond))

	// Per-company breakdown
	fmt.Println()
	fmt.Println("=== Per-Company Breakdown ===")
	fmt.Printf("%-14s %7s %6s %10s %10s %10s\n", "Company", "Reqs", "Errs", "p50", "p95", "p99")
	fmt.Println("---------------------------------------------------------------")
	for _, c := range companies {
		cs := perCompany[c]
		cs.mu.Lock()
		cn := len(cs.latency)
		if cn == 0 {
			cs.mu.Unlock()
			continue
		}
		sort.Slice(cs.latency, func(i, j int) bool { return cs.latency[i] < cs.latency[j] })
		cp50 := cs.latency[cn*50/100]
		cp95 := cs.latency[cn*95/100]
		cp99 := cs.latency[cn*99/100]
		fmt.Printf("%-14s %7d %6d %10s %10s %10s\n",
			c, cs.requests, cs.errors,
			cp50.Round(time.Microsecond),
			cp95.Round(time.Microsecond),
			cp99.Round(time.Microsecond))
		cs.mu.Unlock()
	}
}
