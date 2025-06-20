package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

type TomlParseResult struct {
	Locale string
	Errors []error

	root     map[string]TranslateFunc
	sections map[string]map[string]TranslateFunc
}

type ProcessedLocale struct {
	BaseLocale          string
	ParsedFuncsByLocale map[string]TomlParseResult
}

func ProcessTomlDir(tomlDir string, baseLocale string) (ProcessedLocale, error) {
	// Be case-insensitive since we're dealing with locales based on filenames
	localeRegexp, err := regexp.Compile(`^[a-z]{2}(_[a-z]{2})?$`)
	if err != nil {
		return ProcessedLocale{}, fmt.Errorf("failed to compile locale regex: %w", err)
	}

	if !localeRegexp.MatchString(baseLocale) {
		return ProcessedLocale{}, fmt.Errorf("invalid base locale: %s (expected format 'xx' or 'xx_xx')", baseLocale)
	}

	files, err := filepath.Glob(filepath.Join(tomlDir, "*.toml"))
	if err != nil {
		return ProcessedLocale{}, err
	}
	if len(files) == 0 {
		return ProcessedLocale{}, fmt.Errorf("no files found in %s", tomlDir)
	}

	parsedTomlByLocale := make(map[string]TomlParseResult)

	seenLocales := make(map[string]bool)
	errorsByFile := make(map[string][]error)
	for _, file := range files {
		filename := filepath.Base(file)
		locale := strings.ToLower(strings.TrimSuffix(filename, ".toml"))
		if !localeRegexp.MatchString(locale) {
			fmt.Fprintf(os.Stderr, "ignoring file %s (filename maps to locale '%s', only accepting forms 'xx' or 'xx_xx')\n", file, locale)
			continue
		}
		if seenLocales[locale] {
			fmt.Fprintf(os.Stderr, "ignoring duplicate locale %s from file %s\n", locale, file)
			continue
		}
		seenLocales[locale] = true
		if baseLocale == "" {
			baseLocale = locale
		}
		fileData, err := os.ReadFile(file)
		if err != nil {
			errorsByFile[filename] = append(errorsByFile[filename], fmt.Errorf("failed to read file %s: %w", file, err))
			continue
		}

		parsedToml := parseContent(locale, string(fileData))
		if len(parsedToml.Errors) > 0 {
			for _, err := range parsedToml.Errors {
				errorsByFile[filename] = append(errorsByFile[filename], err)
			}
			continue
		}
		parsedTomlByLocale[locale] = parsedToml
	}

	if len(errorsByFile) > 0 {
		// Bail early, it doesn't make sense to validate the file structures until they have the correct syntax
		var errorMsg strings.Builder
		for file, errors := range errorsByFile {
			errorMsg.WriteString(fmt.Sprintf("%s (%d errors)\n", file, len(errors)))
			for _, e := range errors {
				errorMsg.WriteString(fmt.Sprintf("\t- %s\n", e))
			}
		}
		errorMsg.WriteString("Aborting.\n")
		return ProcessedLocale{}, fmt.Errorf("%s", errorMsg.String())
	}

	if errors := validateAllLocales(baseLocale, parsedTomlByLocale); len(errors) != 0 {
		for locale, errors := range errors {
			var errorMsg strings.Builder
			errorMsg.WriteString(fmt.Sprintf("found %d validation errors", len(errors)))
			for _, err := range errors {
				errorMsg.WriteString("\n- ")
				errorMsg.WriteString(err.Error())
			}
			errorsByFile[locale] = append(errorsByFile[""], fmt.Errorf("%s", errorMsg.String()))
		}
	}

	if len(errorsByFile) > 0 {
		var sb strings.Builder
		for file, errors := range errorsByFile {
			sb.WriteString(fmt.Sprintf("\n%s", file))
			for _, err := range errors {
				sb.WriteString(fmt.Sprintf("\n - %s", err))
			}
		}
		return ProcessedLocale{}, fmt.Errorf("%s", sb.String())
	}

	return ProcessedLocale{
		BaseLocale:          baseLocale,
		ParsedFuncsByLocale: parsedTomlByLocale,
	}, nil
}

var prohibitedNames = map[string]bool{
	"SetLanguage":   true,
	"NewTranslator": true,
}

