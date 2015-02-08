package lexer

import (
	"strings"
)

// Starting state
func stateStart(l *lexer) stateFunc {
	// check the one character tokens
	t := l.next()
	switch t {
	case "":
		return nil
	case tokenNewLine:
		return stateNewLine
	case tokenTab:
		l.emit(tokenTabTy)
		return stateStart
	case tokenPound:
		l.emit(tokenPoundTy)
		return stateComment
	case tokenColon:
		l.emit(tokenColonTy)
		return stateStart
	case tokenQuote:
		l.emit(tokenQuoteTy)
		return stateQuote
	}
	l.backup()

	remains := l.input[l.pos:]

	// check for tabs (four spaces)
	if strings.HasPrefix(remains, tokenFourSpaces) {
		// if its more than four spaces, ignore it all
		if isSpace(l.peek()) {
			return stateSpace
		}
		return stateFourSpaces
	}

	// skip spaces
	if isSpace(l.peek()) {
		return stateSpace
	}

	// check for left brace
	if strings.HasPrefix(remains, tokenLeftBraces) {
		return stateLeftBraces
	}

	// check for arrow
	if strings.HasPrefix(remains, tokenArrow) {
		return stateArrow
	}

	// check for command
	for _, t := range tokenCmds {
		if strings.HasPrefix(remains, t) {
			l.temp = t
			return stateCmd
		}
	}

	return nil
}

func isSpace(s string) bool {
	return s == " " || s == "\t"
}

func stateNewLine(l *lexer) stateFunc {
	l.emit(tokenNewLineTy)
	l.line += 1
	l.lastNewLine = l.pos
	return stateStart
}

// Scan past spaces
func stateSpace(l *lexer) stateFunc {
	for s := l.next(); isSpace(s); s = l.next() {
	}
	l.backup()
	l.start = l.pos
	return stateStart
}

// At an opening quotes, parse until the closing quote
func stateQuote(l *lexer) stateFunc {
	for s := ""; s != tokenQuote; s = l.next() {
	}
	l.backup()
	l.emit(tokenStringTy)
	l.next()
	l.emit(tokenQuoteTy)
	return stateStart

}

// At a command
func stateCmd(l *lexer) stateFunc {
	l.pos += len(l.temp)
	l.emit(tokenCmdTy)
	return stateStart
}

// In a comment. Scan to new line
func stateComment(l *lexer) stateFunc {
	for r := ""; r != "\n"; r = l.next() {
	}
	l.backup()
	l.emit(tokenStringTy)
	l.next()
	return stateStart
}

// At set of four spaces (alternative to a tab)
func stateFourSpaces(l *lexer) stateFunc {
	l.pos += len(tokenFourSpaces)
	l.emit(tokenTabTy)
	return stateStart
}

// At an arrow (=>)
func stateArrow(l *lexer) stateFunc {
	l.pos += len(tokenArrow)
	l.emit(tokenArrowTy)
	return stateStart
}

// On {{
func stateLeftBraces(l *lexer) stateFunc {
	l.pos += len(tokenLeftBraces)
	l.emit(tokenLeftBracesTy)
	return stateBetweenBraces
}

// On }}
func stateRightBraces(l *lexer) stateFunc {
	l.pos += len(tokenRightBraces)
	l.emit(tokenRightBracesTy)
	return stateStart
}

// Within {{ }}
func stateBetweenBraces(l *lexer) stateFunc {
	if strings.HasPrefix(l.input[l.pos:], tokenRightBraces) {
		return stateRightBraces
	}

	s := l.next()
	if isSpace(s) {
		l.start = l.pos
		return stateBetweenBraces
	} else if strings.Contains(tokenNumbers, s) {
		l.backup()
		return stateNumber
	} else if strings.Contains(tokenOps, s) {
		l.emit(tokenOpTy)
		return stateBetweenBraces
	} else if strings.Contains(tokenChars, s) {
		l.backup()
		return stateString
	}

	return stateErr
}

// a number (decimal or hex)
func stateNumber(l *lexer) stateFunc {
	if l.accept("0") && l.accept("xX") {
		l.acceptRun(tokenHex)
	} else {
		l.acceptRun(tokenNumbers)
	}
	l.emit(tokenNumberTy)
	return stateBetweenBraces
}

// a string
func stateString(l *lexer) stateFunc {
	l.acceptRun(tokenChars)
	l.emit(tokenStringTy)
	return stateBetweenBraces
}

// error!
func stateErr(l *lexer) stateFunc {
	l.emit(tokenErrTy)
	return nil
}
