package epm

import (
	"encoding/hex"
	"fmt"
	"github.com/eris-ltd/epm-go/utils"
	"math/big"
	"strconv"
)

func ResolveArgs(args [][]*tree) []string {
	var stringArgs = []string{}
	for _, a := range args {
		// this is a list of trees. assume for now all length one
		aa := a[0]
		stringArgs = append(stringArgs, resolveTree(aa))
	}
	return stringArgs
}

func resolveTree(tr *tree) string {
	if len(tr.children) == 0 {
		// this must be a value (not an op)
		return tr.token.val
	}

	args := []string{}
	for _, a := range tr.children {
		args = append(args, resolveTree(a))
	}
	r := performOp(tr.token.val, args)
	return r
}

func performOp(op string, args []string) string {
	// convert args to big ints
	argsB := []*big.Int{}
	for _, a := range args {
		argsB = append(argsB, string2Big(a))
	}
	var z *big.Int
	switch op {
	case "+":
		z = new(big.Int).Add(argsB[0], argsB[1])
	case "-":
		z = new(big.Int).Sub(argsB[0], argsB[1])
	case "*":
		z = new(big.Int).Mul(argsB[0], argsB[1])
	case "/":
		z = new(big.Int).Div(argsB[0], argsB[1])
	case "%":
		z = new(big.Int).Mod(argsB[0], argsB[1])
	default:
		fmt.Println("unknown op:", op)
	}

	return "0x" + hex.EncodeToString(z.Bytes())
}

func string2Big(s string) *big.Int {

	if !utils.IsHex(s) {
		n, _ := strconv.Atoi(s)
		return big.NewInt(int64(n))

	}
	h := utils.StripHex(s)
	b, _ := hex.DecodeString(h)
	return new(big.Int).SetBytes(b)
}
