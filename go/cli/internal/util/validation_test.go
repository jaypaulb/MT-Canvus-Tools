package util

import (
	"testing"
)

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "non-empty value is valid",
			value:     "test",
			fieldName: "field",
			wantErr:   false,
		},
		{
			name:      "empty value is invalid",
			value:     "",
			fieldName: "field",
			wantErr:   true,
		},
		{
			name:      "whitespace-only value is invalid",
			value:     "   ",
			fieldName: "field",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequired(tt.value, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequired() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid http URL",
			url:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "valid https URL",
			url:     "https://example.com/path",
			wantErr: false,
		},
		{
			name:    "empty URL is invalid",
			url:     "",
			wantErr: true,
		},
		{
			name:    "URL without scheme is invalid",
			url:     "example.com",
			wantErr: true,
		},
		{
			name:    "invalid URL is invalid",
			url:     "ht!tp://invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with subdomain",
			email:   "user@mail.example.com",
			wantErr: false,
		},
		{
			name:    "empty email is invalid",
			email:   "",
			wantErr: true,
		},
		{
			name:    "email without @ is invalid",
			email:   "notanemail",
			wantErr: true,
		},
		{
			name:    "email without domain is invalid",
			email:   "test@",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePositiveInt(t *testing.T) {
	tests := []struct {
		name      string
		value     int64
		fieldName string
		wantErr   bool
	}{
		{
			name:      "positive value is valid",
			value:     10,
			fieldName: "field",
			wantErr:   false,
		},
		{
			name:      "zero is invalid",
			value:     0,
			fieldName: "field",
			wantErr:   true,
		},
		{
			name:      "negative value is invalid",
			value:     -5,
			fieldName: "field",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePositiveInt(tt.value, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePositiveInt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateOneOf(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		allowed   []string
		wantErr   bool
	}{
		{
			name:      "value in allowed list is valid",
			value:     "option1",
			fieldName: "field",
			allowed:   []string{"option1", "option2", "option3"},
			wantErr:   false,
		},
		{
			name:      "value not in allowed list is invalid",
			value:     "option4",
			fieldName: "field",
			allowed:   []string{"option1", "option2", "option3"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOneOf(tt.value, tt.fieldName, tt.allowed)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOneOf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
