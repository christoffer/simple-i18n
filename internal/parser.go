package internal

import (
	"fmt"
	"strings"
)

type TranslateFuncParam struct {
	Name string
	Type string
}

type TranslateFunc struct {
	DocString string
	Name      string
	Params    []TranslateFuncParam
	Body      string
}

func (t *TranslateFunc) Signature() string {
	return fmt.Sprintf("%s(%s) string", t.Name, t.ParamsList())
}

func (t *TranslateFunc) ParamsList() string {
	params := make([]string, len(t.Params))
	for i, param := range t.Params {
		params[i] = fmt.Sprintf("%s %s", param.Name, param.Type)
	}
	return strings.Join(params, ", ")
}

func createDocString(value string) string {
	lines := strings.Split(value, "\n")
	var docLines []string
	for _, line := range lines {
		docLines = append(docLines, "// "+line)
	}
	return strings.Join(docLines, "\n")
}

func parseTranslateFunc(tomlKey string, value string) (TranslateFunc, error) {
	tokens := tokenize(value)

	trParams := make([]TranslateFuncParam, 0)
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
					// Prepend count
					trParams = append([]TranslateFuncParam{{
						Name: "count",
						Type: "int",
					}}, trParams...)
				} else {
					trParams = append(trParams, TranslateFuncParam{
						Name: token.Value,
						Type: "string",
					})
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
				// Prepend count
				trParams = append([]TranslateFuncParam{{
					Name: "count",
					Type: "int",
				}}, trParams...)
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
			returnSingular.WriteString(singularForm)
			returnPlural.WriteString(pluralForm)
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

	// Create properly formatted multiline comment
	docString := createDocString(value)
	
	return TranslateFunc{
		Name:      toPublicName(tomlKey),
		DocString: docString,
		Params:    trParams,
		Body:      body.String(),
	}, nil
}
