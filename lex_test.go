package gps

import "testing"
import "fmt"

func getAllTokens(l *Lexer) []Token {
	var tokens = []Token{}
	for {
		tok, err := l.Tokenize()
		fmt.Printf("%d\t%s\n", tok.Type, tok.String())
		tokens = append(tokens, tok)
		if err != nil {
			break
		}
	}
	return tokens
}

func TestLex(t *testing.T) {
	l := NewLexer("test", "select * from `table` where `a` = xyz")
	toks := getAllTokens(l)
	fmt.Println(toks)
}
