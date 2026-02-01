package mozlz4

import (
	"bytes"
	"testing"
)

func TestCompressDecompress(t *testing.T) {
	testData := []byte(`{"test":"data","nested":{"value":123},"array":[1,2,3,4,5]}`)

	// Compress
	compressed, err := Compress(testData)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	// Verify header
	if len(compressed) < headerSize+sizeBytes {
		t.Fatalf("Compressed data too short: %d bytes", len(compressed))
	}

	if string(compressed[:headerSize]) != headerMagic {
		t.Fatalf("Invalid header: %q", string(compressed[:headerSize]))
	}

	// Decompress
	decompressed, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}

	// Verify data matches
	if !bytes.Equal(testData, decompressed) {
		t.Fatalf("Data mismatch:\nExpected: %s\nGot: %s", testData, decompressed)
	}
}

func TestDecompressInvalidHeader(t *testing.T) {
	invalidData := []byte("invalid header data")
	_, err := Decompress(invalidData)
	if err == nil {
		t.Fatal("Expected error for invalid header, got nil")
	}
}

func TestDecompressTooShort(t *testing.T) {
	shortData := []byte("short")
	_, err := Decompress(shortData)
	if err == nil {
		t.Fatal("Expected error for too-short data, got nil")
	}
}

func TestCompressEmpty(t *testing.T) {
	emptyData := []byte("")
	compressed, err := Compress(emptyData)
	if err != nil {
		t.Fatalf("Compress empty failed: %v", err)
	}

	decompressed, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress empty failed: %v", err)
	}

	if len(decompressed) != 0 {
		t.Fatalf("Expected empty result, got %d bytes", len(decompressed))
	}
}

func TestCompressLargeData(t *testing.T) {
	// Create large JSON-like data
	largeData := bytes.Repeat([]byte(`{"key":"value","number":12345},`), 1000)

	compressed, err := Compress(largeData)
	if err != nil {
		t.Fatalf("Compress large data failed: %v", err)
	}

	// Verify compression actually reduced size
	if len(compressed) >= len(largeData) {
		t.Logf("Warning: compressed size (%d) >= original size (%d)", len(compressed), len(largeData))
	}

	decompressed, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress large data failed: %v", err)
	}

	if !bytes.Equal(largeData, decompressed) {
		t.Fatal("Large data mismatch after compress/decompress")
	}
}
