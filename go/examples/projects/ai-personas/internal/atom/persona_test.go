package atom

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgeString_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"string", `"42"`, "42"},
		{"int", `42`, "42"},
		{"float truncates", `42.7`, "42"},
		{"empty string", `""`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a AgeString
			require.NoError(t, json.Unmarshal([]byte(tt.in), &a))
			assert.Equal(t, tt.want, string(a))
		})
	}
}

func TestGoalsString_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"string", `"buy a car"`, "buy a car"},
		{"array", `["a","b","c"]`, "a\nb\nc"},
		{"empty array", `[]`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var g GoalsString
			require.NoError(t, json.Unmarshal([]byte(tt.in), &g))
			assert.Equal(t, tt.want, string(g))
		})
	}
}

func TestFormatParseRoundTrip(t *testing.T) {
	p := Persona{
		Name:        "Alice",
		Role:        "Procurement Lead",
		Description: "Buys mining equipment",
		Background:  "BSc Mining Engineering",
		Goals:       "lower TCO, faster ROI",
		Age:         "42",
		Sex:         "F",
		Race:        "Hispanic",
	}
	got := ParsePersonaNote(FormatPersonaNote(p))
	assert.Equal(t, p, got)
}

func TestParsePersonaNote_Malformed(t *testing.T) {
	// Missing fields return a zero Persona — the workflow tolerates this
	// and treats the persona as failed downstream.
	got := ParsePersonaNote("this is not a persona note")
	assert.Equal(t, Persona{}, got)
}

func TestGenerateSystemPrompt_ContainsKeyFields(t *testing.T) {
	p := Persona{Name: "Bob", Role: "CTO", Goals: "ship it"}
	prompt := GenerateSystemPrompt(p, "Business model: SaaS")
	assert.Contains(t, prompt, "Bob")
	assert.Contains(t, prompt, "CTO")
	assert.Contains(t, prompt, "SaaS")
	assert.Contains(t, prompt, "ship it")
}
