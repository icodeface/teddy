package parse

// take idea from github.com/yuin/gopher-lua/ast

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/icodeface/teddy/ast"
	"io"
	"strconv"
)

const EOF = -1
const whitespace1 = 1<<'\t' | 1<<' '
const whitespace2 = 1<<'\t' | 1<<'\n' | 1<<'\r' | 1<<' '

type Error struct {
	Pos     ast.Position
	Message string
	Token   string
}

func (e *Error) Error() string {
	pos := e.Pos
	if pos.Line == EOF {
		return fmt.Sprintf("%v at EOF:   %s\n", pos.Source, e.Message)
	} else {
		return fmt.Sprintf("%v line:%d(column:%d) near '%v':   %s\n", pos.Source, pos.Line, pos.Column, e.Token, e.Message)
	}
}

// 十进制数 0～9
func isDecimal(ch int) bool { return '0' <= ch && ch <= '9' }

// 标识符：变量名、函数名、类名、
func isIdent(ch int, pos int) bool {
	return ch == '_' || 'A' <= ch && ch <= 'Z' || 'a' <= ch && ch <= 'z' || isDecimal(ch) && pos > 0
}

// 16进制数
func isDigit(ch int) bool {
	return '0' <= ch && ch <= '9' || 'a' <= ch && ch <= 'f' || 'A' <= ch && ch <= 'F'
}

func writeChar(buf *bytes.Buffer, c int) { buf.WriteByte(byte(c)) }

type Scanner struct {
	Pos    ast.Position
	reader *bufio.Reader
}

func NewScanner(reader io.Reader, source string) *Scanner {
	return &Scanner{
		Pos: ast.Position{
			Source: source,
			Line:   1,
			Column: 0,
		},
		reader: bufio.NewReaderSize(reader, 4096),
	}
}

func (sc *Scanner) Error(tok string, msg string) *Error { return &Error{sc.Pos, msg, tok} }

func (sc *Scanner) skipWhiteSpace(whitespace int64) int {
	ch := sc.Next()
	for ; whitespace&(1<<uint(ch)) != 0; ch = sc.Next() {
	}
	return ch
}

func (sc *Scanner) readNext() int {
	ch, err := sc.reader.ReadByte()
	if err == io.EOF {
		return EOF
	}
	return int(ch)
}

func (sc *Scanner) Newline(ch int) {
	if ch < 0 {
		return
	}
	sc.Pos.Line += 1
	sc.Pos.Column = 0
	next := sc.Peek()
	if ch == '\n' && next == '\r' || ch == '\r' && next == '\n' {
		sc.reader.ReadByte()
	}
}

func (sc *Scanner) Next() int {
	ch := sc.readNext()
	switch ch {
	case '\n', '\r':
		sc.Newline(ch)
		ch = int('\n')
	case EOF:
		sc.Pos.Line = EOF
		sc.Pos.Column = 0
	default:
		sc.Pos.Column++
	}
	return ch
}

func (sc *Scanner) Peek() int {
	ch := sc.readNext()
	if ch != EOF {
		sc.reader.UnreadByte()
	}
	return ch
}

func (sc *Scanner) scanIdent(ch int, buf *bytes.Buffer) error {
	writeChar(buf, ch)
	for isIdent(sc.Peek(), 1) {
		writeChar(buf, sc.Next())
	}
	return nil
}

// 读取10进制数
func (sc *Scanner) scanDecimal(ch int, buf *bytes.Buffer) error {
	writeChar(buf, ch)
	for isDecimal(sc.Peek()) {
		writeChar(buf, sc.Next())
	}
	return nil
}

// 读取数字，包括10进制和16进制、科学计数法、浮点数
func (sc *Scanner) scanNumber(ch int, buf *bytes.Buffer) error {
	if ch == '0' { // octal
		if sc.Peek() == 'x' || sc.Peek() == 'X' {
			writeChar(buf, ch)
			writeChar(buf, sc.Next())
			hasvalue := false
			for isDigit(sc.Peek()) {
				writeChar(buf, sc.Next())
				hasvalue = true
			}
			if !hasvalue {
				return sc.Error(buf.String(), "illegal hexadecimal number")
			}
			return nil
		} else if sc.Peek() != '.' && isDecimal(sc.Peek()) {
			ch = sc.Next()
		}
	}
	sc.scanDecimal(ch, buf)
	if sc.Peek() == '.' {
		sc.scanDecimal(sc.Next(), buf)
	}
	if ch = sc.Peek(); ch == 'e' || ch == 'E' {
		writeChar(buf, sc.Next())
		if ch = sc.Peek(); ch == '-' || ch == '+' {
			writeChar(buf, sc.Next())
		}
		sc.scanDecimal(sc.Next(), buf)
	}

	return nil
}

func (sc *Scanner) scanString(quote int, buf *bytes.Buffer) error {
	ch := sc.Next()
	for ch != quote {
		if ch == '\n' || ch == '\r' || ch < 0 {
			return sc.Error(buf.String(), "unterminated string")
		}
		if ch == '\\' {
			if err := sc.scanEscape(ch, buf); err != nil {
				return err
			}
		} else {
			writeChar(buf, ch)
		}
		ch = sc.Next()
	}
	return nil
}

