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

TODO

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
./bin/simple-i18n -i ../translations -o . -p inter -b sv
```

## Translation

Translation files specify message, template pairs in TOML files. 
The message determines the function name, and the template determines the function output. The function signature is resolved from the template of the base language.

Filenames matter. Only translation files named like a locale (`xx.toml` or `xx_yy.toml`) are processed. The locale is case-insensitive. 

E.g. `en.toml`, `de.toml`, `sv_fi.toml`, and `en_UK.toml` are all processed. A file like `english.toml` is not.


### Content
You can have messages in the root, or in a subsection (one level only). Only string values are supported.

```toml
# en.tomle
title = "Welcome to the site"

[user_page]
title = "User page"
```

### Substitutions

Substitutions are supported in the translation text with `{param}`. Note that `param` must be a valid Go identifier. 

Substitutions appear in the function signature in the order they appear in the text of the base language, _except_ for `count` which is always first. If a substitution is used more than once, only the first usage appear in the signature. 

All substitutions are of type string, except for `count` which is of type int.

```go
// en.toml
place_greeting = "In {place}, we say '{greeting}, {place}!', {count} time{{s}}"

// en.go
t.PlaceGreeting(2, "Paris", "Bonjour") // "In Paris, we say 'Bonjour, Paris!', 2 times"
```

### Pluralization

Pluralization is handled using `{{one|other}}` notation, e.g. `{{cat|cats}}`. A shorthand notation can be used to only specify the plural form `cat{{s}}`. 

Using pluralization in a text will add a `count` parameter to the function. The `{count}` parameter is also available in the text.

```go
// en.toml
apples = "You have {count} apple{{s}}. That's a lot of apple{{s}}!"
critera = "You have {count} {{criterium|criteria}}."

// en.go
t.Apples(1) // "You have 1 apple. That's a lot of apple!"
t.Apples(2) // "You have 2 apples. That's a lot of apples!"
t.Criteria(1) // "You have 1 criterium."
t.Criteria(2) // "You have 2 criteria."
```

## Generated files

The tool generates:

- `base.go`: Interface defining all translation methods. This is based on the base language.
- `<locale>.go`: Implementation for each locale
- `translator.go`: Factory for creating locale-specific translators

## Development

```bash
make build # => bin/simple-i18n
make test # Builds the binary and a test integration app
```

## License

MIT

## Credits

Inspired by the excellent [typesafe-i18n](https://github.com/ivanhofer/typesafe-i18n) for TypeScript.