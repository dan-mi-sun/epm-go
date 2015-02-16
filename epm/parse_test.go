package epm

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
	(alphabet soup (brownies (marmalade (eggplant honey comb)))) => "sup"

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
	tokenLeftBraceTy,
	tokenStringTy,
	tokenLeftBraceTy,
	tokenStringTy,
	tokenStringTy,
	tokenStringTy,
	tokenRightBraceTy,
	tokenRightBraceTy,
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

var text2 = `
transact:
	$alpha => (+ (* 4 (- 9 3)) 5) => A
	"jimbo" => (+ $alpha 3)
`

func TestInterpreter(t *testing.T) {
	e, _ := NewEPM(nil, "")
	e.vars["alpha"] = "0x42"
	p := Parse(text2)
	p.run()
	args, err := e.ResolveArgs("", p.jobs[0].args)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(args)
}