func (sc *Scanner) scanEscape(ch int, buf *bytes.Buffer) error {
	ch = sc.Next()
	switch ch {
	case 'a':
		buf.WriteByte('\a')
	case 'b':
		buf.WriteByte('\b')
	case 'f':
		buf.WriteByte('\f')
	case 'n':
		buf.WriteByte('\n')
	case 'r':
		buf.WriteByte('\r')
	case 't':
		buf.WriteByte('\t')
	case 'v':
		buf.WriteByte('\v')
	case '\\':
		buf.WriteByte('\\')
	case '"':
		buf.WriteByte('"')
	case '\'':
		buf.WriteByte('\'')
	case '\n':
		buf.WriteByte('\n')
	case '\r':
		buf.WriteByte('\n')
		sc.Newline('\r')
	default:
		if '0' <= ch && ch <= '9' {
			bytes := []byte{byte(ch)}
			for i := 0; i < 2 && isDecimal(sc.Peek()); i++ {
				bytes = append(bytes, byte(sc.Next()))
			}
			val, _ := strconv.ParseInt(string(bytes), 10, 32)
			writeChar(buf, int(val))
		} else {
			writeChar(buf, ch)
		}
	}
	return nil
}

func (sc *Scanner) Scan(lexer *Lexer) (ast.Token, error) {
redo:
	var err error
	tok := ast.Token{}
	newline := false

	// 跳过开头的空白符（空格、制表符）
	ch := sc.skipWhiteSpace(whitespace1)

	// 检测换行
	if ch == '\n' || ch == '\r' {
		newline = true
		// 跳过换行以及空白符
		ch = sc.skipWhiteSpace(whitespace2)
	}

	// 如果遇到 )\n( 标记 PNewLine 为 true
	// 否则 PNewLine 仍为 false
	if ch == '(' && lexer.PrevTokenType == ')' {
		lexer.PNewLine = newline
	} else {
		lexer.PNewLine = false
	}

	// 现在正式读到有用的内容了
	var _buf bytes.Buffer
	buf := &_buf
	tok.Pos = sc.Pos

	switch {
	case isIdent(ch, 0):
		// 标识符
		fmt.Println("TIdent", TIdent)
		tok.Type = TIdent
		// 循环读后续
		err = sc.scanIdent(ch, buf)
		tok.Str = buf.String()
		if err != nil {
			goto finally
		}
		// 如果是内置关键字，转化为具体的token type
		if typ, ok := tokenName2Type[tok.Str]; ok {
			tok.Type = typ
		}
	case isDecimal(ch):
		// 数字
		tok.Type = TNumber
		err = sc.scanNumber(ch, buf)
		tok.Str = buf.String()
	default:
		// EOF 括号
		switch ch {
		case EOF:
			tok.Type = EOF
		case '"', '\'':
			// 单双引号
			tok.Type = TString
			err = sc.scanString(ch, buf)
			tok.Str = buf.String()
		case '=':
			if sc.Peek() == '=' {
				tok.Type = TEqeq
				tok.Str = "=="
				sc.Next()
			} else {
				tok.Type = ch
				tok.Str = string(ch)
			}
		case '!':
			if sc.Peek() == '=' {
				tok.Type = TNeq
				tok.Str = "!="
				sc.Next()
			} else {
				err = sc.Error("!", "Invalid '!' token")
			}
		case '<':
			if sc.Peek() == '=' {
				tok.Type = TLte
				tok.Str = "<="
				sc.Next()
			} else {
				tok.Type = ch
				tok.Str = string(ch)
			}
		case '>':
			if sc.Peek() == '=' {
				tok.Type = TGte
				tok.Str = ">="
				sc.Next()
			} else {
				tok.Type = ch
				tok.Str = string(ch)
			}
		case '/':
			if sc.Peek() == '/' {
				// 注释, 一直读到行尾
				for {
					if ch == '\n' || ch == '\r' || ch < 0 {
						break
					}
					ch = sc.Next()
				}
				goto redo
			} else {
				tok.Type = ch
				tok.Str = string(ch)
			}
		case '+', '*', '%', '^', '#', '(', ')', '{', '}', ']', ';', ':', ',':
			// + - * / , { ( [ 等单字节的关键字 将ascii码值作为type值
			tok.Type = ch
			tok.Str = string(ch)
		default:
			writeChar(buf, ch)
			err = sc.Error(buf.String(), "Invalid token")
			goto finally
		}
	}

finally:
	return tok, err
}

type Lexer struct {
	scanner       *Scanner
	Stmts         []ast.Stmt
	PNewLine      bool // processed a new line ?
	Token         ast.Token
	PrevTokenType int
}

func NewLexer(reader io.Reader, name string) *Lexer {
	return &Lexer{NewScanner(reader, name), nil, false, ast.Token{Str: ""}, 0}
}
