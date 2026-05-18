package atom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskKey(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"short", "abc", "abc"},
		{"four", "abcd", "abcd"},
		{"five", "abcde", "*bcde"},
		{"long", "sk-abcdef123456", "***********3456"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, MaskKey(tt.in))
		})
	}
}

func TestStripMarkdownCodeBlock(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"plain", "hello", "hello"},
		{"json fence", "```json\n{\"a\":1}\n```", "{\"a\":1}"},
		{"bare fence", "```\n[1,2,3]\n```", "[1,2,3]"},
		{"surrounding ws", "   ```json\n42\n```   ", "42"},
		{"no fence with backticks", "this `is` fine", "this `is` fine"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, StripMarkdownCodeBlock(tt.in))
		})
	}
}
