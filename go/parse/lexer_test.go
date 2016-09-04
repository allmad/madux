package parse

import (
	"fmt"
	"os"
	"testing"

	"github.com/chzyer/test"
)

func TestLexer(t *testing.T) {
	defer test.New(t)

	fd, err := os.Open("testdata/lexer.txt")
	test.Nil(err)
	defer fd.Close()

	lex := Lex(fd)
	for item := range lex.items {

		fmt.Println(item)
	}
}
