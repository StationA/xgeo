package vm

import (
	"fmt"
	"strings"
)

func (c *XGeoCompiler) Prepare() {
	c.Init()
	c.refs = make(map[string]int)
}

func (c *XGeoCompiler) AddCondJump() {
	code := c.AddCode(OpCOND, -1)
	c.jumpStack = append(c.jumpStack, code)
}

func (c *XGeoCompiler) SetJump() {
	l := len(c.jumpStack) - 1
	lastJump := c.jumpStack[l]
	lastJump.Args[0] = len(c.code)
	c.jumpStack = c.jumpStack[:l]
}

func (c *XGeoCompiler) AddConstant(val Value) {
	index := len(c.constants)
	for i, constant := range c.constants {
		// Reuse duplicate constants
		if constant == val {
			c.AddCode(OpCONST, i)
			return
		}
	}
	c.AddCode(OpCONST, index)
	c.constants = append(c.constants, val)
}

func (c *XGeoCompiler) AllocateRef(ref string) {
	register, refExists := c.refs[ref]
	if !refExists {
		register = c.registerCount
		c.refs[ref] = register
		c.registerCount += 1
	}
}

func (c *XGeoCompiler) PrepareMutate(ref string) {
	// Trim off the leading "@"
	path := ref[1:]
	if path == "" {
		panic("Cannot mutate the entire input")
	}
	c.AddCode(OpLOADG)
	pathParts := strings.Split(path, ".")
	l := len(pathParts) - 1
	for _, prop := range pathParts[:l] {
		c.AddConstant(&Str{prop})
		c.AddCode(OpDEREF)
	}
	c.AddConstant(&Str{pathParts[l]})
}

func (c *XGeoCompiler) AddStore() {
	register := c.registerCount - 1
	c.AddCode(OpSTORE, register)
}

func (c *XGeoCompiler) AddLoad(ref string) {
	if strings.HasPrefix(ref, "@") {
		path := ref[1:]
		c.AddCode(OpLOADG)
		if path != "" {
			for _, prop := range strings.Split(path, ".") {
				c.AddConstant(&Str{prop})
				c.AddCode(OpDEREF)
			}
		}
	} else {
		register, refExists := c.refs[ref]
		if !refExists {
			panic(fmt.Errorf("Undefined reference: %s", ref))
		}
		c.AddCode(OpLOAD, register)
	}
}

func (c *XGeoCompiler) AddCode(op Op, args ...int) *Code {
	code := &Code{op, args}
	c.code = append(c.code, code)
	return code
}

func (c *XGeoCompiler) Compile() error {
	c.Execute()
	// TODO: Do any static type checks?
	return nil
}

func (c *XGeoCompiler) InitVM() *XGeoVM {
	v := NewVM(c.registerCount)
	v.Constants = c.constants[:]
	v.Code = c.code[:]
	return v
}
