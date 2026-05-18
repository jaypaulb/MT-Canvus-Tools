package business

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
)

func TestRequiredNoteTitles(t *testing.T) {
	// Spec: 9 BMC sections.
	assert.Len(t, RequiredNoteTitles, 9)
	assert.Contains(t, RequiredNoteTitles, "VALUE PROPOSITIONS")
	assert.Contains(t, RequiredNoteTitles, "REVENUE STREAMS")
}

func TestAssemble_AllPresent(t *testing.T) {
	notes := make([]canvus.Note, 0, len(RequiredNoteTitles))
	for _, title := range RequiredNoteTitles {
		notes = append(notes, canvus.Note{Title: title, Text: "value for " + title})
	}
	anchors := []canvus.Anchor{{AnchorName: "Personas"}}
	bc, err := assemble(notes, anchors)
	require.NoError(t, err)
	assert.NotNil(t, bc.PersonasAnchor)
	assert.Empty(t, bc.Missing)
	for _, title := range RequiredNoteTitles {
		assert.Contains(t, bc.Body, title)
		assert.Contains(t, bc.Body, "value for "+title)
	}
}

func TestAssemble_MissingNotes(t *testing.T) {
	notes := []canvus.Note{
		{Title: "KEY PARTNERS", Text: "x"},
		{Title: "VALUE PROPOSITIONS", Text: "y"},
	}
	anchors := []canvus.Anchor{{AnchorName: "Personas"}}
	bc, err := assemble(notes, anchors)
	assert.Error(t, err)
	require.NotNil(t, bc)
	assert.NotEmpty(t, bc.Missing)
	// 9 required minus the 2 present = 7 missing.
	assert.Len(t, bc.Missing, 7)
}

func TestAssemble_PersonasAnchorMissing(t *testing.T) {
	notes := make([]canvus.Note, 0, len(RequiredNoteTitles))
	for _, title := range RequiredNoteTitles {
		notes = append(notes, canvus.Note{Title: title, Text: "v"})
	}
	bc, err := assemble(notes, nil)
	assert.Error(t, err)
	assert.Empty(t, bc.Missing)
	assert.Nil(t, bc.PersonasAnchor)
}

func TestAssemble_TitleCaseInsensitive(t *testing.T) {
	notes := []canvus.Note{}
	for _, title := range RequiredNoteTitles {
		// lower-case input should still match the upper-cased required list.
		notes = append(notes, canvus.Note{Title: " " + lowercase(title) + " ", Text: "v"})
	}
	anchors := []canvus.Anchor{{AnchorName: "Personas"}}
	bc, err := assemble(notes, anchors)
	require.NoError(t, err)
	assert.NotNil(t, bc.PersonasAnchor)
}

// lowercase is a tiny local helper so the test stays self-contained.
func lowercase(s string) string {
	out := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		out[i] = c
	}
	return string(out)
}
