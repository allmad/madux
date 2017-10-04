package parser

import (
	"fmt"
	"os"
	"testing"

	"github.com/chzyer/test"
)

func TestLexer(t *testing.T) {
	defer test.New(t)

	fd, err := os.Open("testdata/zsh.txt")
	test.Nil(err)
	defer fd.Close()

	lex := Lex(fd)
	token := NewToken(lex)
	go token.Run()
	for item := range token.out {
		fmt.Println(item)
	}
}
