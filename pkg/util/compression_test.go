package util

import "testing"

func Test_CompressDecompressString(t *testing.T) {
	cases := []struct {
		name       string
		testString string
	}{
		{"compress-decompress empty string", ""},
		{"compress-decompress single-character string", "s"},
		{"compress-decompress regular string", "s21a dadsas _, das !?- wadas14665+"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			compressedBytes, err := CompressString(tc.testString)
			if err != nil {
				t.Fatalf("failed to compress test string: %v", err)
			}

			decompressed, err := Decompress(compressedBytes)
			if err != nil {
				t.Fatalf("failed to decompress test string: %v", err)
			}

			if string(decompressed) != tc.testString {
				t.Fatalf("failed to compress-decompress a test string: got %q, want %q", string(decompressed), tc.testString)
			}
		})
	}
}
