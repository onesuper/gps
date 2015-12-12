package gps

import (
	"bytes"
	"errors"
	"fmt"
	// "log"
	"strings"
	"unicode/utf8"
)

type TokenType int

const (
	Error TokenType = iota
	EOF
	Star
	Sep
	Dot
	Op
	Paren
	Literal
	Number
	String
	DblString
	// KeyWord
	Select
	Distinct
	From
	Where
	Group
	Order
	By
	Having
	Limit
	Join
	Left
	Right
	Inner
	Outer
	On
	As
	Union
	All
	And
	Or
	Between
	True
	False
	Null
	Is
	Not
	Like
	Exists
)

type Token struct {
	Type    TokenType
	Literal string
}

func (t *Token) String() string {
	switch t.Type {
	case EOF:
		return "EOF"
	case Error:
		return t.Literal
	}
	return fmt.Sprintf("%q", t.Literal)
}

type Lexer struct {
	name   string
	input  string
	start  int
	pos    int
	tokens chan Token
}

func NewLexer(name, input string) *Lexer {
	l := &Lexer{
		name:   name,
		input:  input,
		tokens: make(chan Token),
	}
	go l.transform()
	return l
}

func (l *Lexer) Tokenize() (Token, error) {
	t := <-l.tokens
	if t.Type == Error {
		return t, errors.New("Error")
	} else if t.Type == EOF {
		return t, errors.New("EOF")
	}
	return t, nil
}

func (l *Lexer) debugString() string {
	var buffer bytes.Buffer

	for i, w := 0, 0; i < len(l.input); i += w {
		r, width := utf8.DecodeRuneInString(l.input[i:])
		w = width
		if i == l.start {
			buffer.WriteString(".")
		}
		if i == l.pos {
			buffer.WriteString(".")
		}
		buffer.WriteRune(r)

	}
	return buffer.String()
}

// When we recognize a token, we move on with the cursor and
// call this func to emit it back to the caller.
func (l *Lexer) emit(t TokenType) {
	l.tokens <- Token{t, l.cache()}
	l.start = l.pos
}

func (l *Lexer) cache() string {
	return l.input[l.start:l.pos]
}

// ignore the current prune
func (l *Lexer) ignore() {
	l.start = l.pos
}

// eat the next Rune from input
func (l *Lexer) next() rune {
	r, width := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += width
	// log.Printf("[next] %s", l.debugString())
	return r
}

// backup to the start
func (l *Lexer) backup() {
	l.pos = l.start
	// log.Printf("[back] %s", l.debugString())
}

// peek without move forward
func (l *Lexer) peek() rune {
	r, _ := utf8.DecodeRuneInString(l.input[l.pos:])
	// log.Printf("[peek] %q", r)
	return r
}

// determing whether the next Rune is valid
func (l *Lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.peek()) >= 0 {
		return true
	}
	return false
}

// put the error message to the token literal
func (l *Lexer) errorf(format string, args interface{}) {
	prefixed := fmt.Sprintf("ERROR: %s: %s", l.name, format)
	l.tokens <- Token{Error, fmt.Sprintf(prefixed, args)}
}

// The state is a function which takes as input the lexer and return
// a state, which takes as input the lexer and return a state...
type State func(*Lexer) State

func (l *Lexer) transform() {
	for state := expectAny; state != nil; {
		state = state(l)
	}
	close(l.tokens)
}

// The inital state
func expectAny(l *Lexer) State {
	if l.pos >= len(l.input) {
		l.emit(EOF)
		return nil
	}

	switch r := l.next(); {
	case r == '\'':
		return expectString
	case r == '`':
		return expectLiteral
	case r == ' ' || r == '\n':
		l.ignore()
		return expectAny
	case r == '*':
		l.emit(Star)
		return expectAny
	case r == ',':
		l.emit(Sep)
		return expectAny(l)
	case r == '=' || r == '+' || r == '-' || r == '/':
		l.emit(Op)
		return expectAny
	case r == '!':
		if l.peek() != '=' {
			l.errorf("unsupported op: ", l.cache())
		}
		l.emit(Op)
		return expectAny
	case r == '>':
		if l.peek() != '=' && l.peek() != ' ' {
			l.errorf("unsupported op: ", l.cache())
		}
		l.emit(Op)
		return expectAny
	case r == '<':
		if l.peek() != '=' && l.peek() != ' ' && l.peek() != '>' {
			l.errorf("unsupported op: ", l.cache())
		}
		l.emit(Op)
		return expectAny
	case '0' <= r && r <= '9':
		l.backup()
		return expectNumber
	case 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z':
		l.backup()
		return expectKeyword
	default:
		return nil
	}
}

func expectKeyword(l *Lexer) State {
	for r := l.peek(); 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z'; r = l.peek() {
		l.next()
	}

	switch strings.ToUpper(l.cache()) {
	case "SELECT":
		l.emit(Select)
	case "DISTINCT":
		l.emit(Distinct)
	case "FROM":
		l.emit(From)
	case "WHERE":
		l.emit(Where)
	case "GROUP":
		l.emit(Group)
	case "ORDER":
		l.emit(Order)
	case "BY":
		l.emit(By)
	case "HAVING":
		l.emit(Having)
	case "LIMIT":
		l.emit(Limit)
	case "JOIN":
		l.emit(Join)
	case "LEFT":
		l.emit(Left)
	case "RIGHT":
		l.emit(Right)
	case "INNER":
		l.emit(Inner)
	case "OUTER":
		l.emit(Outer)
	case "ON":
		l.emit(On)
	case "AS":
		l.emit(As)
	case "UNION":
		l.emit(Union)
	case "ALL":
		l.emit(All)
	default:
		l.errorf("keyword doesn't exsit: %s", l.cache())
		return nil
	}
	return expectAny
}

func expectString(l *Lexer) State {
	if l.next() == '\'' {
		l.emit(String)
		return expectAny
	} else {
		return expectString
	}
}

func expectLiteral(l *Lexer) State {
	if l.next() == '`' {
		l.tokens <- Token{t, l.cache()}
		l.start = l.pos
		return expectAny
	} else {
		return expectLiteral
	}
}

func expectNumber(l *Lexer) State {
	for '0' <= l.peek() && l.peek() <= '9' {
		l.next()
	}
	if l.peek() == '.' {
		l.next()
	}
	for '0' <= l.peek() && l.peek() <= '9' {
		l.next()
	}
	l.emit(Number)
	return expectAny
}
