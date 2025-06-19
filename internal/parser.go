package internal

import (
	"fmt"
	"strings"
)

type TranslateFunc struct {
	DocString string
	Name      string
	Params    []string
	Body      string
}

func (t *TranslateFunc) Signature() string {
	return fmt.Sprintf("%s(%s) string\n", t.Name, strings.Join(t.Params, ", "))
}

func parseTranslateFunc(tomlKey string, value string) (TranslateFunc, error) {
	tokens := tokenize(value)

	funcArgs := make([]string, 0)
	seenFuncArgs := make(map[string]bool)
	fmtArgs := make([]string, 0)

	var returnSingular strings.Builder
	var returnPlural strings.Builder

	hasPlural := false

	for _, token := range tokens {
		if token.Error != "" {
			return TranslateFunc{}, fmt.Errorf("syntax error: %s, in `%s = \"%s\"`", token.Error, tomlKey, value)
		}
		switch token.Type {
		case TokenText:
			escapedValue := strings.ReplaceAll(token.Value, `%`, `%%`)
			returnSingular.WriteString(escapedValue)
			returnPlural.WriteString(escapedValue)
		case TokenSub:
			if !seenFuncArgs[token.Value] {
				if token.Value == "count" {
					funcArgs = append([]string{"count int"}, funcArgs...)
				} else {
					funcArgs = append(funcArgs, token.Value+" string")
				}
				seenFuncArgs[token.Value] = true
			}
			fmtArgs = append(fmtArgs, token.Value)
			placeholder := "%s"
			if token.Value == "count" {
				placeholder = "%d"
			}
			returnSingular.WriteString(placeholder)
			returnPlural.WriteString(placeholder)
		case TokenPlural:
			if !seenFuncArgs["count"] {
				funcArgs = append([]string{"count int"}, funcArgs...)
				seenFuncArgs["count"] = true
			}
			hasPlural = true
			// Split on |, singular first and plural second
			parts := strings.Split(token.Value, "|")
			singularForm := parts[0]
			var pluralForm string
			if len(parts) == 1 {
				singularForm = ""
				pluralForm = parts[0]
			} else {
				pluralForm = strings.Join(parts[1:], "")
			}
			returnSingular.WriteString(pluralForm)
			returnPlural.WriteString(singularForm)
		}
	}

	var body strings.Builder

	if hasPlural {
		body.WriteString("\tif count == 1 {\n")
		genSprintfReturn(&body, returnSingular.String(), fmtArgs)
		body.WriteString("\t} else {\n")
		genSprintfReturn(&body, returnPlural.String(), fmtArgs)
		body.WriteString("\t}\n")
	} else {
		genSprintfReturn(&body, returnSingular.String(), fmtArgs)
	}

	return TranslateFunc{
		Name:      toPublicName(tomlKey),
		DocString: "// " + value,
		Params:    funcArgs,
		Body:      body.String(),
	}, nil
}
