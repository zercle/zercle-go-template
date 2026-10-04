//go:build unit

package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zercle/zercle-go-template/internal/infrastructure/config"
)

func TestParseByteSize(t *testing.T) {
	t.Parallel()

	valid := []struct {
		name string
		in   string
		want int64
	}{
		{"empty is unset", "", 0},
		{"whitespace is unset", "   ", 0},
		{"1K", "1K", 1024},
		{"1KB", "1KB", 1024},
		{"1k lowercase", "1k", 1024},
		{"1kb lowercase", "1kb", 1024},
		{"512B", "512B", 512},
		{"1M", "1M", 1048576},
		{"1MB", "1MB", 1048576},
		{"1G", "1G", 1073741824},
		{"1GB", "1GB", 1073741824},
		{"bare number", "1024", 1024},
		{"max int64 bare", "9223372036854775807", 9223372036854775807},
		{"surrounding whitespace", " 1MB ", 1048576},
		{"1KiB", "1KiB", 1024},
		{"1MiB", "1MiB", 1048576},
		{"1GiB", "1GiB", 1073741824},
		{"1kib lowercase", "1kib", 1024},
		{"512KiB", "512KiB", 524288},
	}
	for _, tc := range valid {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := config.ParseByteSize(tc.in)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}

	invalid := []struct {
		name string
		in   string
	}{
		{"2.5M non-integer", "2.5M"},
		{"negative", "-1M"},
		{"abc", "abc"},
		{"zero", "0"},
		{"overflow guard", "9999999999999G"},
	}
	for _, tc := range invalid {
		t.Run("invalid_"+tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := config.ParseByteSize(tc.in)
			require.Error(t, err, "unparseable limit must error, not silently disable")
		})
	}
}
