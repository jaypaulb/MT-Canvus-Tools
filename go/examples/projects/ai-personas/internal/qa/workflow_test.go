package qa

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerationWaitMessage(t *testing.T) {
	tests := []struct {
		model    string
		contains string
	}{
		{"gemini-2.5-flash-lite", "30 seconds"},
		{"gemini-2.5-flash", "60 seconds"},
		{"gemini-2.5-pro", "few minutes"},
		{"some-unknown-model", "please wait"},
		{"", "please wait"},
	}
	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			msg := generationWaitMessage(tt.model)
			assert.Contains(t, msg, tt.contains)
		})
	}
}

func TestFilterNonEmpty(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"all kept", []string{"a", "b"}, []string{"a", "b"}},
		{"drop empty", []string{"", "a", "", "b", ""}, []string{"a", "b"}},
		{"all empty", []string{"", ""}, []string{}},
		{"nil", nil, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, filterNonEmpty(tt.in))
		})
	}
}
