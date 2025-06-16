package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/christoffer/simple-i18n/internal"
)

func main() {
	var tomlDir string
	flag.StringVar(&tomlDir, "i", "translations", "Directory containing TOML files")

	var outputDir string
	flag.StringVar(&outputDir, "o", "i18n", "Output directory for generated files")

	var verbose bool
	flag.BoolVar(&verbose, "v", false, "Enable verbose output")

	var packageName string
	flag.StringVar(&packageName, "p", "", "Package name for generated files (defaults to output directory name)")

	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	if packageName == "" {
		packageName = filepath.Base(outputDir)
	}

	validatePackageName(packageName)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		bail("Error creating output directory: %v", err)
	}

	tomlDataByLocale, err := internal.ProcessTomlDir(tomlDir)
	if err != nil {
		bail("Error parsing TOML files: %v", err)
	}

	if len(tomlDataByLocale) == 0 {
		bail("No TOML files found in %s", tomlDir)
	}

	var baseData *internal.TomlData
	allLocales := make([]string, 0)
	for locale, tomlData := range tomlDataByLocale {
		content, err := internal.GetTranslationImpl(tomlData, packageName)
		if err != nil {
			bail("Error generating translation implementation for %s: %v", locale, err)
		}
		writeFile(tomlData.Locale+".go", outputDir, content, verbose)
		allLocales = append(allLocales, locale)
		if baseData == nil {
			baseData = &tomlData
		}
	}

	//goland:noinspection ALL Can't be nil: We check that allLocales is non-empty, and we pick the first one
	baseLocaleData := *baseData
	if content, err := internal.GetBaseTranslation(baseLocaleData, packageName); err != nil {
		bail("Error generating base translation interface: %v", err)
	} else {
		writeFile("base.go", outputDir, content, verbose)
	}

	if content, err := internal.GetTranslator(allLocales, packageName, baseLocaleData); err != nil {
		bail("Error generating translator: %v", err)
	} else {
		writeFile("translator.go", outputDir, content, verbose)
	}

	fmt.Printf("Generated translation files for locales: %s\n", strings.Join(allLocales, ", "))
}

func writeFile(filename string, outputDir string, content []byte, verbose bool) {
	outfile := filepath.Join(outputDir, filename)
	if err := os.WriteFile(outfile, content, 0644); err != nil {
		bail("error writing file %s: %v\n", outfile, err)
	}
	if verbose {
		fmt.Printf("Wrote to %d bytes %s\n", len(content), outfile)
	}
}

func bail(msg string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func validatePackageName(packageName string) {
	if packageName == "" {
		bail("Package name cannot be empty")
	}
	// Regular expression to validate package names
	var validPackageNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if !validPackageNameRegex.MatchString(packageName) {
		bail("Invalid package name: %s", packageName)
	}
}