func parseContent(locale string, tomlData string) TomlParseResult {
	data := TomlParseResult{
		Locale:   locale,
		Errors:   make([]error, 0),
		root:     make(map[string]TranslateFunc),
		sections: make(map[string]map[string]TranslateFunc),
	}

	var tomlContent map[string]any
	if _, err := toml.Decode(tomlData, &tomlContent); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("failed to decode TOML content: %w", err))
		return data // fatal
	}

	for k := range tomlContent {
		generatedName := toPublicName(k)
		if prohibitedNames[generatedName] {
			data.Errors = append(data.Errors, fmt.Errorf("'%s' conflicts with '%s' and cannot be used as translation key", k, generatedName))
			continue
		}

		// Root entries
		entry := tomlContent[k]
		if val, ok := entry.(string); ok {
			trFunc, err := parseTranslateFunc(k, val)
			if err != nil {
				data.Errors = append(data.Errors, err)
			} else {
				data.root[k] = trFunc
			}
			continue
		}

		// Sections
		if section, ok := entry.(map[string]any); ok {
			sectionFuncs := make(map[string]TranslateFunc)
			for sectionKey, sectionVal := range section {
				if strVal, ok := sectionVal.(string); ok {
					trFunc, err := parseTranslateFunc(sectionKey, strVal)
					if err != nil {
						data.Errors = append(data.Errors, err)
					} else {
						sectionFuncs[sectionKey] = trFunc
					}
				} else {
					data.Errors = append(data.Errors, fmt.Errorf("expected string under %s > %s, but found '%v'", k, sectionKey, sectionVal))
				}
			}
			data.sections[k] = sectionFuncs
			continue
		}

		data.Errors = append(data.Errors, fmt.Errorf("unexpected type for key %s: %T", k, entry))
	}

	return data
}

func validateSection(baseMap, otherMap map[string]TranslateFunc, sectionName string, otherLocale string) []error {
	errors := make([]error, 0)
	keyName := func(key string) string {
		if sectionName == "" {
			return key
		}
		return fmt.Sprintf("[%s]: %s", sectionName, key)
	}

	for key, baseFunc := range baseMap {
		otherFunc, exists := otherMap[key]
		if !exists {
			errors = append(errors, fmt.Errorf("%s is missing translation '%s'", otherLocale, keyName(key)))
		}

		baseSig := baseFunc.Signature()
		otherSig := otherFunc.Signature()

		if baseSig != otherSig {
			errors = append(errors, fmt.Errorf("%s has the wrong signature for '%s'. Should be `%s`, but was `%s`", otherLocale, keyName(key), baseSig, otherSig))
		}
	}

	for key := range otherMap {
		if _, exists := baseMap[key]; !exists {
			errors = append(errors, fmt.Errorf("%s has an unknown translation '%s'", otherLocale, keyName(key)))
		}
	}
	return errors
}

func validateAllLocales(baseLocale string, localeToData map[string]TomlParseResult) map[string][]error {
	errors := make(map[string][]error)
	baseLocaleData, ok := localeToData[baseLocale]
	if !ok {
		errors[baseLocale] = append(errors[baseLocale], fmt.Errorf("base locale '%s' not found in provided locales", baseLocale))
		return errors // critical error
	}

	for otherLocale, otherLocaleData := range localeToData {
		if otherLocale == baseLocale {
			continue
		}

		if sectionErrors := validateSection(baseLocaleData.root, otherLocaleData.root, "", otherLocale); len(sectionErrors) != 0 {
			for _, err := range sectionErrors {
				errors[otherLocale] = append(errors[otherLocale], err)
			}
		}

		for sectionName, baseSection := range baseLocaleData.sections {
			otherSection, exists := otherLocaleData.sections[sectionName]
			if !exists {
				errors[otherLocale] = append(errors[otherLocale], fmt.Errorf("%s is missing section [%s]", otherLocale, sectionName))
				continue
			}

			if sectionErrors := validateSection(baseSection, otherSection, sectionName, otherLocale); len(sectionErrors) != 0 {
				for _, err := range sectionErrors {
					errMsg := fmt.Errorf("%s: %s", otherLocale, err.Error())
					errors[otherLocale] = append(errors[otherLocale], errMsg)
				}
			}
		}

		for sectionName := range otherLocaleData.sections {
			if _, exists := baseLocaleData.sections[sectionName]; !exists {
				errors[otherLocale] = append(errors[otherLocale], fmt.Errorf("%s has unknown section [%s]", otherLocale, sectionName))
			}
		}
	}

	return errors
}
