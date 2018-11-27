package gx

import (
	"fmt"
	"strings"
)

type Op byte

const (
	OpJMP Op = iota
	OpJMPT
	OpJMPF
	OpLABEL
	OpEMIT
	OpSTORE
	OpLOAD
	OpLOADG
	OpDEREF
	OpMUT
	OpCONST
	OpEQ
	OpNEQ
	OpLT
	OpLTE
	OpGT
	OpGTE
	OpCALL
	OpADD
	OpSUB
	OpMUL
	OpDIV
	OpEXP
)

type Code struct {
	Op   Op
	Args []int
}

func (c *Code) String() string {
	if len(c.Args) == 0 {
		return c.Op.String()
	}
	var sargs []string
	for _, arg := range c.Args {
		sargs = append(sargs, fmt.Sprintf("%d", arg))
	}
	return fmt.Sprintf("%s %s", c.Op, strings.Join(sargs, ","))
}
