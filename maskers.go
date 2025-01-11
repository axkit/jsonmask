// This file contains default maskers that are used to mask the JSON data.
// All functions have the same signature: func(string) []byte.
// The input string is a JSON string that represents a single value (with quotes).
package jsonmask

import (
	"bytes"
	"strings"
)

// Upper returns the input string in uppercase.
func Upper(s string) []byte {
	return bytes.ToUpper([]byte(s))
}

// Lower returns the input string in lowercase.
func Lower(s string) []byte {
	return bytes.ToLower([]byte(s))
}

// InitialChar returns the first character of the input string in uppercase.
func InitialChar(s string) []byte {
	if len(s) > 2 {
		return bytes.ToUpper([]byte(s[:2] + `"`))
	}
	return []byte(s)
}

// PrefixFn returns a function that prefixes the input string with the specified length.
func PrefixFn(length int, addEllipsis bool) func(string) []byte {
	return func(s string) []byte {
		if len(s) <= length+2 { // Include the opening and closing quotes
			return []byte(s)
		}

		res := s[:length+1] // Include the opening quote
		if addEllipsis {
			res += `..."`
		} else {
			res += `"`
		}
		return []byte(res)
	}
}

// Truncate masks the input string to an empty string if it is not NULL.
func Truncate(s string) []byte {
	if len(s) > 2 && strings.ToUpper(s) != `NULL` {
		return []byte(`""`)
	}
	return []byte(s)
}

// Null masks the input string to NULL without quotes.
func Null(s string) []byte {
	return []byte(`null`)
}

// Email masks the input string holding email address.
func Email(email string) []byte {
	var invalidEmail = []byte(`"invalid_email_format"`)

	// Check for the presence of quotes
	if len(email) < 2 || email[0] != '"' || email[len(email)-1] != '"' {
		return invalidEmail
	}

	// Process the email while keeping quotes in place
	emailBytes := []byte(email)

	// Find the position of the '@' symbol
	atIndex := -1
	for i := 1; i < len(emailBytes)-1; i++ { // Start at 1 to skip the opening quote
		if emailBytes[i] == '@' {
			atIndex = i
			break
		}
	}
	if atIndex <= 1 || atIndex == len(emailBytes)-1 {
		return invalidEmail
	}

	// Mask the local part
	if atIndex-1 <= 2 {
		for i := 2; i < atIndex; i++ { // Skip the opening quote
			emailBytes[i] = '*'
		}
	} else {
		for i := 2; i < atIndex-1; i++ {
			emailBytes[i] = '*'
		}
	}

	// Mask the domain part up to the last dot
	lastDotIndex := -1
	for i := len(emailBytes) - 2; i > atIndex; i-- { // Start from the end, skipping the closing quote
		if emailBytes[i] == '.' {
			lastDotIndex = i
			break
		}
	}
	if lastDotIndex == -1 {
		return invalidEmail
	}

	// Mask the domain part, leaving the last two levels visible
	for i := atIndex + 2; i < lastDotIndex; i++ {
		emailBytes[i] = '*'
	}

	return emailBytes
}

// Zero masks the input string holding numeric value to 0 without quotes.
func Zero(s string) []byte {
	return []byte(`0`)
}
