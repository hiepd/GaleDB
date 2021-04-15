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
		sym, val := lexer.Scan()
		return nil, fmt.Errorf("invalid statement with: %d %s", sym, val)
	}
	return lexer.ParseTree, nil
}
