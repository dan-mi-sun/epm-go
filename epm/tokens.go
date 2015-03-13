package epm

import (
	"fmt"
)

func (t token) String() string {
	s := fmt.Sprintf("Line %-2d, Col %-2d \t %-6s \t", t.loc.line, t.loc.col, t.typ.String())
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
	case tokenBlingTy:
		return "[Bling]"
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
	case tokenLeftDiffTy:
		return "[LeftDiff]"
	case tokenRightDiffTy:
		return "[RightDiff]"
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
	tokenLeftBraceTy                    // (
	tokenRightBraceTy                   // )
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
	tokenBlingTy                        // $
	tokenLeftDiffTy                     // !{
	tokenRightDiffTy                    // !}
	tokenUnderscoreTy                   // _
)

// min args per command
var CommandArgs = map[string]int{
	"deploy":        2,
	"modify-deploy": 4,
	"transact":      2,
	"query":         3,
	"log":           2,
	"set":           2,
	"endow":         2,
	"test":          1,
	"epm":           1,
}

// tokens and special chars
var (
	tokenCmds        = CommandArgs
	tokenLeftBraces  = "{{"
	tokenRightBraces = "}}"
	tokenLeftBrace   = "("
	tokenRightBrace  = ")"
	tokenLeftDiff    = "!{"
	tokenRightDiff   = "!}"
	tokenArrow       = "=>"
	tokenFourSpaces  = "    "
	tokenQuote       = "\""
	tokenColon       = ":"
	tokenTab         = "\t"
	tokenNewLine     = "\n"
	tokenPound       = "#"
	tokenSpace       = " "
	tokenBling       = "$"
	tokenUnderscore  = "_"

	tokenNumbers = "0123456789"
	tokenHex     = "0123456789abcdefABCDEF"
	tokenOps     = "+-*/%"
	tokenChars   = "abcdefghijklmnopqrstuvwqyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-/_."
)
