package favicon

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCacheReadWrite(t *testing.T) {
	tmp := t.TempDir()
	f := NewWithCache(tmp)

	// Prepare a fake page URL
	pageURL := "https://example.com/path"
	dataURL := "data:image/png;base64,iVBORw0KGgo="

	// Nothing cached yet
	if got := f.readFromCache(pageURL); got != "" {
		t.Fatalf("expected empty cache, got %q", got)
	}

	// Write and then read
	f.writeToCache(pageURL, dataURL)
	if got := f.readFromCache(pageURL); got != dataURL {
		t.Fatalf("cache mismatch: want %q got %q", dataURL, got)
	}

	// Verify filename is based on host
	u := filepath.Join(tmp, "example.com.txt")
	if _, err := os.Stat(u); err != nil {
		t.Fatalf("expected cache file at %s: %v", u, err)
	}
}

func TestFetchAsDataURL_UsesCacheOnSubsequentCalls(t *testing.T) {
	// Set up a server that returns a favicon once
	testData := []byte{0x89, 0x50, 0x4E, 0x47} // PNG header
	served := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/favicon.ico" {
			served++
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusOK)
			w.Write(testData)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))

	tmp := t.TempDir()
	f := NewWithCache(tmp)

	// First call should hit the network and populate cache
	res1 := f.FetchAsDataURL(ts.URL + "/page")
	if !strings.HasPrefix(res1, "data:image/png;base64,") {
		t.Fatalf("unexpected data URL: %q", res1)
	}
	if served == 0 {
		t.Fatalf("expected network to be used on first call")
	}

	// Close server to ensure subsequent call cannot hit network
	ts.Close()

	// Second call should use cache and still return the data URL
	res2 := f.FetchAsDataURL(ts.URL + "/page")
	if res2 == "" {
		t.Fatalf("expected cached data URL, got empty string")
	}
}
