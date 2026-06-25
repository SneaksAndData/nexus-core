package util

import (
	"bytes"
	"compress/gzip"
	"log"
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
		log.Fatal(err)
	}

	return buf.Bytes(), nil
}
