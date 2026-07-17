package util

import (
	"bytes"
	"compress/gzip"
	"io"
)

// CompressString gzips a string and returns the resulting byte array
func CompressString(input string) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	if _, err := writer.Write([]byte(input)); err != nil {
		// close in case of a write error and ignore anything that might add to original
		_ = writer.Close()
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decompress takes gzipped bytes and returns the original string
func Decompress(compressed []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	// Ensure the reader is closed to free up resources
	defer func(reader *gzip.Reader) {
		_ = reader.Close()
	}(reader)

	// Read all decompressed data
	decompressedBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return decompressedBytes, nil
}
