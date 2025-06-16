package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
	flag.StringVar(&packageName, "p", "i18n", "Package name for generated files")

	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

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
		content, err := internal.GetTranslationImpl(tomlData)
		if err != nil {
			bail("Error generating translation implementation for %s: %v", locale, err)
		}
		writeFile(tomlData.Locale+".go", outputDir, content, verbose)
		allLocales = append(allLocales, locale)
		if baseData == nil {
			baseData = &tomlData
		}
	}

	if content, err := internal.WriteBaseTranslation(*baseData, outputDir); err != nil {
		bail("Error generating base translation interface: %v", err)
	} else {
		writeFile("base.go", outputDir, content, verbose)
	}

	if err := internal.WriteTranslator(allLocales, outputDir, *baseData); err != nil {
		bail("Error generating manager: %v", err)
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
