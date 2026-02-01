package favicon

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestPreCacheFavicons(t *testing.T) {
	// Create temp cache directory
	tempDir, err := os.MkdirTemp("", "favicon-precache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Track requests
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write([]byte("fake-favicon-data"))
	}))
	defer server.Close()

	fetcher := NewWithCache(tempDir)

	// Create test URLs - all with same host
	urls := []string{
		server.URL + "/page1",
		server.URL + "/page2",
		server.URL + "/page3",
	}

	// Pre-cache with 3 workers
	result := fetcher.PreCacheFavicons(urls, 3)

	// Verify results
	if result.Total != len(urls) {
		t.Errorf("Expected total=%d, got %d", len(urls), result.Total)
	}

	// All URLs processed (fetched or from cache after first fetch)
	if result.Fetched+result.Cached != len(urls) {
		t.Errorf("Expected fetched+cached=%d, got %d+%d=%d", 
			len(urls), result.Fetched, result.Cached, result.Fetched+result.Cached)
	}

	if result.Failed != 0 {
		t.Errorf("Expected failed=0, got %d", result.Failed)
	}

	// Verify cache files were created
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	// Should have 1 cache file (all URLs use same host)
	if len(entries) != 1 {
		t.Errorf("Expected 1 cache file, got %d", len(entries))
	}

	// Pre-cache again - should use cache
	requestCount = 0
	result2 := fetcher.PreCacheFavicons(urls, 3)

	if result2.Cached != len(urls) {
		t.Errorf("Expected all cached on second run, got cached=%d", result2.Cached)
	}

	if result2.Fetched != 0 {
		t.Errorf("Expected fetched=0 on second run, got %d", result2.Fetched)
	}

	if requestCount != 0 {
		t.Errorf("Expected 0 HTTP requests on second run (using cache), got %d", requestCount)
	}
}

func TestPreCacheFavicons_EmptyList(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "favicon-precache-empty-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	fetcher := NewWithCache(tempDir)
	result := fetcher.PreCacheFavicons([]string{}, 5)

	if result.Total != 0 {
		t.Errorf("Expected total=0 for empty list, got %d", result.Total)
	}
}

func TestPreCacheFavicons_WithFailures(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "favicon-precache-fail-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Server that returns errors
	errorCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorCount++
		if errorCount%2 == 0 {
			// Return 404 for every other request
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write([]byte("fake-favicon-data"))
	}))
	defer server.Close()

	fetcher := NewWithCache(tempDir)

	urls := []string{
		server.URL + "/page1",
		server.URL + "/page2",
	}

	result := fetcher.PreCacheFavicons(urls, 2)

	// One should succeed, one should fail (both use same host, so same favicon)
	// Actually, since they share the same host, the first success will cache it
	// and the second will use the cache
	if result.Total != len(urls) {
		t.Errorf("Expected total=%d, got %d", len(urls), result.Total)
	}

	// At least one should have been processed
	if result.Fetched+result.Cached+result.Failed != result.Total {
		t.Errorf("Sum of fetched+cached+failed should equal total")
	}
}

func TestPreCacheFavicons_InvalidURLs(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "favicon-precache-invalid-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	fetcher := NewWithCache(tempDir)

	urls := []string{
		"",                    // Empty
		"not-a-url",           // Invalid
		"ftp://example.com",   // Unsupported scheme
	}

	result := fetcher.PreCacheFavicons(urls, 2)

	if result.Total != len(urls) {
		t.Errorf("Expected total=%d, got %d", len(urls), result.Total)
	}

	// All should fail
	if result.Failed != 2 { // Empty URL is skipped, not counted as failed
		t.Errorf("Expected failed=2, got %d", result.Failed)
	}
}

func TestPreCacheFavicons_DefaultWorkers(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "favicon-precache-workers-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write([]byte("fake-favicon-data"))
	}))
	defer server.Close()

	fetcher := NewWithCache(tempDir)

	urls := []string{server.URL + "/page1"}

	// Test with workers=0 (should use default)
	result := fetcher.PreCacheFavicons(urls, 0)

	if result.Total != 1 {
		t.Errorf("Expected total=1, got %d", result.Total)
	}

	if result.Fetched != 1 {
		t.Errorf("Expected fetched=1, got %d", result.Fetched)
	}
}

func BenchmarkPreCacheFavicons(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "favicon-precache-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write([]byte("fake-favicon-data"))
	}))
	defer server.Close()

	fetcher := NewWithCache(tempDir)

	// Create 100 URLs
	urls := make([]string, 100)
	for i := 0; i < 100; i++ {
		urls[i] = fmt.Sprintf("%s/page%d", server.URL, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear cache between runs
		os.RemoveAll(tempDir)
		os.MkdirAll(tempDir, 0755)
		
		fetcher.PreCacheFavicons(urls, 10)
	}
}

func ExampleFetcher_PreCacheFavicons() {
	// Create temp directory for example
	tempDir, _ := os.MkdirTemp("", "example-*")
	defer os.RemoveAll(tempDir)

	fetcher := NewWithCache(tempDir)

	urls := []string{
		"https://github.com",
		"https://golang.org",
		"https://example.com",
	}

	// Pre-cache with 5 workers
	result := fetcher.PreCacheFavicons(urls, 5)

	fmt.Printf("Total: %d, Cached: %d, Fetched: %d, Failed: %d\n",
		result.Total, result.Cached, result.Fetched, result.Failed)
}
