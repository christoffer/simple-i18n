package internal

import (
	"fmt"
	"strings"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "plain text",
			input: "Hello world",
			expected: []Token{
				{Type: TokenText, Value: "Hello world", Start: 0, End: 11},
			},
		},
		{
			name:  "single substitution",
			input: "Hi {name}",
			expected: []Token{
				{Type: TokenText, Value: "Hi ", Start: 0, End: 3},
				{Type: TokenSub, Value: "name", Start: 3, End: 9},
			},
		},
		{
			name:  "multiple substitutions",
			input: "Hello {name}, you have {count} messages",
			expected: []Token{
				{Type: TokenText, Value: "Hello ", Start: 0, End: 6},
				{Type: TokenSub, Value: "name", Start: 6, End: 12},
				{Type: TokenText, Value: ", you have ", Start: 12, End: 23},
				{Type: TokenSub, Value: "count", Start: 23, End: 30},
				{Type: TokenText, Value: " messages", Start: 30, End: 39},
			},
		},
		{
			name:  "simple plural",
			input: "You have {count} item{{s}}",
			expected: []Token{
				{Type: TokenText, Value: "You have ", Start: 0, End: 9},
				{Type: TokenSub, Value: "count", Start: 9, End: 16},
				{Type: TokenText, Value: " item", Start: 16, End: 21},
				{Type: TokenPlural, Value: "s", Start: 21, End: 26},
			},
		},
		{
			name:  "substitution and plural",
			input: "Hello {name}, you have {count} item{{s}}",
			expected: []Token{
				{Type: TokenText, Value: "Hello ", Start: 0, End: 6},
				{Type: TokenSub, Value: "name", Start: 6, End: 12},
				{Type: TokenText, Value: ", you have ", Start: 12, End: 23},
				{Type: TokenSub, Value: "count", Start: 23, End: 30},
				{Type: TokenText, Value: " item", Start: 30, End: 35},
				{Type: TokenPlural, Value: "s", Start: 35, End: 40},
			},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []Token{},
		},
		{
			name:  "only substitution",
			input: "{name}",
			expected: []Token{
				{Type: TokenSub, Value: "name", Start: 0, End: 6},
			},
		},
		{
			name:  "only plural",
			input: "{{s}}",
			expected: []Token{
				{Type: TokenPlural, Value: "s", Start: 0, End: 5},
			},
		},
		{
			name:  "consecutive substitutions",
			input: "{first}{second}",
			expected: []Token{
				{Type: TokenSub, Value: "first", Start: 0, End: 7},
				{Type: TokenSub, Value: "second", Start: 7, End: 15},
			},
		},
		{
			name:  "consecutive plurals",
			input: "{{first}}{{second}}",
			expected: []Token{
				{Type: TokenPlural, Value: "first", Start: 0, End: 9},
				{Type: TokenPlural, Value: "second", Start: 9, End: 19},
			},
		},
		{
			name:  "empty substitution",
			input: "Hi {}",
			expected: []Token{
				{Type: TokenText, Value: "Hi ", Start: 0, End: 3},
				{Type: TokenSub, Value: "", Start: 3, End: 5},
			},
		},
		{
			name:  "empty plural",
			input: "Item{{}}",
			expected: []Token{
				{Type: TokenText, Value: "Item", Start: 0, End: 4},
				{Type: TokenPlural, Value: "", Start: 4, End: 8},
			},
		},
		{
			name:  "bad case: nested substitutions",
			input: "{name{nested}}",
			expected: []Token{
				{Type: TokenSub, Value: "name{nested", Start: 0, End: 13},
				{Type: TokenText, Value: "}", Start: 13, End: 14},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tokenize(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d tokens, got %d\n\n%s", len(tt.expected), len(result), printAllTokens(result))
				return
			}

			for i, token := range result {
				expected := tt.expected[i]
				if token.Type != expected.Type {
					e := presentTokenType(expected.Type)
					a := presentTokenType(token.Type)
					t.Errorf("Token %d: expected type %s, got %s", i, e, a)
				}
				if token.Value != expected.Value {
					t.Errorf("Token %d: expected value %q, got %q", i, expected.Value, token.Value)
				}
				if token.Start != expected.Start {
					t.Errorf("Token %d: expected start %d, got %d", i, expected.Start, token.Start)
				}
				if token.End != expected.End {
					t.Errorf("Token %d: expected end %d, got %d", i, expected.End, token.End)
				}
			}
		})
	}
}

func printAllTokens(tokens []Token) string {
	var sb strings.Builder
	for i, token := range tokens {
		tokenType := presentTokenType(token.Type)
		t := fmt.Sprintf("%s, Value: %q, Start: %d, End: %d", tokenType, token.Value, token.Start, token.End)
		sb.WriteString(fmt.Sprintf("Token %d: %s\n", i, t))
	}
	return sb.String()
}

func presentTokenType(tokenType TokenType) string {
	switch tokenType {
	case TokenText:
		return "Text"
	case TokenSub:
		return "Substitution"
	case TokenPlural:
		return "Plural"
	default:
		return "Unknown"
	}
}
