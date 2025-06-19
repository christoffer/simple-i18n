package main

import (
	"fmt"
	"github.com/christoffer/simple-i18n/cmd/test/generated"
)

func printAllTranslations(t *i18n.Translator, language string) {
	fmt.Printf("=== %s ===\n", language)
	fmt.Printf("Root message: %s\n", t.RootMessage())

	menu := t.Menu()

	fmt.Printf("Family (1, cat): %s\n", menu.Family(1, "cat"))
	fmt.Printf("Family (2, dog): %s\n", menu.Family(2, "dog"))

	fmt.Printf("Message (1, John): %s\n", menu.Message(1, "John"))
	fmt.Printf("Message (3, Mary): %s\n", menu.Message(3, "Mary"))

	specials := t.Specials()

	fmt.Printf("Count (42): %s\n", specials.Count(42))
	fmt.Printf("Criteria (1): %s\n", specials.Criteria(1))
	fmt.Printf("Criteria (0): %s\n", specials.Criteria(0))

	fmt.Printf("Point (1): %s\n", specials.Point(1))
	fmt.Printf("Point (5): %s\n", specials.Point(5))

	fmt.Printf("Escaped: %s\n", specials.Ecaped())

	fmt.Println()
}

func main() {
	t := i18n.New()
	languages := []string{"en", "en_uk", "sv"}
	for _, lang := range languages {
		if err := t.SetLanguage(lang); err != nil {
			fmt.Printf("Error setting language %s: %v\n", lang, err)
			continue
		}
		printAllTranslations(t, lang)
	}
}
