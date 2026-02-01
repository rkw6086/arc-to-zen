package favicon

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildFaviconURL(t *testing.T) {
	f := New()

	tests := []struct {
		name        string
		pageURL     string
		expected    string
		expectError bool
	}{
		{
			name:        "https URL",
			pageURL:     "https://example.com/page",
			expected:    "https://example.com/favicon.ico",
			expectError: false,
		},
		{
			name:        "http URL",
			pageURL:     "http://example.com/page",
			expected:    "http://example.com/favicon.ico",
			expectError: false,
		},
		{
			name:        "URL with port",
			pageURL:     "https://example.com:8080/page",
			expected:    "https://example.com:8080/favicon.ico",
			expectError: false,
		},
		{
			name:        "URL with path",
			pageURL:     "https://example.com/path/to/page",
			expected:    "https://example.com/favicon.ico",
			expectError: false,
		},
		{
			name:        "empty URL",
			pageURL:     "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid scheme",
			pageURL:     "ftp://example.com/page",
			expected:    "",
			expectError: true,
		},
		{
			name:        "chrome extension",
			pageURL:     "chrome-extension://abc123/page.html",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := f.buildFaviconURL(tt.pageURL)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

func TestNormalizeContentType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"image/x-icon", "image/x-icon"},
		{"image/vnd.microsoft.icon", "image/x-icon"},
		{"image/ico", "image/x-icon"},
		{"image/icon", "image/x-icon"},
		{"image/png", "image/png"},
		{"image/jpeg", "image/jpeg"},
		{"image/jpg", "image/jpeg"},
		{"image/gif", "image/gif"},
		{"image/svg+xml", "image/svg+xml"},
		{"image/webp", "image/webp"},
		{"image/png; charset=utf-8", "image/png"},
		{"unknown/type", "image/x-icon"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeContentType(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestEncodeAsDataURL(t *testing.T) {
	f := New()

	tests := []struct {
		name        string
		data        []byte
		contentType string
		expected    string
	}{
		{
			name:        "PNG data",
			data:        []byte{0x89, 0x50, 0x4E, 0x47}, // PNG header
			contentType: "image/png",
			expected:    "data:image/png;base64,iVBORw==",
		},
		{
			name:        "empty data",
			data:        []byte{},
			contentType: "image/png",
			expected:    "",
		},
		{
			name:        "x-icon data",
			data:        []byte{0x00, 0x00, 0x01, 0x00},
			contentType: "image/x-icon",
			expected:    "data:image/x-icon;base64,AAABAA==",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.encodeAsDataURL(tt.data, tt.contentType)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFetchFavicon(t *testing.T) {
	// Create test server
	testData := []byte{0x89, 0x50, 0x4E, 0x47} // PNG header
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/favicon.ico" {
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusOK)
			w.Write(testData)
		} else if r.URL.Path == "/notfound" {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	f := New()

	t.Run("successful fetch", func(t *testing.T) {
		data, contentType, err := f.fetchFavicon(ts.URL + "/favicon.ico")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if contentType != "image/png" {
			t.Errorf("expected content type image/png, got %q", contentType)
		}
		if len(data) != len(testData) {
			t.Errorf("expected %d bytes, got %d", len(testData), len(data))
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		_, _, err := f.fetchFavicon(ts.URL + "/notfound")
		if err == nil {
			t.Error("expected error for 404, got none")
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		_, _, err := f.fetchFavicon("http://this-domain-does-not-exist-12345.com/favicon.ico")
		if err == nil {
			t.Error("expected error for invalid URL, got none")
		}
	})
}

func TestFetchAsDataURL(t *testing.T) {
	// Create test server with a valid favicon
	testData := []byte{0x89, 0x50, 0x4E, 0x47} // PNG header
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/favicon.ico" {
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusOK)
			w.Write(testData)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	f := New()

	t.Run("successful fetch and encode", func(t *testing.T) {
		result := f.FetchAsDataURL(ts.URL + "/page")
		if result == "" {
			t.Error("expected non-empty data URL")
		}
		if !strings.HasPrefix(result, "data:image/png;base64,") {
			t.Errorf("expected data URL to start with 'data:image/png;base64,', got %q", result)
		}

		// Verify it's valid base64 by decoding
		parts := strings.Split(result, ",")
		if len(parts) != 2 {
			t.Errorf("expected 2 parts in data URL, got %d", len(parts))
		} else {
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				t.Errorf("failed to decode base64: %v", err)
			}
			if len(decoded) != len(testData) {
				t.Errorf("decoded data length mismatch: expected %d, got %d", len(testData), len(decoded))
			}
		}
	})

	t.Run("favicon not found", func(t *testing.T) {
		// Create a separate server that doesn't serve favicon.ico
		ts404 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts404.Close()

		result := f.FetchAsDataURL(ts404.URL + "/nonexistent")
		if result != "" {
			t.Errorf("expected empty string for 404, got %q", result)
		}
	})

	t.Run("invalid page URL", func(t *testing.T) {
		result := f.FetchAsDataURL("not-a-url")
		if result != "" {
			t.Errorf("expected empty string for invalid URL, got %q", result)
		}
	})

	t.Run("unsupported scheme", func(t *testing.T) {
		result := f.FetchAsDataURL("ftp://example.com/page")
		if result != "" {
			t.Errorf("expected empty string for FTP URL, got %q", result)
		}
	})
}
