package gx

import (
	"bufio"
	"fmt"
	"os"
)

func (vm *XGeoVM) DumpConstants() {
	fmt.Println("Constants table:")
	for i, constant := range vm.Constants {
		fmt.Printf("  [%d] = %v\n", i, constant)
	}
}

func (vm *XGeoVM) DumpEnv() {
	fmt.Println("Env:")
	for k, val := range vm.env {
		fmt.Printf("  [%s] = %v\n", k, val)
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
		fmt.Printf("%d : %v\n", i, code)
	}
}

func (vm *XGeoVM) DumpState() {
	fmt.Print("====== VM STATE ======\n\n")
	vm.DumpEnv()
	vm.DumpStack()
	fmt.Print("\n======================\n\n")
}

func (vm *XGeoVM) DumpStep() {
	fmt.Printf("========= @%d =========\n\n", vm.pc)
	vm.DumpEnv()
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
