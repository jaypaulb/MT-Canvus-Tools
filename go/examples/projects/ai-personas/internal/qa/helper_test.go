package qa

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelperTracker(t *testing.T) {
	tr := NewHelperTracker()
	id, ok := tr.Take("q")
	assert.False(t, ok)
	assert.Equal(t, "", id)

	tr.Set("q", "helper-1")
	id, ok = tr.Take("q")
	assert.True(t, ok)
	assert.Equal(t, "helper-1", id)

	// Take is destructive: a second Take returns nothing.
	_, ok = tr.Take("q")
	assert.False(t, ok)
}

func TestExtractQuestion(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"plain", "Why are we here?", "Why are we here?"},
		{"strips helper prefix", "Generating answers --> Why are we here?", "Why are we here?"},
		{"strips please wait suffix", "Why are we here? Please wait...", "Why are we here?"},
		{"strips both", "Helper text --> Why? Please wait...", "Why?"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, extractQuestion(tt.in))
		})
	}
}

func TestHasQuestion(t *testing.T) {
	assert.True(t, hasQuestion("Is this a question?"))
	assert.True(t, hasQuestion("  spaces wrapped?   "))
	assert.False(t, hasQuestion("Not a question."))
	assert.False(t, hasQuestion(""))
}
