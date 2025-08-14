package frontend

// Tokenize convenient utility to turn the script into an array of tokens
func Tokenize(input string) ([]Token, error) {
	lexer := NewLexerFromString(input)
	var tokens []Token
	for {
		tok, err := lexer.NextToken()
		if err != nil {
			return nil, err
		}
		if tok.tokenType == comment {
			continue
		}
		tokens = append(tokens, *tok)
		if tok.tokenType == eof {
			break
		}
	}
	return tokens, nil
}
