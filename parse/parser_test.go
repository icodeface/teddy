package parse_test

import (
	"bytes"
	"fmt"
	"github.com/icodeface/teddy/parse"
	"testing"
)

func TestParse(t *testing.T) {
	code := `
a = 2
b = 1 /* this is 
a multi line 
comment 
*/
if (a > b) {
	c = 1
	d = "abc123~" //fuck
	e = 1 + a * b
	return e
}`
	fmt.Printf("code:\n%v\n\n", code)
	r := bytes.NewReader([]byte(code))
	chunk, err := parse.Parse(r, "test.teddy")
	if err != nil {
		fmt.Println("error", err)
		return
	}
	fmt.Println(chunk)
}
