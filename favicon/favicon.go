package favicon

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	// Maximum favicon size to download (1MB)
	maxFaviconSize = 1024 * 1024
	// HTTP timeout for favicon requests
	httpTimeout = 5 * time.Second
	// Default number of concurrent fetchers for pre-caching
	defaultWorkers = 10
	// Marker for cached failures (so we don't retry unreachable URLs)
	failedMarker = "FAILED"
)

// Fetcher handles fetching and encoding favicons
type Fetcher struct {
	client   *http.Client
	cacheDir string
}

// New creates a new Fetcher with default settings and cache directory (~/.arc-to-zen/favicons)
func New() *Fetcher {
	cache := defaultCacheDir()
	return &Fetcher{
		client: &http.Client{
			Timeout: httpTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Allow up to 5 redirects
				if len(via) >= 5 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		cacheDir: cache,
	}
}

// NewWithCache creates a new Fetcher with a custom cache directory
func NewWithCache(cacheDir string) *Fetcher {
	if cacheDir == "" {
		cacheDir = defaultCacheDir()
	}
	_ = os.MkdirAll(cacheDir, 0o755)
	return &Fetcher{
		client: &http.Client{Timeout: httpTimeout},
		cacheDir: cacheDir,
	}
}

// FetchAsDataURL fetches a favicon from the given URL and returns it as a data URL
// Returns empty string if fetch fails
func (f *Fetcher) FetchAsDataURL(pageURL string) string {
	faviconURL, err := f.buildFaviconURL(pageURL)
	if err != nil {
		return ""
	}

	// Try cache first
	if cached := f.readFromCache(pageURL); cached != "" {
		// Return empty for failed markers (tab imports without favicon)
		if cached == failedMarker {
			return ""
		}
		return cached
	}

	data, contentType, err := f.fetchFavicon(faviconURL)
	if err != nil {
		return ""
	}

	dataURL := f.encodeAsDataURL(data, contentType)
	f.writeToCache(pageURL, dataURL)
	return dataURL
}

// buildFaviconURL constructs the favicon URL from a page URL
func (f *Fetcher) buildFaviconURL(pageURL string) (string, error) {
	if pageURL == "" {
		return "", fmt.Errorf("empty URL")
	}

	// Parse the URL
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Only handle HTTP/HTTPS
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", fmt.Errorf("unsupported scheme: %s", parsedURL.Scheme)
	}

	// Build favicon URL: {scheme}://{host}/favicon.ico
	faviconURL := fmt.Sprintf("%s://%s/favicon.ico", parsedURL.Scheme, parsedURL.Host)
	return faviconURL, nil
}

// fetchFavicon downloads the favicon from the given URL
func (f *Fetcher) fetchFavicon(faviconURL string) ([]byte, string, error) {
	resp, err := f.client.Get(faviconURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch favicon: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Get content type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		// Default to x-icon if not specified
		contentType = "image/x-icon"
	}

	// Read body with size limit
	limitedReader := io.LimitReader(resp.Body, maxFaviconSize)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read favicon: %w", err)
	}

	// Check if we hit the size limit
	if len(data) >= maxFaviconSize {
		return nil, "", fmt.Errorf("favicon too large")
	}

	return data, contentType, nil
}

// encodeAsDataURL encodes favicon data as a data URL
func (f *Fetcher) encodeAsDataURL(data []byte, contentType string) string {
	if len(data) == 0 {
		return ""
	}

	// Normalize content type to a standard format
	contentType = normalizeContentType(contentType)

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(data)

	// Return data URL
	return fmt.Sprintf("data:%s;base64,%s", contentType, encoded)
}

