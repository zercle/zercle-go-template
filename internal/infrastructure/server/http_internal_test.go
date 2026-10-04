//go:build unit

package server

import "testing"

func TestParseBodyLimitBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want int64
	}{
		{"empty", "", 0},
		{"only whitespace", "   ", 0},
		{"1K", "1K", 1024},
		{"1KB", "1KB", 1024},
		{"1k lowercase", "1k", 1024},
		{"1kb lowercase", "1kb", 1024},
		{"512B", "512B", 512},
		{"1M", "1M", 1048576},
		{"1MB", "1MB", 1048576},
		{"1G", "1G", 1073741824},
		{"1GB", "1GB", 1073741824},
		{"2.5M non-integer", "2.5M", 0},
		{"negative", "-1M", 0},
		{"abc", "abc", 0},
		{"bare number", "1024", 1024},
		{"overflow guard", "9999999999999G", 0},
		{"max int64 bare", "9223372036854775807", 9223372036854775807},
		{"surrounding whitespace", " 1MB ", 1048576},
		{"1KiB", "1KiB", 1024},
		{"1MiB", "1MiB", 1048576},
		{"1GiB", "1GiB", 1073741824},
		{"1kib lowercase", "1kib", 1024},
		{"512KiB", "512KiB", 524288},
		{"2MiB", "2MiB", 2097152},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := parseBodyLimitBytes(tc.in); got != tc.want {
				t.Errorf("parseBodyLimitBytes(%q) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}
