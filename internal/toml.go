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

type ProcessResult struct {
	BaseLocale       string
	TomlDataByLocale map[string]TomlData
}

func ProcessTomlDir(tomlDir string, baseLocale string) (ProcessResult, error) {
	// Be case-insensitive since we're dealing with locales based on filenames
	localeRegexp, err := regexp.Compile(`^[a-z]{2}(_[a-z]{2})?$`)
	if err != nil {
		return ProcessResult{}, fmt.Errorf("failed to compile locale regex: %w", err)
	}

	if !localeRegexp.MatchString(baseLocale) {
		return ProcessResult{}, fmt.Errorf("invalid base locale: %s (expected format 'xx' or 'xx_xx')", baseLocale)
	}

	files, err := filepath.Glob(filepath.Join(tomlDir, "*.toml"))
	if err != nil {
		return ProcessResult{}, err
	}

	if len(files) == 0 {
		return ProcessResult{}, err
	}

	dataByLocale := make(map[string]TomlData)

	seenLocales := make(map[string]bool)
	for _, file := range files {
		locale := strings.ToLower(strings.TrimSuffix(filepath.Base(file), ".toml"))
		if !localeRegexp.MatchString(locale) {
			fmt.Fprintf(os.Stderr, "ignoring file %s (filename maps to locale '%s', only accepting forms 'xx' or 'xx_xx')\n", file, locale)
			continue
		}
		if seenLocales[locale] {
			fmt.Fprintf(os.Stderr, "ignoring duplicate locale %s in file %s\n", locale, file)
			continue
		}
		seenLocales[locale] = true
		if baseLocale == "" {
			baseLocale = locale
		}
		fileData, err := os.ReadFile(file)
		if err != nil {
			return ProcessResult{}, err
		}

		data, err := parseContent(string(fileData))
		if err != nil {
			return ProcessResult{}, fmt.Errorf("failed to parse %s: %w", file, err)
		}
		data.Locale = locale
		dataByLocale[locale] = data
	}

	if errors := validateAllLocales(baseLocale, dataByLocale); len(errors) != 0 {
		var errorMsg strings.Builder
		errorMsg.WriteString(fmt.Sprintf("found %d validation errors", len(errors)))
		for _, err := range errors {
			errorMsg.WriteString("\n- ")
			errorMsg.WriteString(err.Error())
		}
		return ProcessResult{}, fmt.Errorf("%s", errorMsg.String())
	}

	return ProcessResult{
		BaseLocale:       baseLocale,
		TomlDataByLocale: dataByLocale,
	}, nil
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

func validateSection(base, other map[string]string, sectionName string, otherLocale string) []error {
	errors := make([]error, 0)
	keyName := func(key string) string {
		if sectionName == "" {
			return key
		}
		return fmt.Sprintf("[%s]: %s", sectionName, key)
	}

	for key, baseValue := range base {
		otherValue, exists := other[key]
		if !exists {
			errors = append(errors, fmt.Errorf("%s is missing translation '%s'", otherLocale, keyName(key)))
		}
		sigBase := GetFuncSignatureId(key, baseValue)
		sigOther := GetFuncSignatureId(key, otherValue)
		if sigBase != sigOther {
			errors = append(errors, fmt.Errorf("%s has the wrong signature for '%s'. Should be `%s`, but was `%s`", otherLocale, keyName(key), sigBase, sigOther))
		}
	}

	for key := range other {
		if _, exists := base[key]; !exists {
			errors = append(errors, fmt.Errorf("%s has an unknown translation '%s'", otherLocale, keyName(key)))
		}
	}
	return errors
}

func validateSections(base, other map[string]map[string]string, otherLocale string) []error {
	errors := make([]error, 0)

	for sectionName, baseSection := range base {
		otherSection, exists := other[sectionName]
		if !exists {
			errors = append(errors, fmt.Errorf("%s is missing section [%s]", otherLocale, sectionName))
			continue
		}

		if sectionErrors := validateSection(baseSection, otherSection, sectionName, otherLocale); len(sectionErrors) != 0 {
			for _, err := range sectionErrors {
				errMsg := fmt.Errorf("%s: %s", otherLocale, err.Error())
				errors = append(errors, errMsg)
			}
		}
	}

	for sectionName := range other {
		if _, exists := base[sectionName]; !exists {
			errors = append(errors, fmt.Errorf("%s has unknown section [%s]", otherLocale, sectionName))
		}
	}

	return errors
}

func validateAllLocales(baseLocale string, localeToData map[string]TomlData) []error {
	errors := make([]error, 0)
	baseLocaleData, ok := localeToData[baseLocale]
	if !ok {
		errors = append(errors, fmt.Errorf("base locale '%s' not found in provided locales", baseLocale))
		return errors // critical error
	}

	for otherLocale, otherLocaleData := range localeToData {
		if otherLocale == baseLocale {
			continue
		}

		if sectionErrors := validateSection(baseLocaleData.root, otherLocaleData.root, "", otherLocale); len(sectionErrors) != 0 {
			for _, err := range sectionErrors {
				errors = append(errors, err)
			}
		}

		if sectionErrors := validateSections(baseLocaleData.sections, otherLocaleData.sections, otherLocale); len(sectionErrors) != 0 {
			for _, err := range sectionErrors {
				errors = append(errors, err)
			}
		}
	}

	return errors
}
