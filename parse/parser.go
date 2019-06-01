package parse

import (
	"fmt"
	"github.com/icodeface/teddy/ast"
	"io"
)

func parse(l *Lexer) int {
	tok, err := l.scanner.Scan(l)
	if err != nil {
		fmt.Println("scan error", err)
	}
	fmt.Println("got token", tok.Str, ",position:", tok.Pos)
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
