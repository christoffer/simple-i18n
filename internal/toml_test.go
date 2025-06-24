package internal

import (
	"strings"
	"testing"
)

func TestParseContent_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		toml          string
		expectError   bool
		errorContains string
	}{
		{
			name:          "invalid TOML syntax",
			toml:          "invalid toml [[[",
			expectError:   true,
			errorContains: "failed to decode TOML",
		},
		{
			name:          "unsupported non-string type",
			toml:          "age = 123",
			expectError:   true,
			errorContains: "unexpected type",
		},
		{
			name:          "nested sections not supported",
			toml:          "[foo.bar]\nkey = \"value\"",
			expectError:   true,
			errorContains: "expected string under foo > bar",
		},
		{
			name:          "prohibited name SetLanguage",
			toml:          "set_language = \"test\"",
			expectError:   true,
			errorContains: "conflicts with 'SetLanguage'",
		},
		{
			name:          "prohibited name NewTranslator",
			toml:          "new_translator = \"test\"",
			expectError:   true,
			errorContains: "conflicts with 'NewTranslator'",
		},
		{
			name:          "non-string in section",
			toml:          "[section]\nkey = 123",
			expectError:   true,
			errorContains: "expected string under section > key",
		},
		{
			name:          "malformed substitution syntax",
			toml:          "greeting = \"Hello {name\"",
			expectError:   true,
			errorContains: "syntax error",
		},
		{
			name:          "malformed plural syntax",
			toml:          "items = \"{{count item\"",
			expectError:   true,
			errorContains: "syntax error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseContent("en", tt.toml)

			if tt.expectError {
				if len(result.Errors) == 0 {
					t.Errorf("Expected error for %s, but got none", tt.name)
				} else {
					found := false
					for _, err := range result.Errors {
						if strings.Contains(err.Error(), tt.errorContains) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error containing '%s', but got: %v", tt.errorContains, result.Errors)
					}
				}
			} else {
				if len(result.Errors) > 0 {
					t.Errorf("Expected no error for %s, but got: %v", tt.name, result.Errors)
				}
			}
		})
	}
}

func TestValidateSection_MismatchedSignatures(t *testing.T) {
	// Test that validation catches signature mismatches between locales
	base := map[string]TranslateFunc{
		"greet": {
			Name:   "Greet",
			Params: []TranslateFuncParam{{Name: "name", Type: "string"}},
		},
	}

	other := map[string]TranslateFunc{
		"greet": {
			Name: "Greet",
			Params: []TranslateFuncParam{
				{Name: "name", Type: "string"},
				{Name: "count", Type: "int"},
			},
		},
	}

	errors := validateSection(base, other, "", "fr")
	if len(errors) == 0 {
		t.Error("Expected validation error for mismatched signatures")
	}

	if !strings.Contains(errors[0].Error(), "wrong signature") {
		t.Errorf("Expected signature mismatch error, got: %v", errors[0])
	}
}

func TestValidateSection_MissingTranslations(t *testing.T) {
	base := map[string]TranslateFunc{
		"hello":   {Name: "Hello"},
		"goodbye": {Name: "Goodbye"},
	}

	other := map[string]TranslateFunc{
		"hello": {Name: "Hello"},
		// Missing "goodbye"
	}

	errors := validateSection(base, other, "", "fr")
	if len(errors) == 0 {
		t.Error("Expected validation error for missing translation")
	}

	if !strings.Contains(errors[0].Error(), "missing translation") {
		t.Errorf("Expected missing translation error, got: %v", errors[0])
	}
}

func TestValidateSection_ExtraTranslations(t *testing.T) {
	base := map[string]TranslateFunc{
		"hello": {Name: "Hello"},
	}

	other := map[string]TranslateFunc{
		"hello": {Name: "Hello"},
		"extra": {Name: "Extra"}, // Not in base
	}

	errors := validateSection(base, other, "", "fr")
	if len(errors) == 0 {
		t.Error("Expected validation error for extra translation")
	}

	if !strings.Contains(errors[0].Error(), "unknown translation") {
		t.Errorf("Expected unknown translation error, got: %v", errors[0])
	}
}
