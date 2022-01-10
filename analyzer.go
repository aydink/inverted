package main

import "unicode"

type Token struct {
	start, end, position uint16
	value                string
}

type Tokenizer interface {
	Tokenize(string) []Token
}

type TokenFilterer interface {
	Filter([]Token) []Token
}

type Analyzer interface {
	Analyze(string) []Token
}

type SimpleAnalyzer struct {
	tokenizer    Tokenizer
	tokenFilters []TokenFilterer
}

func NewSimpleAnalyzer(t Tokenizer) *SimpleAnalyzer {
	return &SimpleAnalyzer{t, make([]TokenFilterer, 0)}
}

func (sa *SimpleAnalyzer) AddTokenFilter(f TokenFilterer) {
	sa.tokenFilters = append(sa.tokenFilters, f)
}

func (sa *SimpleAnalyzer) Analyze(s string) []Token {
	t := sa.tokenizer.Tokenize(s)
	for _, tf := range sa.tokenFilters {
		t = tf.Filter(t)
	}
	return t
}

/*

var f = func(c rune) bool {
	return !unicode.IsLetter(c) && !unicode.IsNumber(c)
}
*/

type SimpleTokenizer struct{}

func NewSimpleTokenizer() SimpleTokenizer {
	return SimpleTokenizer{}
}

func (tk SimpleTokenizer) Tokenize(s string) []Token {
	var posToken uint16 = 0

	tokens := []Token{}
	token := Token{start: 0, end: 0, position: 0, value: ""}

	insideToken := false

	for pos, char := range s {
		if unicode.IsLetter(char) || unicode.IsNumber(char) {
			if !insideToken {
				insideToken = true
				token.start = uint16(pos)
			}
		} else {
			if insideToken {
				insideToken = false
				token.end = uint16(pos)
				token.value = s[token.start:token.end]

				// handle zero length tokens
				if token.start != token.end {
					token.position = posToken
					posToken++
					tokens = append(tokens, token)
				}
			}
		}
	}

	if insideToken {
		token.end = uint16(len(s))
		token.value = s[token.start:token.end]

		// handle zero length tokens
		if token.start != token.end {
			token.position = posToken
			posToken++
			tokens = append(tokens, token)
		}
	}

	return tokens
}

func getUniqueTokens(tokens []Token) []Token {
	uniqueTokens := make(map[string]Token)

	for _, token := range tokens {
		uniqueTokens[token.value] = token
	}

	temp := make([]Token, len(uniqueTokens))

	for _, v := range uniqueTokens {
		temp = append(temp, v)
	}

	return temp
}