// readFromCache tries to read a cached data URL for the page host
// Returns the cached value, or empty string if not cached
// Returns failedMarker if the URL previously failed (so we can skip it)
func (f *Fetcher) readFromCache(pageURL string) string {
	if f.cacheDir == "" {
		return ""
	}
	u, err := url.Parse(pageURL)
	if err != nil || u.Host == "" {
		return ""
	}
	_ = os.MkdirAll(f.cacheDir, 0o755)
	path := filepath.Join(f.cacheDir, sanitizeFilename(u.Host)+".txt")
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

// cacheFailure writes a failure marker so we don't retry unreachable URLs
func (f *Fetcher) cacheFailure(pageURL string) {
	if f.cacheDir == "" {
		return
	}
	u, err := url.Parse(pageURL)
	if err != nil || u.Host == "" {
		return
	}
	_ = os.MkdirAll(f.cacheDir, 0o755)
	path := filepath.Join(f.cacheDir, sanitizeFilename(u.Host)+".txt")
	_ = os.WriteFile(path, []byte(failedMarker), 0o644)
}

// writeToCache writes the data URL to cache
func (f *Fetcher) writeToCache(pageURL, dataURL string) {
	if f.cacheDir == "" || dataURL == "" {
		return
	}
	u, err := url.Parse(pageURL)
	if err != nil || u.Host == "" {
		return
	}
	_ = os.MkdirAll(f.cacheDir, 0o755)
	path := filepath.Join(f.cacheDir, sanitizeFilename(u.Host)+".txt")
	_ = os.WriteFile(path, []byte(dataURL), 0o644)
}

// defaultCacheDir returns ~/.arc-to-zen/favicons and ensures it exists
func defaultCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	dir := filepath.Join(home, ".arc-to-zen", "favicons")
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

// sanitizeFilename makes a safe filename from a host (includes port if present)
func sanitizeFilename(name string) string {
	// Replace path separators and spaces/colons with underscore
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", " ", "_")
	name = replacer.Replace(name)
	// Limit length to a reasonable size
	if len(name) > 200 {
		name = name[:200]
	}
	return name
}

// normalizeContentType normalizes the content type for common favicon formats
func normalizeContentType(contentType string) string {
	// Remove any charset or other parameters
	parts := strings.Split(contentType, ";")
	mimeType := strings.TrimSpace(parts[0])

	// Normalize common types
	switch mimeType {
	case "image/vnd.microsoft.icon", "image/ico", "image/icon":
		return "image/x-icon"
	case "image/png":
		return "image/png"
	case "image/jpeg", "image/jpg":
		return "image/jpeg"
	case "image/gif":
		return "image/gif"
	case "image/svg+xml":
		return "image/svg+xml"
	case "image/webp":
		return "image/webp"
	default:
		// Default to x-icon if unknown
		return "image/x-icon"
	}
}

// PreCacheResult contains statistics about the pre-caching operation
type PreCacheResult struct {
	Total      int // Total URLs processed
	Cached     int // URLs already in cache
	Fetched    int // URLs successfully fetched
	Failed     int // URLs that failed to fetch
}

// ProgressCallback is called during pre-caching to report progress
// processed is the number of URLs processed so far, total is the total count
type ProgressCallback func(processed, total int)

// PreCacheFavicons fetches favicons for multiple URLs in parallel
// Returns statistics about the operation
func (f *Fetcher) PreCacheFavicons(urls []string, workers int) *PreCacheResult {
	return f.PreCacheFaviconsWithProgress(urls, workers, nil)
}

// ClearCache removes all cached favicons
// Returns the number of files removed
func (f *Fetcher) ClearCache() (int, error) {
	if f.cacheDir == "" {
		return 0, fmt.Errorf("no cache directory configured")
	}

	entries, err := os.ReadDir(f.cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read cache directory: %w", err)
	}

	removed := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".txt") {
			path := filepath.Join(f.cacheDir, entry.Name())
			if err := os.Remove(path); err == nil {
				removed++
			}
		}
	}
	return removed, nil
}

// ClearFailedCache removes only the failed favicon markers from cache
// Returns the number of failed entries removed
func (f *Fetcher) ClearFailedCache() (int, error) {
	if f.cacheDir == "" {
		return 0, fmt.Errorf("no cache directory configured")
	}

	entries, err := os.ReadDir(f.cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read cache directory: %w", err)
	}

	removed := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".txt") {
			path := filepath.Join(f.cacheDir, entry.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			if string(content) == failedMarker {
				if err := os.Remove(path); err == nil {
					removed++
				}
			}
		}
	}
	return removed, nil
}

// GetCacheStats returns statistics about the current cache
func (f *Fetcher) GetCacheStats() (total, successful, failed int, err error) {
	if f.cacheDir == "" {
		return 0, 0, 0, fmt.Errorf("no cache directory configured")
	}

	entries, err := os.ReadDir(f.cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, 0, nil
		}
		return 0, 0, 0, fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".txt") {
			total++
			path := filepath.Join(f.cacheDir, entry.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			if string(content) == failedMarker {
				failed++
			} else {
				successful++
			}
		}
	}
	return total, successful, failed, nil
}

// PreCacheFaviconsWithProgress fetches favicons for multiple URLs in parallel
// with an optional progress callback
func (f *Fetcher) PreCacheFaviconsWithProgress(urls []string, workers int, progress ProgressCallback) *PreCacheResult {
	if workers <= 0 {
		workers = defaultWorkers
	}

	result := &PreCacheResult{
		Total: len(urls),
	}

	if len(urls) == 0 {
		return result
	}

	// Create channels for work distribution
	urlChan := make(chan string, len(urls))
	var wg sync.WaitGroup
	var mu sync.Mutex
	processed := 0

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pageURL := range urlChan {
				if pageURL == "" {
					mu.Lock()
					processed++
					if progress != nil {
						progress(processed, result.Total)
					}
					mu.Unlock()
					continue
				}

				// Check if already cached (includes failed markers)
				if cached := f.readFromCache(pageURL); cached != "" {
					mu.Lock()
					result.Cached++
					processed++
					if progress != nil {
						progress(processed, result.Total)
					}
					mu.Unlock()
					continue
				}

				// Fetch favicon
				faviconURL, err := f.buildFaviconURL(pageURL)
				if err != nil {
					f.cacheFailure(pageURL) // Cache the failure
					mu.Lock()
					result.Failed++
					processed++
					if progress != nil {
						progress(processed, result.Total)
					}
					mu.Unlock()
					continue
				}

				data, contentType, err := f.fetchFavicon(faviconURL)
				if err != nil {
					f.cacheFailure(pageURL) // Cache the failure
					mu.Lock()
					result.Failed++
					processed++
					if progress != nil {
						progress(processed, result.Total)
					}
					mu.Unlock()
					continue
				}

				dataURL := f.encodeAsDataURL(data, contentType)
				mu.Lock()
				if dataURL != "" {
					f.writeToCache(pageURL, dataURL)
					result.Fetched++
				} else {
					result.Failed++
				}
				processed++
				if progress != nil {
					progress(processed, result.Total)
				}
				mu.Unlock()
			}
		}()
	}

	// Send URLs to workers
	for _, url := range urls {
		urlChan <- url
	}
	close(urlChan)

	// Wait for all workers to finish
	wg.Wait()

	return result
}
