package parser

import (
	"fmt"
)

type Parser struct{}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(sql string) (Statement, error) {
	lexer := NewLexer([]byte(sql))
	if yyParse(lexer) != 0 {
		if sym, val := lexer.Scan(); sym != 0 {
			return nil, fmt.Errorf("invalid statement ending with: %d %s", sym, val)
		}
	}
	return lexer.ParseTree, nil
}
