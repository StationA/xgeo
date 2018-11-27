package gx

import (
	"fmt"
	"strings"
)

type Signature struct {
	Accepts []Value
	Returns Value
}

func (s *Signature) Matches(args ...Value) bool {
	if len(s.Accepts) != len(args) {
		return false
	}
	for i, argType := range s.Accepts {
		if argType.Type() != args[i].Type() {
			return false
		}
	}
	return true
}

func (s *Signature) Display(funcName string) string {
	argString := ""
	returnString := "?"
	if s.Accepts != nil && len(s.Accepts) > 0 {
		var argTypes []string
		for _, argType := range s.Accepts {
			argTypes = append(argTypes, argType.Type())
		}
		argString = strings.Join(argTypes, ",")
	}
	if s.Returns != nil {
		returnString = s.Returns.Type()
	}
	return fmt.Sprintf("%s(%s) -> %s", funcName, argString, returnString)
}

type NativeCall struct {
	Signature *Signature
	Call      func(args ...Value) (Value, error)
}

type Builtin struct {
	Name        string
	NativeCalls []*NativeCall
}

func (b *Builtin) Call(args ...Value) (Value, error) {
	for _, call := range b.NativeCalls {
		if call.Signature.Matches(args...) {
			return call.Call(args...)
		}
	}
	expectedSignature := &Signature{Accepts: args}
	expectedCall := expectedSignature.Display(b.Name)
	return nil, fmt.Errorf("No matching function call: %s", expectedCall)
}

var Builtins = []*Builtin{
	// String builtins
	Lower,
	Upper,
	Strip,
	Substr,
	Replace,
	// Math builtins
	Abs,
	Round,
	Sqrt,
	// Type builtins
	CastBool,
	CastInt,
	CastFloat,
	CastStr,
}

func LookupBuiltin(funcName string) int {
	for i, builtin := range Builtins {
		if builtin.Name == funcName {
			return i
		}
	}
	return -1
}
