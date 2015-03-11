package main

import (
	"github.com/eris-ltd/epm-go/Godeps/_workspace/src/github.com/ebuchman/lexer"
	"log"
)

var text = `
# this is a comment

deploy:
	"something.lll" => "something else"

 # this is another comment
  # is this too?

transact:
	"ok" => "dokay" => {{ monkey }} 

	"nice" => {{ 0x43 }}

`

func main() {
	l := lexer.Lex(text)
	for t := range l.Chan() {
		log.Println(t)
	}

}
