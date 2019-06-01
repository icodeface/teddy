package parse_test

import (
	"bytes"
	"fmt"
	"github.com/icodeface/teddy/parse"
	"testing"
)

func TestParse(t *testing.T) {
	code := "if (a > b) {\n	c = 1\n}"
	fmt.Printf("code:\n%v\n\n", code)
	r := bytes.NewReader([]byte(code))
	chunk, err := parse.Parse(r, "test")
	if err != nil {
		fmt.Println("error", err)
		return
	}
	fmt.Println(chunk)
}
