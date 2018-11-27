package vm

import (
	"fmt"
	"github.com/stationa/xgeo/model"
)

const (
	RegisterCount = 256
)

type XGeoVM struct {
	Constants     []Value
	Code          []*Code
	dumpOnCrash   bool
	debug         bool
	registerCount int
	registers     []Value
	stack         []Value
	pc            int
}

func NewVM(registerCount int) *XGeoVM {
	return &XGeoVM{
		registerCount: registerCount,
		registers:     make([]Value, registerCount),
	}
}

func (vm *XGeoVM) SetDumpOnCrash(doIt bool) {
	vm.dumpOnCrash = doIt
}

func (vm *XGeoVM) SetDebug(debug bool) {
	vm.debug = debug
}

func (vm *XGeoVM) Reset() {
	vm.registers = make([]Value, RegisterCount)
	vm.stack = []Value{}
	vm.pc = 0
}

func (vm *XGeoVM) Run(input interface{}, output chan interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			if vm.dumpOnCrash {
				vm.DumpState()
			}
			panic(r)
		}
	}()
	vm.Reset()
	if vm.debug {
		vm.DumpConstants()
		vm.DumpCode()
	}
	for {
		if vm.debug {
			vm.DumpStep()
		}
		stop, err := vm.step(input, output)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	if vm.debug {
		vm.DumpState()
	}
	return nil
}

func (vm *XGeoVM) step(input interface{}, output chan interface{}) (bool, error) {
	code := vm.Code[vm.pc]
	jmp := 1
	switch code.Op {
	case OpCONST:
		index := code.Args[0]
		if index >= len(vm.Constants) {
			panic("Invalid constant access")
		}
		vm.push(vm.Constants[index])
	case OpLOADG:
		vm.push(&Raw{input})
	case OpDEREF:
		prop := vm.pop().(*Str).NativeValue
		ctx := vm.pop()
		res, _ := vm.deref(ctx, prop)
		vm.push(res)
	case OpMUT:
		val := vm.pop()
		prop := vm.pop().(*Str).NativeValue
		ctx := vm.pop()
		vm.mut(ctx, prop, val)
	case OpLOAD:
		register := code.Args[0]
		val := vm.registers[register]
		vm.push(val)
	case OpSTORE:
		val := vm.pop()
		register := code.Args[0]
		vm.registers[register] = val
	case OpADD:
		right := vm.pop()
		left := vm.pop()
		res, _ := left.Add(right)
		vm.push(res)
	case OpSUB:
		right := vm.pop()
		left := vm.pop()
		res, _ := left.Sub(right)
		vm.push(res)
	case OpMUL:
		right := vm.pop()
		left := vm.pop()
		res, _ := left.Mul(right)
		vm.push(res)
	case OpDIV:
		right := vm.pop()
		left := vm.pop()
		res, _ := left.Div(right)
		vm.push(res)
	case OpEQ:
		right := vm.pop()
		left := vm.pop()
		res, _ := left.Eq(right)
		vm.push(res)
	case OpNEQ:
		right := vm.pop()
		left := vm.pop()
		res, _ := left.Neq(right)
		vm.push(res)
	case OpGT:
		right := vm.pop()
		left := vm.pop()
		res, _ := left.Gt(right)
		vm.push(res)
	case OpGTE:
		right := vm.pop()
		left := vm.pop()
		res, _ := left.Gte(right)
		vm.push(res)
	case OpLT:
		right := vm.pop()
		left := vm.pop()
		res, _ := left.Lt(right)
		vm.push(res)
	case OpLTE:
		right := vm.pop()
		left := vm.pop()
		res, _ := left.Lte(right)
		vm.push(res)
	case OpJMP:
		ip := code.Args[0]
		jmp = ip - vm.pc
	case OpJMPT:
		shouldJump := vm.pop().(*Bool).NativeValue
		if shouldJump {
			ip := code.Args[0]
			jmp = ip - vm.pc
		}
	case OpJMPF:
		shouldJump := !vm.pop().(*Bool).NativeValue
		if shouldJump {
			ip := code.Args[0]
			jmp = ip - vm.pc
		}
	case OpEMIT:
		val := vm.pop()
		output <- val.Raw()
	default:
		panic(fmt.Errorf("Op-code not yet implemented!: %s", code.Op))
	}
	vm.pc += jmp
	if vm.pc >= len(vm.Code) {
		// Stop if the program counter hits the end of the code
		return true, nil
	}
	return false, nil
}

func (vm *XGeoVM) push(val Value) {
	vm.stack = append(vm.stack, val)
}

func (vm *XGeoVM) pop() Value {
	if len(vm.stack) == 0 {
		panic("Stack empty!")
	}
	l := len(vm.stack) - 1
	val := vm.stack[l]
	vm.stack = vm.stack[:l]
	return val
}

func (vm *XGeoVM) deref(val Value, prop string) (Value, error) {
	raw := val.Raw()
	switch raw := raw.(type) {
	case *model.Feature:
		return vm.access(raw, prop), nil
	case map[string]string:
		return &Str{raw[prop]}, nil
	default:
		panic("Unsupported dereference")
	}
}

func (vm *XGeoVM) mut(ctx Value, prop string, val Value) error {
	raw := ctx.Raw()
	switch raw := raw.(type) {
	case *model.Feature:
		return vm.mutate(raw, prop, val)
	case map[string]string:
		raw[prop] = val.Raw().(string)
	default:
		panic("Unsupported dereference")
	}
	return nil
}
