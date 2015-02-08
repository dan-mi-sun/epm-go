package lexer

// This is my first lexer
// It is heavily inspired by Rob Pike's "Lexical Scanning in Go"
// https://www.youtube.com/watch?v=HxaD_trXwRE
// In fact, I got so excited after watching that,
// I immediately started writing this

import (
	"strings"
)

// when in a state, do an action, which brings us to another state
// so a state is really a state+action ie. stateFunc that returns the next
// stateFunc. Done when it returns nil
type stateFunc func(*lexer) stateFunc

// the lexer object
type lexer struct {
	input  string // input string to lex
	length int    // length of the input string
	pos    int    // current pos
	start  int    // start of current token

	line        int // current line number
	lastNewLine int // pos of last new line

	tokens chan token // channel to emit tokens over

	temp string // a place to hold eg. commands
}

// a token
type token struct {
	typ tokenType
	val string

	loc location
}

// location for error reporting
type location struct {
	line int
	col  int
}

// Lex the input, returning the lexer
// Tokens can be fetched off the channel
func Lex(input string) *lexer {
	l := &lexer{
		input:  input,
		length: len(input),
		pos:    0,
		tokens: make(chan token),
	}
	go l.run()
	return l
}

// Return the tokens channel
func (l *lexer) Chan() chan token {
	return l.tokens
}

// Run the lexer
// This is the most beautiful function in the world
func (l *lexer) run() {
	for state := stateStart; state != nil; state = state(l) {
		// :D
	}
	close(l.tokens)
}

// Return next character in the string
// To hell with utf8 :p
func (l *lexer) next() string {
	if l.pos >= l.length {
		return ""
	}
	b := l.input[l.pos : l.pos+1]
	l.pos += 1
	return b
}

// backup a step
func (l *lexer) backup() {
	l.pos -= 1
}

// peek ahead a character without consuming
func (l *lexer) peek() string {
	s := l.next()
	l.backup()
	return s
}

// consume a token and push out on the channel
func (l *lexer) emit(ty tokenType) {
	l.tokens <- token{
		typ: ty,
		val: l.input[l.start:l.pos],
		loc: location{
			line: l.line,
			col:  l.pos - l.lastNewLine,
		},
	}
	l.start = l.pos
}

func (l *lexer) accept(options string) bool {
	if strings.Contains(options, l.next()) {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(options string) {
	for s := l.next(); strings.Contains(options, s); s = l.next() {
	}
	l.backup()
}
