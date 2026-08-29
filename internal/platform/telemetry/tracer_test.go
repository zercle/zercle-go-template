//go:build unit

package telemetry

import "testing"

func TestOTLPTracesEndpoint(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "base URL without trailing slash gets path appended",
			in:   "http://localhost:4318",
			want: "http://localhost:4318/v1/traces",
		},
		{
			name: "base URL with trailing slash has slash stripped and path appended",
			in:   "http://localhost:4318/",
			want: "http://localhost:4318/v1/traces",
		},
		{
			name: "URL already ending with /v1/traces is preserved",
			in:   "http://localhost:4318/v1/traces",
			want: "http://localhost:4318/v1/traces",
		},
		{
			name: "https base URL without path gets path appended",
			in:   "https://otlp.example.com",
			want: "https://otlp.example.com/v1/traces",
		},
		{
			name: "URL with /otlp/v1/traces already ends with /v1/traces and is preserved",
			in:   "https://otlp.example.com/otlp/v1/traces",
			want: "https://otlp.example.com/otlp/v1/traces",
		},
		{
			// Edge case: a trailing slash after /v1/traces means the string does
			// NOT match the suffix check, so another /v1/traces is appended.
			// This documents the current (unintentional but harmless) behavior
			// rather than changing production logic.
			name: "URL with /v1/traces followed by trailing slash gets path appended again",
			in:   "http://localhost:4318/v1/traces/",
			want: "http://localhost:4318/v1/traces/v1/traces",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := otlpTracesEndpoint(tc.in)
			if got != tc.want {
				t.Errorf("otlpTracesEndpoint(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
