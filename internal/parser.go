package internal

import "strings"

type Token struct {
	Type  TokenType
	Value string
	Start int // Index into input string, inclusive
	End   int // Index into input string, exclusive
}

type TokenType int

const (
	TokenText TokenType = iota
	TokenSub
	TokenPlural
)

type ParseResult struct {
	HasPlural    bool
	Variables    []string
	SingularForm string
	PluralForm   string
}

func tokenize(input string) []Token {
	tokens := make([]Token, 0)
	i := 0
	tokenStart := 0
	curType := TokenText

	peek := func() byte {
		if i+1 < len(input) {
			return input[i+1]
		}
		return 0
	}

	endToken := func(t TokenType, tokenEnd int) {
		if tokenStart == tokenEnd {
			return // no content
		}
		value := input[tokenStart:tokenEnd]
		if curType == TokenSub {
			value = input[tokenStart+1 : tokenEnd-1]
		}
		if curType == TokenPlural {
			value = input[tokenStart+2 : tokenEnd-2]
		}
		tokens = append(tokens, Token{
			Type:  curType,
			Value: value,
			Start: tokenStart,
			End:   tokenEnd,
		})
	}

	for i < len(input) {
		c := input[i]
		if c == '{' {
			if curType == TokenText {
				endToken(TokenText, i)
				tokenStart = i
				if peek() == '{' {
					i += 1 // skip next '{'
					curType = TokenPlural
				} else {
					curType = TokenSub
				}
			}
		}

		if c == '}' {
			if curType == TokenSub {
				endToken(TokenSub, i+1)
				tokenStart = i + 1
				curType = TokenText
			} else if curType == TokenPlural {
				if peek() == '}' {
					endToken(TokenPlural, i+2)
					tokenStart = i + 2
					i += 1 // skip peeked '}'
					curType = TokenText
				}
			}
		}

		i++
	}

	// End current token if there's text left
	if tokenStart < len(input) {
		endToken(curType, len(input))
	}

	return tokens
}

type TranslateFunc struct {
	Params []string
	Body   string
}

func parseTranslateFunc(value string) (TranslateFunc, error) {
	tokens := tokenize(value)

	funcArgs := make([]string, 0)
	seenFuncArgs := make(map[string]bool)
	fmtArgs := make([]string, 0)

	var returnSingular strings.Builder
	var returnPlural strings.Builder

	hasPlural := false

	for _, token := range tokens {
		switch token.Type {
		case TokenText:
			// TODO(christoffer): Escape %
			returnSingular.WriteString(token.Value)
			returnPlural.WriteString(token.Value)
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
		Params: funcArgs,
		Body:   body.String(),
	}, nil
}
