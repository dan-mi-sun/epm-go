package parse

import (
	"fmt"
	"testing"
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

var tokens = []tokenType{
	tokenNewLineTy,
	tokenPoundTy,
	tokenStringTy,
	tokenNewLineTy,
	tokenNewLineTy,
	tokenCmdTy,
	tokenColonTy,
	tokenNewLineTy,
	tokenTabTy,
	tokenQuoteTy,
	tokenStringTy,
	tokenQuoteTy,
	tokenArrowTy,
	tokenQuoteTy,
	tokenStringTy,
	tokenQuoteTy,
	tokenNewLineTy,
	tokenNewLineTy,
	tokenPoundTy,
	tokenStringTy,
	tokenNewLineTy,
	tokenNewLineTy,
	tokenPoundTy,
	tokenStringTy,
	tokenNewLineTy,
	tokenNewLineTy,
	tokenCmdTy,
	tokenColonTy,
	tokenNewLineTy,
	tokenTabTy,
	tokenQuoteTy,
	tokenStringTy,
	tokenQuoteTy,
	tokenArrowTy,
	tokenQuoteTy,
	tokenStringTy,
	tokenQuoteTy,
	tokenArrowTy,
	tokenLeftBracesTy,
	tokenStringTy,
	tokenRightBracesTy,
	tokenNewLineTy,
	tokenNewLineTy,
	tokenTabTy,
	tokenQuoteTy,
	tokenStringTy,
	tokenQuoteTy,
	tokenArrowTy,
	tokenLeftBracesTy,
	tokenNumberTy,
	tokenRightBracesTy,
	tokenNewLineTy,
	tokenNewLineTy,
	tokenNewLineTy,
}

func TestLexer(t *testing.T) {
	fmt.Println([]byte(text))
	l := Lex(text)
	i := 0
	for tok := range l.Chan() {
		fmt.Println(tok)
		if tok.typ != tokens[i] {
			t.Fatal("Error", tok.typ, tokens[i])
		}
		i += 1
	}
}
