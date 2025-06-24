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
	flag.StringVar(&tomlDir, "i", "translations", "Input dir containing TOML files")

	var outputDir string
	flag.StringVar(&outputDir, "o", "i18n", "Output directory for generated files")

	var verbose bool
	flag.BoolVar(&verbose, "v", false, "Enable verbose output")

	var packageName string
	flag.StringVar(&packageName, "p", "", "Package name for generated files (defaults to output directory name)")

	var baseLocale string
	flag.StringVar(&baseLocale, "b", "", "Base locale for translations (defaults to the first locale found in input dir)")

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

	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		bail("Error creating output directory: %s", err)
	}

	processResult, err := internal.ProcessTomlDir(tomlDir, baseLocale)
	if err != nil {
		bail("Generation prevented:\n%s", err)
	}

	if len(processResult.ParsedFuncsByLocale) == 0 {
		bail("No TOML files found in %s", tomlDir)
	}

	allLocales := make([]string, 0)
	for locale, tomlData := range processResult.ParsedFuncsByLocale {
		content, err := internal.GetTranslationImpl(tomlData, packageName, verbose)
		if err != nil {
			bail("Error generating translation implementation for %s: %v", locale, err)
		}
		writeFile(tomlData.Locale+".go", outputDir, content, verbose)
		allLocales = append(allLocales, locale)
	}

	baseLocaleData := processResult.ParsedFuncsByLocale[processResult.BaseLocale]
	if content, err := internal.GetBaseTranslation(baseLocaleData, packageName, verbose); err != nil {
		bail("Error generating base translation interface: %v", err)
	} else {
		writeFile("base.go", outputDir, content, verbose)
	}

	if content, err := internal.GetTranslator(allLocales, baseLocaleData, packageName, verbose); err != nil {
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
