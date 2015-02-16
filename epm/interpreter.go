package epm

import (
	"encoding/hex"
	"fmt"
	"github.com/eris-ltd/epm-go/utils"
	"math/big"
	"strconv"
)

// which arg is a "set var"
func setter(cmd string) int {
	switch cmd {
	case "deploy", "modify-deploy":
		return 1
	case "query":
		return 2
	default:
		return -1
	}
}

func (e *EPM) ResolveArgs(cmd string, args [][]*tree) ([]string, error) {
	var stringArgs = []string{}
	for i, a := range args {
		// this is a list of trees. assume for now all length one
		aa := a[0]
		if i == setter(cmd) {
			stringArgs = append(stringArgs, aa.token.val)
			continue
		}
		r, err := e.resolveTree(aa)
		if err != nil {
			return nil, err
		}
		stringArgs = append(stringArgs, r)
	}
	return stringArgs, nil
}

func (e *EPM) resolveTree(tr *tree) (string, error) {
	if len(tr.children) == 0 {
		t := tr.token
		if t.typ == tokenOpTy {
			return "", fmt.Errorf("Operator %s found at leaf", t.val)
		}
		if tr.identifier {
			if v, err := e.VarSub(t.val); err != nil {
				return "", err
			} else {
				t.val = v
			}
		}
		return t.val, nil
	}

	args := []string{}
	for _, a := range tr.children {
		r, err := e.resolveTree(a)
		if err != nil {
			return "", err
		}
		args = append(args, r)
	}
	return performOp(tr.token.val, args)
}

func performOp(op string, args []string) (string, error) {
	// convert args to big ints
	argsB := []*big.Int{}
	for _, a := range args {
		b, err := string2Big(a)
		if err != nil {
			return "", err
		}
		argsB = append(argsB, b)
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
		return "", fmt.Errorf("unknown op: %s", op)
	}

	return "0x" + hex.EncodeToString(z.Bytes()), nil
}

func string2Big(s string) (*big.Int, error) {

	if !utils.IsHex(s) {
		n, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		return big.NewInt(int64(n)), nil

	}
	h := utils.StripHex(s)
	b, err := hex.DecodeString(h)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(b), nil
}
