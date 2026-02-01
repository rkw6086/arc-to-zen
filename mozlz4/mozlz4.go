package mozlz4

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/pierrec/lz4/v4"
)

const (
	// Mozilla LZ4 header: "mozLz40\0"
	headerMagic = "mozLz40\x00"
	headerSize  = 8
	sizeBytes   = 4
)

// Decompress decompresses Mozilla LZ4 format data
// Format: 8-byte header + 4-byte uncompressed size (LE) + LZ4 block data
func Decompress(data []byte) ([]byte, error) {
	if len(data) < headerSize+sizeBytes {
		return nil, fmt.Errorf("data too short: %d bytes", len(data))
	}

	// Verify header
	header := string(data[:headerSize])
	if header != headerMagic {
		return nil, fmt.Errorf("invalid Mozilla LZ4 header: expected %q, got %q", headerMagic, header)
	}

	// Read uncompressed size (little-endian)
	uncompressedSize := binary.LittleEndian.Uint32(data[headerSize : headerSize+sizeBytes])

	// Extract compressed content
	compressedData := data[headerSize+sizeBytes:]

	// Decompress
	decompressed := make([]byte, uncompressedSize)
	n, err := lz4.UncompressBlock(compressedData, decompressed)
	if err != nil {
		return nil, fmt.Errorf("LZ4 decompression failed: %w", err)
	}

	return decompressed[:n], nil
}

// Compress compresses data to Mozilla LZ4 format
// Format: 8-byte header + 4-byte uncompressed size (LE) + LZ4 block data
func Compress(data []byte) ([]byte, error) {
	// Calculate max compressed size
	maxCompressed := lz4.CompressBlockBound(len(data))
	compressed := make([]byte, maxCompressed)

	// Compress
	n, err := lz4.CompressBlock(data, compressed, nil)
	if err != nil {
		return nil, fmt.Errorf("LZ4 compression failed: %w", err)
	}
	compressed = compressed[:n]

	// Build Mozilla LZ4 format
	var buf bytes.Buffer

	// Write header
	if _, err := buf.WriteString(headerMagic); err != nil {
		return nil, err
	}

	// Write uncompressed size (little-endian)
	sizeHeader := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeHeader, uint32(len(data)))
	if _, err := buf.Write(sizeHeader); err != nil {
		return nil, err
	}

	// Write compressed data
	if _, err := buf.Write(compressed); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DecompressReader decompresses Mozilla LZ4 data from a reader
func DecompressReader(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return Decompress(data)
}
