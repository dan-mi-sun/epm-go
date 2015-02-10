package parse

import (
	"fmt"
	"log"
)

type parseStateFunc func(p *parser) parseStateFunc

type parser struct {
	l    *lexer
	last token
	jobs []job // jobs to execute

	peekCount int // 1 if we've peeked

	inJob bool // are we in a job

	arg []tree // current arg

	tree  *tree // top of current tree
	treeP *tree // a pointer into current tree
	job   *job  // current job
}

type job struct {
	cmd  string
	args [][]tree
}

type tree struct {
	token    token
	parent   token
	children []token

	identifier bool // is the token a variable reference
}

func Parse(input string) *parser {
	l := Lex(input)
	p := &parser{
		l:    l,
		jobs: []job{},
		tree: new(tree),
		job:  new(job),
	}
	//go p.run()
	return p
}

func (p *parser) next() token {
	if p.peekCount == 1 {
		p.peekCount = 0
		return p.last

	}
	p.last = <-p.l.Chan()
	return p.last
}

func (p *parser) peek() token {
	if p.peekCount == 1 {
		return p.last
	}
	p.next()
	p.peekCount = 1
	return p.last
}

func (p *parser) backup() {
	p.peekCount = 1
}

func (p *parser) run() {
	for state := parseStateStart; state != nil; state = state(p) {
	}
	if p.last.typ == tokenErrTy {
		// PRINT ERROR!
	}
}

func (p *parser) accept(options []token) bool {
	return true
}

// return a parseStateFunc that prints the error and triggers exit (returns nil)
// closures++
func (p *parser) Error(s string) parseStateFunc {
	return func(pp *parser) parseStateFunc {
		// TODO: print location too
		log.Println("Error:", s)
		return nil
	}

}

func parseStateStart(p *parser) parseStateFunc {
	t := p.next()
	// scan past spaces, new lines, and comments
	switch t.typ {
	case tokenErrTy:
		return nil
	case tokenNewLineTy, tokenTabTy, tokenSpaceTy:
		return parseStateStart
	case tokenPoundTy:
		return parseStateComment
	case tokenCmdTy:
		cmd := t.val
		t = p.next()
		if t.typ != tokenColonTy {
			return p.Error("Commands must be followed by a colon")
		}
		j := &job{
			cmd:  cmd,
			args: [][]tree{},
		}
		//p.jobs = append(p.jobs, j)
		p.job = j
		return parseStateCommand
	}

	return p.Error(fmt.Sprintf("Unknown expression while looking for command: %s", t.val))
}

func parseStateComment(p *parser) parseStateFunc {
	p.next()
	if p.inJob {
		return parseStateCommand
	} else {
		return parseStateStart
	}
}

func parseStateCommand(p *parser) parseStateFunc {
	p.inJob = true
	t := p.next()
	switch t.typ {
	case tokenErrTy:
		p.jobs = append(p.jobs, *p.job)
		return nil
	case tokenPoundTy:
		return parseStateComment
	case tokenNewLineTy:
		return parseStateCommand
	case tokenTabTy, tokenArrowTy:
		return parseStateArg
	case tokenCmdTy:
		p.jobs = append(p.jobs, *p.job)
		p.backup()
		return parseStateStart
	}

	return p.Error("Command args must be indented")
}

// An argument is a list of trees
// Most will be length one and depth 0 (eg. a string, number, variable)
// Others will be list of string/number/var/expression
func parseStateArg(p *parser) parseStateFunc {
	p.arg = []tree{}
	var t = p.next()

	// a single arg may have multiple elements, and is terminated by => or \n
	for ; t.typ != tokenArrowTy && t.typ != tokenNewLineTy; t = p.next() {
		switch t.typ {
		case tokenNumberTy:
			// numbers are easy
			tr := tree{token: t}
			p.arg = append(p.arg, tr)
		case tokenQuoteTy:
			// catch a quote delineated string
			t2 := p.next()
			if t2.typ != tokenStringTy {
				return p.Error(fmt.Sprintf("Invalid token following quote: %s", t2.val))
			}
			q := p.next()
			if q.typ != tokenQuoteTy {
				return p.Error(fmt.Sprintf("Missing ending quote"))
			}

			tr := tree{token: t2}
			p.arg = append(p.arg, tr)
		case tokenStringTy:
			// new variable (string without quotes)
			tr := tree{token: t}
			p.arg = append(p.arg, tr)
		case tokenBlingTy:
			// known variable
			v := p.next()
			if v.typ != tokenStringTy {
				return p.Error(fmt.Sprintf("Invalid variable name: %s", v.val))
			}
			// setting identifier means epm will
			// look it up in symbols table
			tr := tree{
				token:      v,
				identifier: true,
			}
			p.arg = append(p.arg, tr)
		case tokenLeftBracesTy:

		case tokenNewLineTy:
		}
	}

	// add the arg to the job
	p.job.args = append(p.job.args, p.arg)

	if t.typ == tokenArrowTy {
		p.backup()
	}
	return parseStateCommand
}

func parseStateBrace(p *parser) parseStateFunc {
	return nil
}

func parseStateString(p *parser) parseStateFunc {
	s := ""
	for t := p.next(); t.typ == tokenStringTy; t = p.peek() {
		s += t.val + " "
		p.next() // consume the lookahead
	}
	//job := len(p.jobs) - 1
	//p.jobs.args = append(p.jobs.args, s)
	t := p.peek()
	switch t.typ {
	case tokenArrowTy:
		p.next() // consume the arrow
		return parseStateCommand
	case tokenStringTy:
		return parseStateCommand
	case tokenNewLineTy:
		return parseStateCommand
	}

	return nil
}
