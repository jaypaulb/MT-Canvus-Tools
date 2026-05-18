package monitor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsWhiteBackground(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"#FFFFFFFF", true},
		{"#ffffffff", true},
		{"#FFFFFF", true},
		{"  #FFFFFFFF  ", true},
		{"#000000", false},
		{"#fffffe", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.want, isWhiteBackground(tt.in))
		})
	}
}
