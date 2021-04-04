package parser

import (
	"bytes"
	"errors"
	"unicode"

	"github.com/sirupsen/logrus"
)

var keywords = map[string]int{
	"select": SELECT,
	"from":   FROM,
}

//go:generate go run golang.org/x/tools/cmd/goyacc -l -o sql.go sql.y
type Lexer struct {
	Input     []byte
	ParseTree Statement
	Pos       int
	Err       error
}

func NewLexer(input []byte) *Lexer {
	return &Lexer{
		Input: input,
	}
}

func (l *Lexer) Lex(lval *yySymType) int {
	sym, val := l.Scan()
	lval.str = val
	logrus.WithField("source", "lexer").Infof("scanning %s [%d]", val, sym)
	return sym
}

func (l *Lexer) Scan() (int, string) {
	for b := l.next(); b != 0; b = l.next() {
		switch {
		case unicode.IsSpace(rune(b)):
			continue
		case unicode.IsLetter(rune(b)):
			l.backup()
			sym, val := l.scanString()
			return sym, val
		default:
			switch b {
			case ';':
				return 0, ""
			case '*':
				return ASTERISK, ""
			}
			return LEX_ERROR, ""
		}
	}
	return 0, ""
}

func (l *Lexer) scanString() (int, string) {
	buf := bytes.NewBuffer(nil)
	for {
		b := l.next()
		switch {
		case unicode.IsLetter(rune(b)):
			buf.WriteByte(b)
			break
		case unicode.IsDigit(rune(b)):
			buf.WriteByte(b)
			break
		default:
			l.backup()
			str := buf.String()
			val, ok := keywords[str]
			if !ok {
				return NAME, str
			}
			return val, str
		}
	}
}

func (l *Lexer) backup() {
	if l.Pos == -1 {
		return
	}
	l.Pos--
}

func (l *Lexer) next() byte {
	if l.Pos >= len(l.Input) || l.Pos == -1 {
		l.Pos = -1
		return 0
	}
	l.Pos++
	return l.Input[l.Pos-1]
}

func (l *Lexer) Error(s string) {
	l.Err = errors.New(s)
}
