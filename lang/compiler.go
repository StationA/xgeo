package lang

import (
	"fmt"
	"github.com/stationa/xgeo/vm"
	"strings"
)

func (c *XGeoCompiler) Prepare() {
	c.Init()
	c.refs = make(map[string]int)
}

func (c *XGeoCompiler) AddCondJump() {
	code := c.AddCode(vm.OpCOND, -1)
	c.jumpStack = append(c.jumpStack, code)
}

func (c *XGeoCompiler) SetJump() {
	l := len(c.jumpStack) - 1
	lastJump := c.jumpStack[l]
	lastJump.Args[0] = len(c.code)
	c.jumpStack = c.jumpStack[:l]
}

func (c *XGeoCompiler) AddConstant(val vm.Value) {
	index := len(c.constants)
	for i, constant := range c.constants {
		// Reuse duplicate constants
		if constant == val {
			c.AddCode(vm.OpCONST, i)
			return
		}
	}
	c.AddCode(vm.OpCONST, index)
	c.constants = append(c.constants, val)
}

func (c *XGeoCompiler) AllocateRef(ref string) {
	if strings.HasPrefix(ref, "@") {
		panic("Not implemented!")
	} else {
		register, refExists := c.refs[ref]
		if !refExists {
			register = c.registerCount
			c.refs[ref] = register
			c.registerCount += 1
		}
	}
}

func (c *XGeoCompiler) AddStore() {
	register := c.registerCount - 1
	c.AddCode(vm.OpSTORE, register)
}

func (c *XGeoCompiler) AddLoad(ref string) {
	if strings.HasPrefix(ref, "@") {
		c.AddCode(vm.OpLOADG)
		for _, prop := range strings.Split(ref[1:], ".") {
			c.AddConstant(&vm.Str{prop})
			c.AddCode(vm.OpDEREF)
		}
	} else {
		register, refExists := c.refs[ref]
		if !refExists {
			panic(fmt.Errorf("Undefined reference: %s", ref))
		}
		c.AddCode(vm.OpLOAD, register)
	}
}

func (c *XGeoCompiler) AddCode(op vm.Op, args ...int) *vm.Code {
	code := &vm.Code{op, args}
	c.code = append(c.code, code)
	return code
}

func (c *XGeoCompiler) Compile() error {
	c.Execute()
	// TODO: Do any static type checks?
	return nil
}

func (c *XGeoCompiler) InitVM() *vm.XGeoVM {
	v := vm.NewVM(c.registerCount)
	v.Constants = c.constants[:]
	v.Code = c.code[:]
	return v
}
