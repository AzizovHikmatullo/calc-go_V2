package calc

import (
	"errors"
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
		err      error
	}{
		{
			name:     "Simple expression",
			input:    "1+2*3",
			expected: []string{"1", "+", "2", "*", "3"},
		},
		{
			name:     "Expression with spaces",
			input:    " 1 + 2 * 3 ",
			expected: []string{"1", "+", "2", "*", "3"},
		},
		{
			name:     "Expression with parentheses",
			input:    "(1+2)*3",
			expected: []string{"(", "1", "+", "2", ")", "*", "3"},
		},
		{
			name:     "Decimal numbers",
			input:    "1.5+2.75*3.2",
			expected: []string{"1.5", "+", "2.75", "*", "3.2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Tokenize(tt.input)
			if !errorsAreEqual(err, tt.err) {
				t.Errorf("expected error %v, got %v", tt.err, err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestInfixToPostfix(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
		err      error
	}{
		{
			name:     "Simple expression",
			input:    []string{"1", "+", "2", "*", "3"},
			expected: []string{"1", "2", "3", "*", "+"},
		},
		{
			name:     "Parentheses",
			input:    []string{"(", "1", "+", "2", ")", "*", "3"},
			expected: []string{"1", "2", "+", "3", "*"},
		},
		{
			name:     "Complex expression",
			input:    []string{"3", "+", "4", "*", "2", "/", "(", "1", "-", "5", ")", "*", "2", "/", "3"},
			expected: []string{"3", "4", "2", "*", "1", "5", "-", "/", "2", "*", "3", "/", "+"},
		},
		{
			name:     "Mismatched parentheses (missing closing)",
			input:    []string{"(", "1", "+", "2", "*", "3"},
			expected: nil,
			err:      errors.New("mismatched parentheses"),
		},
		{
			name:     "Mismatched parentheses (missing opening)",
			input:    []string{"1", "+", "2", ")", "*", "3"},
			expected: nil,
			err:      errors.New("mismatched parentheses"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := InfixToPostfix(tt.input)
			if !errorsAreEqual(err, tt.err) {
				t.Errorf("expected error %v, got %v", tt.err, err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid integer",
			input:    "123",
			expected: true,
		},
		{
			name:     "Valid float",
			input:    "123.456",
			expected: true,
		},
		{
			name:     "Negative number",
			input:    "-123",
			expected: true,
		},
		{
			name:     "Invalid number",
			input:    "abc",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNumber(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Addition operator",
			input:    "+",
			expected: true,
		},
		{
			name:     "Subtraction operator",
			input:    "-",
			expected: true,
		},
		{
			name:     "Multiplication operator",
			input:    "*",
			expected: true,
		},
		{
			name:     "Division operator",
			input:    "/",
			expected: true,
		},
		{
			name:     "Invalid operator",
			input:    "%",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsOperator(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func errorsAreEqual(err1, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	}
	if err1 != nil && err2 != nil {
		return err1.Error() == err2.Error()
	}
	return false
}
