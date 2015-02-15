package parse

import (
	"fmt"
	"testing"
)

var text1 = `
# this is a comment

deploy:
	"something.lll" => "something else"

 # this is another comment

# is this too?

transact:
	"ok" => "dokay" => monkey

	$monkey => "nice" => 0x43
	(alphabet soup (brownies marmalade)) => "sup"

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
	tokenStringTy,
	tokenNewLineTy,
	tokenNewLineTy,
	tokenTabTy,
	tokenBlingTy,
	tokenStringTy,
	tokenArrowTy,
	tokenQuoteTy,
	tokenStringTy,
	tokenQuoteTy,
	tokenArrowTy,
	tokenNumberTy,
	tokenNewLineTy,
	tokenTabTy,
	tokenLeftBraceTy,
	tokenStringTy,
	tokenStringTy,
	tokenLeftBraceTy,
	tokenStringTy,
	tokenStringTy,
	tokenRightBraceTy,
	tokenRightBraceTy,
	tokenArrowTy,
	tokenQuoteTy,
	tokenStringTy,
	tokenQuoteTy,
	tokenNewLineTy,
	tokenNewLineTy,
}

func TestLexer(t *testing.T) {
	fmt.Println([]byte(text1))
	l := Lex(text1)
	i := 0
	for tok := range l.Chan() {
		fmt.Println(tok)
		if tok.typ != tokens[i] {
			t.Fatal("Error", tok.typ, tokens[i])
		}
		i += 1
	}
}

// TODO: proper test
func TestParse(t *testing.T) {
	p := Parse(text1)
	p.run()
	for _, j := range p.jobs {
		fmt.Println("##########")
		fmt.Println(j.cmd, len(j.args))
		for _, a := range j.args {
			//fmt.Println(a[0].token.val)
			PrintTree(a[0])
		}

	}
}
