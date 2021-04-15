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
	"where":  WHERE,
	"and":    AND,
	"=":      RELATION,
	"<":      RELATION,
	">":      RELATION,
	">=":     RELATION,
	"<=":     RELATION,
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
	switch v := val.(type) {
	case string:
		lval.str = v
	case int:
		lval.num = v
	default:
		panic("invalid lexing type")
	}
	logrus.WithField("source", "lexer").Debugf("scanning %s [%d]", val, sym)
	return sym
}

func (l *Lexer) Scan() (int, interface{}) {
	for b := l.next(); b != 0; b = l.next() {
		switch {
		case unicode.IsSpace(rune(b)):
			continue
		case unicode.IsLetter(rune(b)):
			l.backup()
			sym, val := l.scanString()
			return sym, val
		case unicode.IsDigit(rune(b)):
			l.backup()
			sym, val := l.scanNumber()
			return sym, val
		default:
			switch b {
			case '=', '<', '>':
				l.backup()
				sym, val := l.scanRelation()
				return sym, val
			case '+', '-':
				return OPERATOR, string(b)
			case ',':
				return COMMA, ","
			case ';':
				return 0, ";"
			default:
				return LEX_ERROR, string(b)
			}
		}
	}
	return 0, ""
}

func (l *Lexer) scanRelation() (int, string) {
	buf := bytes.NewBuffer(nil)
	for {
		b := l.next()
		switch b {
		case '=', '<', '>':
			buf.WriteByte(b)
		default:
			l.backup()
			str := buf.String()
			val, ok := keywords[str]
			if !ok {
				return LEX_ERROR, str
			}
			return val, str
		}
	}
}

func (l *Lexer) scanString() (int, string) {
	buf := bytes.NewBuffer(nil)
	for {
		b := l.next()
		if unicode.IsLetter(rune(b)) || unicode.IsDigit(rune(b)) || b == '_' {
			buf.WriteByte(b)
		} else {
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

func (l *Lexer) scanNumber() (int, int) {
	res := 0
	for {
		b := l.next()
		if unicode.IsDigit(rune(b)) {
			res *= 10
			res += int(b - '0')
		} else if unicode.IsSpace(rune(b)) || b == ';' {
			l.backup()
			return NUMBER, res
		} else {
			return LEX_ERROR, 0
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
