package jsonmask

import (
	"testing"
)

func TestUpper(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, `"HELLO"`},
		{`"Hello"`, `"HELLO"`},
		{`""`, `""`},
	}

	for _, tt := range tests {
		result := string(Upper(tt.input))
		if result != tt.expected {
			t.Errorf("Upper(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestLower(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"HELLO"`, `"hello"`},
		{`"Hello"`, `"hello"`},
		{`""`, `""`},
	}

	for _, tt := range tests {
		result := string(Lower(tt.input))
		if result != tt.expected {
			t.Errorf("Lower(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestInitialChar(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, `"H"`},
		{`"H"`, `"H"`},
		{`""`, `""`},
	}

	for _, tt := range tests {
		result := string(InitialChar(tt.input))
		if result != tt.expected {
			t.Errorf("InitialChar(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestPrefixFn(t *testing.T) {
	tests := []struct {
		length      int
		addEllipsis bool
		input       string
		expected    string
	}{
		{3, true, `"hello"`, `"hel..."`},
		{2, false, `"hello"`, `"he"`},
		{5, true, `"hello"`, `"hello"`},
	}

	for _, tt := range tests {
		fn := PrefixFn(tt.length, tt.addEllipsis)
		result := string(fn(tt.input))
		if result != tt.expected {
			t.Errorf("PrefixFn(%d, %t)(%s) = %s; want %s", tt.length, tt.addEllipsis, tt.input, result, tt.expected)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, `""`},
		{`null`, `null`},
		{`NULL`, `NULL`},
		{`""`, `""`},
	}

	for _, tt := range tests {
		result := string(Truncate(tt.input))
		if result != tt.expected {
			t.Errorf("Truncate(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNull(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, `null`},
		{`""`, `null`},
	}

	for _, tt := range tests {
		result := string(Null(tt.input))
		if result != tt.expected {
			t.Errorf("Null(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"user@example.com"`, `"u**r@e******.com"`},
		{`"ab@domain.com"`, `"a*@d*****.com"`},
		{`"invalid"`, `"invalid_email_format"`},
		{`"@missinglocal.com"`, `"invalid_email_format"`},
		{`missingquotes@example.com`, `"invalid_email_format"`},
	}

	for _, tt := range tests {
		result := string(Email(tt.input))
		if result != tt.expected {
			t.Errorf("Email(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestZero(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"123", "0"},
		{"", "0"},
		{"0", "0"},
	}

	for _, tt := range tests {
		result := string(Zero(tt.input))
		if result != tt.expected {
			t.Errorf("Zero(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}
