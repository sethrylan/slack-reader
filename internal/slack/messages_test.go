package slack_test

import (
	"testing"

	"github.com/sethrylan/slack-reader/internal/slack"
)

func TestNormalizeTimestamp(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1772147524.763449", "1772147524.763449"},
		{"1772147524763449", "1772147524.763449"},
		{"1234567890.000001", "1234567890.000001"},
		{"1234567890000001", "1234567890.000001"},
		// Short or unusual values pass through unchanged
		{"12345", "12345"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := slack.NormalizeTimestamp(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeTimestamp(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
