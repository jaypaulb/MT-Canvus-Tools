package persona

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoreRoundTrip(t *testing.T) {
	s := NewStore()
	assert.False(t, s.Has("q1"))
	_, ok := s.Get("q1")
	assert.False(t, ok)

	s.Set("q1", []string{"n1", "n2", "n3"})
	assert.True(t, s.Has("q1"))
	ids, ok := s.Get("q1")
	assert.True(t, ok)
	assert.Equal(t, []string{"n1", "n2", "n3"}, ids)
}

func TestStore_DefensiveCopy(t *testing.T) {
	s := NewStore()
	original := []string{"a", "b"}
	s.Set("q1", original)

	// Mutating the source slice after Set must not affect stored IDs.
	original[0] = "MUTATED"
	got, _ := s.Get("q1")
	assert.Equal(t, []string{"a", "b"}, got)

	// Mutating the returned slice must not affect the stored copy either.
	got[0] = "ALSO_MUTATED"
	got2, _ := s.Get("q1")
	assert.Equal(t, []string{"a", "b"}, got2)
}
