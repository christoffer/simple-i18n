package internal

import (
	"testing"
)

func TestParseSingle(t *testing.T) {
	toml := `
		simple_greet = "Welcome"
		greet = "Hello, {name}!"

		[delete_modal]
		title = "Delete"
		confirm = "OK"
	`

	enData, err := parseContent(toml)
	if err != nil {
		t.Fatalf("parseContent failed: %v", err)
	}

	if len(enData.root) != 2 {
		t.Errorf("Expected 2 root entries, got %d", len(enData.root))
	}

	if enData.root["simple_greet"] != "Welcome" {
		t.Errorf("Expected 'Welcome', got '%s'", enData.root["simple_greet"])
	}

	if enData.root["greet"] != "Hello, {name}!" {
		t.Errorf("Expected 'Hello, {name}!', got '%s'", enData.root["greet"])
	}

	if len(enData.sections) != 1 {
		t.Errorf("Expected 1 section, got %d", len(enData.sections))
	}

	deleteModal, ok := enData.sections["delete_modal"]
	if !ok {
		t.Error("Expected 'delete_modal' section to be present")
	}

	if deleteModal["title"] != "Delete" {
		t.Errorf("Expected 'Delete', got '%s'", deleteModal["title"])
	}

	if deleteModal["confirm"] != "OK" {
		t.Errorf("Expected 'OK', got '%s'", deleteModal["confirm"])
	}
}

func TestParseTomlFiles_InvalidToml(t *testing.T) {
	_, err := parseContent("unicorns")
	if err == nil {
		t.Error("Expected error for invalid TOML content")
	}
}

func TestParseTomlFiles_UnsupportedType(t *testing.T) {
	_, err := parseContent(`foo = 12`)
	if err == nil {
		t.Error("Expected error for unsupported type in TOML content")
	}

	_, err = parseContent(`
		[foo.bar]
		single_level_only = "true"
	`)
	if err == nil {
		t.Error("Expected error for unsupported nested structure in TOML content")
	}
}

