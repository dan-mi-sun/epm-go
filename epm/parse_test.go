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

func testLexer(t *testing.T, input string, ground []tokenType) {
	l := Lex(input)
	i := 0
	for tok := range l.Chan() {
		fmt.Println(tok)
		if ground != nil && tok.typ != ground[i] {
			t.Fatal("Error", tok.typ, tokens[i])
		}
		i += 1
	}

}

func TestLexer(t *testing.T) {
	testLexer(t, text1, tokens)
}

var text1b = `
deploy:
    ok => doja
`

var tokensB = []tokenType{
	tokenNewLineTy,
	tokenCmdTy,
	tokenColonTy,
	tokenNewLineTy,
	tokenTabTy,
	tokenStringTy,
	tokenArrowTy,
	tokenStringTy,
	tokenNewLineTy,
}

func TestLexer2(t *testing.T) {
	testLexer(t, text1b, tokensB)
}

var text1c = `
!{bob
deploy:
    ok => doja
!}bob
`

var tokensC = []tokenType{
	tokenNewLineTy,
	tokenLeftDiffTy,
	tokenStringTy,
	tokenNewLineTy,
	tokenCmdTy,
	tokenColonTy,
	tokenNewLineTy,
	tokenTabTy,
	tokenStringTy,
	tokenArrowTy,
	tokenStringTy,
	tokenNewLineTy,
	tokenRightDiffTy,
	tokenStringTy,
	tokenNewLineTy,
}

func TestLexer3(t *testing.T) {
	testLexer(t, text1c, tokensC)
}

// TODO: proper test
func TestParse(t *testing.T) {
	p := Parse(text1)
	p.run()
	printJobs(p.jobs)
}

func printJobs(jobs []Job) {
	for _, j := range jobs {
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

var text2b = `
transact:
	$alpha => (+ (* {{A}} (- 9 {{B}})) 5) => A
	"jimbo" => (+ $alpha 3)
`

func TestInterpreter2(t *testing.T) {
	e, _ := NewEPM(nil, "")
	e.vars["alpha"] = "0x42"
	e.vars["A"] = "0x5"
	e.vars["B"] = "0x3"
	p := Parse(text2b)
	p.run()
	args, err := e.ResolveArgs("", p.jobs[0].args)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(args)
}

var text3 = `
deploy:
	"a.lll" => {{BOB}}
`

var text4 = `
transact:
	{{BOB}} => "jim" 0x34 (+ (* 2 0x4) 0x1)
`

var text5 = `
!{bob
transact:
	0x5 => "jim" 0x34 (+ (* 2 0x4) 0x1)
!}bob
`

func TestDeploy(t *testing.T) {
	p := Parse(text3)
	p.run()
	// setup EPM object with ChainInterface
	e, _ := NewEPM(nil, "")

	e.jobs = p.jobs
	printJobs(e.jobs)

	// epm execute jobs
	e.ExecuteJobs()
}

func TestTransact(t *testing.T) {
	p := Parse(text4)
	p.run()
	// setup EPM object with ChainInterface
	e, _ := NewEPM(nil, "")

	e.jobs = p.jobs
	printJobs(e.jobs)

	// epm execute jobs
	e.ExecuteJobs()
}

func TestDiff(t *testing.T) {
	p := Parse(text5)
	p.run()
	// setup EPM object with ChainInterface
	e, _ := NewEPM(nil, "")

	e.jobs = p.jobs
	printJobs(e.jobs)
	fmt.Println("Diff sched:", p.diffsched)

	// epm execute jobs
	e.ExecuteJobs()
}
