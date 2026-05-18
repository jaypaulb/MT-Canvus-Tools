package qa

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountSuccess(t *testing.T) {
	tests := []struct {
		name    string
		answers []string
		errs    []error
		want    int
	}{
		{"all good", []string{"a", "b", "c"}, []error{nil, nil, nil}, 3},
		{"empty answer counts as failure",
			[]string{"a", "", "c"}, []error{nil, nil, nil}, 2},
		{"error counts as failure",
			[]string{"a", "b", "c"}, []error{nil, fmt.Errorf("oops"), nil}, 2},
		{"none", []string{"", "", ""}, []error{nil, nil, nil}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, countSuccess(tt.answers, tt.errs))
		})
	}
}

func TestJoinSemi(t *testing.T) {
	assert.Equal(t, "", joinSemi(nil))
	assert.Equal(t, "a", joinSemi([]string{"a"}))
	assert.Equal(t, "a; b; c", joinSemi([]string{"a", "b", "c"}))
}
