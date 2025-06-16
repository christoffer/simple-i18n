package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

type TomlData struct {
	Locale string

	root     map[string]string
	sections map[string]map[string]string
}

func ProcessTomlDir(tomlDir string) (map[string]TomlData, error) {
	files, err := filepath.Glob(filepath.Join(tomlDir, "*.toml"))
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, err
	}

	langToTranslation := make(map[string]TomlData)

	// Need to be case-insensitive since we grab it from the filesystem
	localeRegexp, err := regexp.Compile(`^[a-z]{2}(_[a-z]{2})?$`)
	if err != nil {
		return nil, fmt.Errorf("failed to compile locale regex: %w", err)
	}
	for _, file := range files {
		locale := strings.ToLower(strings.TrimSuffix(filepath.Base(file), ".toml"))
		if !localeRegexp.MatchString(locale) {
			fmt.Fprintf(os.Stderr, "ignoring non-locale named file %s (got locale '%s', only accepting 'xx' or 'xx_xx')\n", file, locale)
			continue
		}
		fileData, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		data, err := parseContent(string(fileData))
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", file, err)
		}
		data.Locale = locale
		langToTranslation[locale] = data
	}

	return langToTranslation, nil
}

var prohibitedNames = map[string]bool{
	"SetLanguage": true,
}

func parseContent(tomlData string) (TomlData, error) {
	data := TomlData{
		root:     make(map[string]string),
		sections: make(map[string]map[string]string),
	}

	var tomlContent map[string]any
	if _, err := toml.Decode(tomlData, &tomlContent); err != nil {
		return TomlData{}, fmt.Errorf("failed to decode TOML content: %w", err)
	}

	for k := range tomlContent {
		generatedName := toPublicName(k)
		if prohibitedNames[generatedName] {
			return TomlData{}, fmt.Errorf("'%s' conflicts with '%s' and cannot be used as translation key", k, generatedName)
		}

		entry := tomlContent[k]
		if val, ok := entry.(string); ok {
			data.root[k] = val
			continue
		}

		if val, ok := entry.(map[string]any); ok {
			section := make(map[string]string)
			for sectionKey, sectionVal := range val {
				if strVal, ok := sectionVal.(string); ok {
					section[sectionKey] = strVal
				} else {
					return TomlData{}, fmt.Errorf("expected string under %s > %s, but found '%v'", k, sectionKey, sectionVal)
				}
			}
			data.sections[k] = section
			continue
		}

		return TomlData{}, fmt.Errorf("unexpected type for key %s: %T", k, entry)
	}
	return data, nil
}
