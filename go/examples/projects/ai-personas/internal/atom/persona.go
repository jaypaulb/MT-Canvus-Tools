package atom

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// AgeString is a string field that JSON-unmarshals from either a string or
// a number (Gemini sometimes returns "42", sometimes 42).
type AgeString string

// UnmarshalJSON implements json.Unmarshaler for AgeString.
func (a *AgeString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*a = AgeString(s)
		return nil
	}
	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		*a = AgeString(fmt.Sprintf("%d", int(n)))
		return nil
	}
	return fmt.Errorf("AgeString: cannot unmarshal %s", string(data))
}

// GoalsString is a string field that JSON-unmarshals from either a string
// or an array of strings (joined with newlines).
type GoalsString string

// UnmarshalJSON implements json.Unmarshaler for GoalsString.
func (g *GoalsString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*g = GoalsString(s)
		return nil
	}
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*g = GoalsString(strings.Join(arr, "\n"))
		return nil
	}
	return fmt.Errorf("GoalsString: cannot unmarshal %s", string(data))
}

// Persona represents a customer persona for focus group simulation.
type Persona struct {
	Name        string      `json:"name"`
	Role        string      `json:"role"`
	Description string      `json:"description"`
	Background  string      `json:"background"`
	Goals       GoalsString `json:"goals"`
	Age         AgeString   `json:"age"`
	Sex         string      `json:"sex"`
	Race        string      `json:"race"`
}

// personaNoteFormat is the human-readable layout used to render a persona
// into a Canvus note. Parsing relies on the same field markers.
const personaNoteFormat = "🧑 Name: %s\n\n💼 Role: %s\n\n📝 Description: %s\n\n🏫 Background: %s\n\n🎯 Goals: %s\n\n🎂 Age: %s\n\n⚧ Sex: %s\n\n🌍 Race: %s"

// FormatPersonaNote formats a persona for display in a Canvus note.
func FormatPersonaNote(p Persona) string {
	return fmt.Sprintf(personaNoteFormat,
		p.Name, p.Role, p.Description, p.Background,
		string(p.Goals), string(p.Age), p.Sex, p.Race)
}

// personaNoteRegex matches the markers produced by personaNoteFormat.
var personaNoteRegex = regexp.MustCompile(
	`(?m)^🧑 Name: (.*)[\s\S]*^💼 Role: (.*)[\s\S]*^📝 Description: (.*)[\s\S]*^🏫 Background: (.*)[\s\S]*^🎯 Goals: (.*)[\s\S]*^🎂 Age: (.*)[\s\S]*^⚧ Sex: (.*)[\s\S]*^🌍 Race: (.*)$`)

// ParsePersonaNote reverses FormatPersonaNote, returning a zero Persona on
// malformed input.
func ParsePersonaNote(text string) Persona {
	p := Persona{}
	matches := personaNoteRegex.FindStringSubmatch(text)
	if len(matches) == 9 {
		p.Name = matches[1]
		p.Role = matches[2]
		p.Description = matches[3]
		p.Background = matches[4]
		p.Goals = GoalsString(matches[5])
		p.Age = AgeString(matches[6])
		p.Sex = matches[7]
		p.Race = matches[8]
	}
	return p
}

// GenerateSystemPrompt returns the focus-group system prompt for a persona.
func GenerateSystemPrompt(persona Persona, businessContext string) string {
	return fmt.Sprintf(`Assume the role of the following persona for a business focus group. You are a client or potential client of the business. You are in a general purpose focus group for the business. Here is the business outline:

%s

Persona:
Name: %s
Role: %s
Description: %s
Background: %s
Goals: %s
Age: %s
Sex: %s
Race: %s

When asked a question or provided with some info, you must only respond as the persona assigned and in the voice of that persona. Your responses should be short and sweet and structured as if given verbally. You should not repeat the question or reiterate points from the question as this would not be natural for a conversational style interaction verbally. Do not start your answer by restating the question. Do not use phrases like 'As a persona...' or 'If I were...'. Just answer as if you are the person.`,
		businessContext,
		persona.Name,
		persona.Role,
		persona.Description,
		persona.Background,
		string(persona.Goals),
		string(persona.Age),
		persona.Sex,
		persona.Race,
	)
}
