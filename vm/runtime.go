package vm

import (
	"bufio"
	"fmt"
	"github.com/stationa/xgeo/model"
	"os"
)

const (
	RegisterCount = 256
)

type XGeoVM struct {
	Constants     []Value
	Code          []*Code
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

func (vm *XGeoVM) SetDebug(debug bool) {
	vm.debug = debug
}

func (vm *XGeoVM) DumpConstants() {
	fmt.Println("Constants table:")
	for i, constant := range vm.Constants {
		fmt.Printf("  [%d] = %v\n", i, constant)
	}
}

func (vm *XGeoVM) DumpRegisters() {
	fmt.Println("Registers:")
	for i, val := range vm.registers[:vm.registerCount] {
		fmt.Printf("  [%d] = %v\n", i, val)
	}
}

func (vm *XGeoVM) DumpStack() {
	fmt.Println("Stack:")
	for i, _ := range vm.stack {
		fmt.Print("  ")
		if i == 0 {
			fmt.Print("→ ")
		} else {
			fmt.Print("| ")
		}
		v := vm.stack[len(vm.stack)-i-1]
		fmt.Printf("%s\n", v)
	}
	if len(vm.stack) == 0 {
		fmt.Println("  <empty>")
	}
}

func (vm *XGeoVM) DumpCode() {
	fmt.Println("Code listing:")
	for i, code := range vm.Code {
		fmt.Print("  ")
		if i == vm.pc {
			fmt.Print("*")
		} else {
			fmt.Print("@")
		}
		fmt.Printf("  %d : %v\n", i, code)
	}
}

func (vm *XGeoVM) DumpState() {
	fmt.Print("====== VM STATE ======\n\n")
	vm.DumpRegisters()
	vm.DumpStack()
	fmt.Print("\n======================\n\n")
}

func (vm *XGeoVM) DumpStep() {
	fmt.Printf("========= @%d =========\n\n", vm.pc)
	vm.DumpRegisters()
	vm.DumpStack()
	fmt.Println("Step:")
	for i, code := range vm.Code {
		if i == vm.pc {
			fmt.Printf("  *%d : %v\n", i, code)
		}
	}
	fmt.Print("\n======================\n\n")
	fmt.Print("Press <ENTER> to continue...\n")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func (vm *XGeoVM) Reset() {
	vm.registers = make([]Value, RegisterCount)
	vm.stack = []Value{}
	vm.pc = 0
}

func (vm *XGeoVM) Run(feature *model.Feature) *model.Feature {
	defer func() {
		if r := recover(); r != nil {
			vm.DumpState()
			panic(r)
		}
	}()
	vm.Reset()
	if vm.debug {
		vm.DumpConstants()
		vm.DumpCode()
	}
	for vm.pc < len(vm.Code) {
		if vm.debug {
			vm.DumpStep()
		}
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
			vm.push(&Raw{feature})
		case OpDEREF:
			prop := vm.pop().(*Str).NativeValue
			val := vm.pop()
			vm.deref(val, prop)
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
		case OpCOND:
			shouldJump := !vm.pop().(*Bool).NativeValue
			if shouldJump {
				ip := code.Args[0]
				jmp = ip - vm.pc
			}
		case OpEMIT:
			val := vm.pop()
			vm.emit(val)
		default:
			panic(fmt.Errorf("Op-code not yet implemented!: %s", code.Op))
		}
		vm.pc += jmp
	}
	if vm.debug {
		vm.DumpState()
	}
	return feature
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
		return vm.Access(raw, prop), nil
	case map[string]string:
		return &Str{raw[prop]}, nil
	default:
		panic("Unsupported dereference")
	}
}

func (vm *XGeoVM) emit(val Value) {
	fmt.Printf("Emitting value: %v\n", val)
}
