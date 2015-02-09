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

	lookahead token // hold the look ehad
	peekCount int   // 1 if we've peeked

	inJob bool // are we in a job

	tree *tree // current tree
	job  *job  //current job
}

type job struct {
	cmd  string
	args [][]tree
}

type tree struct {
	token    *token
	parent   *token
	children []*token
}

func Parse(input string) *parser {
	l := Lex(input)
	p := &parser{
		l:    l,
		jobs: []job{},
		tree: new(tree),
		job:  new(job),
	}
	go p.run()
	return p
}

func (p *parser) next() token {
	if p.peekCount == 1 {
		p.peekCount = 0
		return p.lookahead

	}
	p.last = <-p.l.Chan()
	return p.last
}

func (p *parser) peek() token {
	if p.peekCount == 1 {
		return p.lookahead
	}
	p.lookahead = p.next()
	p.peekCount = 1
	return p.lookahead
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
		t = p.next()
		if t.typ != tokenColonTy {
			return p.Error("Commands must be followed by a colon")
		}
		j := job{
			cmd:  t.val,
			args: [][]tree{},
		}
		p.jobs = append(p.jobs, j)
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
	t := p.next()
	switch t.typ {
	case tokenErrTy:
		return nil
	case tokenNewLineTy:
		t = p.next()
		if t.typ != tokenTabTy {
			return p.Error("Command args must be indented")
		}
		return parseStateCommand
	// a command is a list of args
	// an arg is a list of trees
	// usually just length one
	// trees begin as variables, strings, or braces
	case tokenStringTy:
		return parseStateString
	case tokenLeftBracesTy:
		return parseStateBrace
	}

	return nil
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
