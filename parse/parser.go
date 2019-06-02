package parse

import (
	"github.com/icodeface/teddy/ast"
	"io"
)

func parse(lx *Lexer) int {
	for lx.Lex() > 0 {

	}
	return 0
}

func Parse(reader io.Reader, name string) (chunk []ast.Stmt, err error) {
	l := NewLexer(reader, name)
	chunk = nil
	defer func() {
		if e := recover(); e != nil {
			err, _ = e.(error)
		}
	}()
	parse(l)
	chunk = l.Stmts
	return
}
