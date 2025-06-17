# simple-i18n

A tool for simplifying translations in Go applications. Generates type-safe translation functions from TOML files.

## Example

```toml
# en.toml
greeting = "Hello, {name}!"

[sidebar]
notifications = "You have {count} new notification{{s}} in {inbox}"
```

```toml
# sv.toml
greeting = "Hej {name}!"

[sidebar]
notifcations = "Du har {count} meddalande{{n}} i {inbox}"
```

```go
import (
	"fmt"		
	"my-project/i18n"
)

func main() {
	name := "Christoffer"
	numMessages := 42
	
	t := i18n.New()
	
	fmt.Println(t.Greeting(name)) // "Hello, Christoffer!"
	fmt.Println(t.SideBar().Notifications(1, "emails")) // "You have 1 new notification in emails"
	fmt.Println(t.SideBar().Notifications(numMessages, "direct messages")) // "You have 42 new notifications in direct messages"
	
	t.SetLocale("sv")
	fmt.Println(t.Greeting(name)) // "Hej, Christoffer!"
}
```
## Features

- **Type-safe translations**: Generates Go interfaces and implementations for each locale
- **Simple setup**: Just drop TOML with locale names (e.g. `en.toml`) into `translations/` to process them. All files are validated against a base locale, no format configuration required.
- **Pluralization**: Handles plural forms with `{{s}}` notation
- **Parameter interpolation**: Supports `{param}` placeholders in translation strings  
- **Runtime safety**: All translation files are validated at generation time. No missing keys or invalid formats during runtime. 

## Installation

```bash
go build -o bin/simple-i18n ./cmd/cli/main.go
```

## Usage

```bash
./bin/simple-i18n [options]
```

### Options

- `-i <dir>`: Input directory containing TOML files (default: "translations")
- `-o <dir>`: Output directory for generated files (default: "i18n") 
- `-p <name>`: Package name for generated files (default: output directory)
- `-b <locale>`: Base locale for translations (default: first locale found)
- `-v`: Enable verbose output

### Example

```bash
./bin/simple-i18n -i ./translations -o ./i18n -p i18n -b en
```

## Translation Files

Translation files are TOML files that specify a key and a translation template. The key determines the function name, and the template determines the function signature and the output.

### Structure

You can have messages in the root, or in a subsection (one level only).

#### Input

```toml
# en.toml
message = "Hi"

[my_section]
message = "Hello from section"
```

#### Usage

```go
package main

import (
	"fmt"
	"i18n"
)

func main() {
	t := i18n.New()
	fmt.Printf("%s", t.Message())
	fmt.Printf("%s", t.MySection().Message())
}

```

#### Generated

```go
package i18n

type TranslationEn_MySection struct {}

func (t *TranslationEn_MySection) Message() string {
    return "Hello from section"
}

type TranslationEn struct {
	mySection     TranslationEn_MySection
}

func (t *TranslationEn) Message() string {
    return "Hi"
}
```

### Parameters

Parameters are specified using `{param}` syntax. The param name must be a valid Go identifier. All generated arguments are of type `string`, except for `count` which is of type `int`.

**en.toml**:
```toml
root_message = "Welcome"
flavorPhrase = "Organize Colors of Canceled Cookies"

[menu]
message = "The {name} has {count} notification{{s}}"
family = "The {thing} was happy with a family of {thing}{{s}}"

[specials]
count = "Count: {count}"
score = "Score: {score}"
point = "Point{{s}}"
```

**sv.toml**:
```toml
root_message = "Välkommen"  
flavorPhrase = "Organisera färger av avbrutna kakor"

[menu]
message = "{name} har {count} notis{{er}}"
family = "{thing} var glad med en familj av {thing}{{ar}}"

[specials]
count = "Antal: {count}"
score = "Poäng: {score}"
point = "Poäng{{}}"
```

## Generated Code

The tool generates:

- `base.go`: Interface defining all translation methods
- `<locale>.go`: Implementation for each locale
- `translator.go`: Factory for creating locale-specific translators

### Usage in Go Code

```go
import "your-project/i18n"

// Create translator for a specific locale
translator := i18n.NewTranslator("sv")

// Use translations with parameters
message := translator.MenuMessage("John", 5) // "John har 5 notiser"
```

## Development

```bash
make build # => bin/simple-i18n
make test # Builds the binary and a test integration app
```

The project structure:
- `cmd/cli/`: CLI application
- `cmd/test/`: Integration test app
- `internal/`: Logic
