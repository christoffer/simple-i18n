package main

import (
	"fmt"
	"github.com/christoffer/simple-i18n/cmd/test/generated"
)

func printAllTranslations(t *i18n.Translator, language string) {
	fmt.Printf("=== %s ===\n", language)
	fmt.Printf("Root message: %s\n", t.RootMessage())
	fmt.Printf("flavorPhrase: %s\n", t.FlavorPhrase())

	menu := t.Menu()

	fmt.Printf("Family (1, cat): %s\n", menu.Family(1, "cat"))
	fmt.Printf("Family (2, dog): %s\n", menu.Family(2, "dog"))

	fmt.Printf("Message (1, John): %s\n", menu.Message(1, "John"))
	fmt.Printf("Message (3, Mary): %s\n", menu.Message(3, "Mary"))

	specials := t.Specials()

	fmt.Printf("Count (42): %s\n", specials.Count(42))
	fmt.Printf("Score (33): %s\n", specials.Score("33"))

	fmt.Printf("Point (1): %s\n", specials.Point(1))
	fmt.Printf("Point (5): %s\n", specials.Point(5))

	fmt.Println()
}

func main() {
	// Create translator instance
	t := i18n.New()

	// Test all supported languages
	languages := []string{"en", "en_uk", "sv"}

	for _, lang := range languages {
		if err := t.SetLanguage(lang); err != nil {
			fmt.Printf("Error setting language %s: %v\n", lang, err)
			continue
		}
		printAllTranslations(t, lang)
	}

	// Demonstrate seamless switching in a function (new direct API)
	fmt.Println("=== Seamless Function Usage ===")
	getUserMessage := func(translator *i18n.Translator, username string, count int) string {
		return translator.Menu().Message(count, username)
	}

	if err := t.SetLanguage("en"); err != nil {
		fmt.Printf("Error setting language to English: %v\n", err)
		return
	}
	fmt.Printf("English function: %s\n", getUserMessage(t, "Alice", 5))

	if err := t.SetLanguage("sv"); err != nil {
		fmt.Printf("Error setting language to Swedish: %v\n", err)
		return
	}
	fmt.Printf("Swedish function: %s\n", getUserMessage(t, "Alice", 5))
}
