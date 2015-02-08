package lexer

import (
	"fmt"
)

func (t token) String() string {
	s := fmt.Sprintf("Line %d, Col %d \t %-6s \t", t.loc.line, t.loc.col, t.typ.String())
	switch t.typ {
	case tokenEOFTy:
		return s + "EOF"
	case tokenErrTy:
		return s + t.val
	}
	/*if len(t.val) > 10 {
		return fmt.Sprintf("%.10q...", t.val)
	}*/
	return s + fmt.Sprintf("%q", t.val)
}

// token types
type tokenType int

func (t tokenType) String() string {
	switch t {
	case tokenErrTy:
		return "[Error]"
	case tokenEOFTy:
		return "[EOF]"
	case tokenCmdTy:
		return "[Cmd]"
	case tokenLeftBracesTy:
		return "[LeftBraces]"
	case tokenRightBracesTy:
		return "[RightBraces]"
	case tokenNumberTy:
		return "[Number]"
	case tokenArrowTy:
		return "[Arrow]"
	case tokenColonTy:
		return "[Colon]"
	case tokenStringTy:
		return "[String]"
	case tokenQuoteTy:
		return "[Quote]"
	case tokenTabTy:
		return "[Tab]"
	case tokenNewLineTy:
		return "[NewLine]"
	case tokenPoundTy:
		return "[Pound]"
	case tokenOpTy:
		return "[Op]"
	case tokenSpaceTy:
		return "[Space]"
	}
	return "[Unknown]"
}

// token types
const (
	tokenErrTy         tokenType = iota // error
	tokenEOFTy                          // end of file
	tokenCmdTy                          // command (deploy, modify-deploy, transact, etc.)
	tokenLeftBracesTy                   // {{
	tokenRightBracesTy                  // }}
	tokenNumberTy                       // int or hex
	tokenArrowTy                        // =>
	tokenColonTy                        // :
	tokenStringTy                       // variable name, contents of quotes, comments
	tokenQuoteTy                        // "
	tokenTabTy                          // \t or four spaces
	tokenNewLineTy                      // \n
	tokenPoundTy                        // #
	tokenOpTy                           // math ops (+, -, *, /, %)
	tokenSpaceTy                        // debugging
)

// tokens and special chars
var (
	tokenCmds        = []string{"deploy", "modify-deploy", "transact", "endow", "deploy"}
	tokenLeftBraces  = "{{"
	tokenRightBraces = "}}"
	tokenArrow       = "=>"
	tokenFourSpaces  = "    "
	tokenQuote       = "\""
	tokenColon       = ":"
	tokenTab         = "\t"
	tokenNewLine     = "\n"
	tokenPound       = "#"
	tokenSpace       = " "

	tokenNumbers = "0123456789"
	tokenHex     = "0123456789abcdef"
	tokenOps     = "+-*/%"
	tokenChars   = "abcdefghijklmnopqrstuvwqyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
)
