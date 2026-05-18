package util

import (
	"fmt"
	"net/mail"
	"net/url"
	"strings"
)

// ValidateRequired validates that a value is provided and not empty.
func ValidateRequired(value, name string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

// ValidateURL validates that a string is a valid URL.
func ValidateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must include a scheme (http:// or https://)")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must include a host")
	}

	return nil
}

// ValidateEmail validates that a string is a valid email address.
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email address: %w", err)
	}

	return nil
}

// ValidatePositiveInt validates that an integer is positive (> 0).
func ValidatePositiveInt(value int64, name string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be a positive integer", name)
	}
	return nil
}

// ValidateNonNegativeInt validates that an integer is non-negative (>= 0).
func ValidateNonNegativeInt(value int, name string) error {
	if value < 0 {
		return fmt.Errorf("%s must be non-negative", name)
	}
	return nil
}

// ValidateOneOf validates that a value is one of the allowed values.
func ValidateOneOf(value string, name string, allowed []string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of: %s", name, strings.Join(allowed, ", "))
}
