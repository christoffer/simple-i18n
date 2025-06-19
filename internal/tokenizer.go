package internal

type Token struct {
	Type  TokenType
	Value string
	Start int // Index into input string, inclusive
	End   int // Index into input string, exclusive
	Error string
}

type TokenType int

const (
	TokenText TokenType = iota
	TokenSub
	TokenPlural
)

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

	endToken := func(t TokenType, tokenEnd int, isPremature bool) {
		errorMessage := ""
		if tokenStart == tokenEnd {
			return // no content
		}
		value := input[tokenStart:tokenEnd]
		if curType == TokenSub {
			if isPremature {
				errorMessage = "missing end '}'"
				value = input[tokenStart+1 : tokenEnd]
			} else {
				value = input[tokenStart+1 : tokenEnd-1]
			}
		}
		if curType == TokenPlural {
			if isPremature {
				errorMessage = "missing end '}}'"
				value = input[tokenStart+2 : tokenEnd]
			} else {
				value = input[tokenStart+2 : tokenEnd-2]
			}
		}
		tokens = append(tokens, Token{
			Type:  curType,
			Value: value,
			Start: tokenStart,
			End:   tokenEnd,
			Error: errorMessage,
		})
	}

	for i < len(input) {
		c := input[i]
		if c == '{' {
			if curType == TokenText {
				endToken(TokenText, i, false)
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
				endToken(TokenSub, i+1, false)
				tokenStart = i + 1
				curType = TokenText
			} else if curType == TokenPlural {
				if peek() == '}' {
					endToken(TokenPlural, i+2, false)
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
		endToken(curType, len(input), true)
	}

	return tokens
}
