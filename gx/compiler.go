package gx

import (
	"fmt"
	"strings"
)

func (c *XGeoCompiler) Prepare() {
	c.Init()
	c.refs = make(map[string]int)
}

func (c *XGeoCompiler) StartCall(funcName string) {
	builtin := LookupBuiltin(funcName)
	call := &Code{
		OpCALL,
		[]int{builtin, 0},
	}
	c.callStack = append(c.callStack, call)
}

func (c *XGeoCompiler) AddCallArg() {
	l := len(c.callStack) - 1
	lastCall := c.callStack[l]
	lastCall.Args[1] += 1
}

func (c *XGeoCompiler) EmitCall() {
	l := len(c.callStack) - 1
	lastCall := c.callStack[l]
	c.EmitCode(lastCall.Op, lastCall.Args...)
	c.callStack = c.callStack[:l]
}

func (c *XGeoCompiler) EmitJumpIfFalse() {
	code := c.EmitCode(OpJMPF, -1)
	c.jumpStack = append(c.jumpStack, code)
}

func (c *XGeoCompiler) EmitJumpIfTrue() {
	code := c.EmitCode(OpJMPT, -1)
	c.jumpStack = append(c.jumpStack, code)
}

func (c *XGeoCompiler) EmitJump() {
	code := c.EmitCode(OpJMP, -1)
	c.jumpStack = append(c.jumpStack, code)
}

func (c *XGeoCompiler) EmitJumpAfter() {
	after := len(c.code) + 2
	c.EmitCode(OpJMP, after)
}

func (c *XGeoCompiler) SetJump() {
	l := len(c.jumpStack) - 1
	lastJump := c.jumpStack[l]
	lastJump.Args[0] = len(c.code)
	c.jumpStack = c.jumpStack[:l]
}

func (c *XGeoCompiler) SetJumpAfter() {
	l := len(c.jumpStack) - 1
	lastJump := c.jumpStack[l]
	lastJump.Args[0] = len(c.code) + 1
	c.jumpStack = c.jumpStack[:l]
}

func (c *XGeoCompiler) EmitConstant(val Value) {
	index := len(c.constants)
	for i, constant := range c.constants {
		// Reuse duplicate constants
		if constant.Raw() == val.Raw() {
			c.EmitCode(OpCONST, i)
			return
		}
	}
	c.EmitCode(OpCONST, index)
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
	c.EmitCode(OpLOADG)
	pathParts := strings.Split(path, ".")
	l := len(pathParts) - 1
	for _, prop := range pathParts[:l] {
		c.EmitConstant(&Str{prop})
		c.EmitCode(OpDEREF)
	}
	c.EmitConstant(&Str{pathParts[l]})
}

func (c *XGeoCompiler) EmitStore() {
	register := c.registerCount - 1
	c.EmitCode(OpSTORE, register)
}

func (c *XGeoCompiler) EmitLoad(ref string) {
	if strings.HasPrefix(ref, "@") {
		path := ref[1:]
		c.EmitCode(OpLOADG)
		if path != "" {
			for _, prop := range strings.Split(path, ".") {
				c.EmitConstant(&Str{prop})
				c.EmitCode(OpDEREF)
			}
		}
	} else {
		register, refExists := c.refs[ref]
		if !refExists {
			panic(fmt.Errorf("Undefined reference: %s", ref))
		}
		c.EmitCode(OpLOAD, register)
	}
}

func (c *XGeoCompiler) EmitCode(op Op, args ...int) *Code {
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
