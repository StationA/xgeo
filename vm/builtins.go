package vm

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

func typeError(val Value, funcSig string) error {
	return fmt.Errorf("Invalid type %T for function %s", val, funcSig)
}

var Lower = &Builtin{
	Name: "lower",
	NativeCalls: []*NativeCall{
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Str{},
				},
				Returns: &Str{},
			},
			Call: func(args ...Value) (Value, error) {
				str := args[0].(*Str)
				s := strings.ToLower(str.NativeValue)
				return &Str{s}, nil
			},
		},
	},
}

var Upper = &Builtin{
	Name: "upper",
	NativeCalls: []*NativeCall{
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Str{},
				},
				Returns: &Str{},
			},
			Call: func(args ...Value) (Value, error) {
				str := args[0].(*Str)
				s := strings.ToUpper(str.NativeValue)
				return &Str{s}, nil
			},
		},
	},
}

var Strip = &Builtin{
	Name: "strip",
	NativeCalls: []*NativeCall{
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Str{},
				},
				Returns: &Str{},
			},
			Call: func(args ...Value) (Value, error) {
				str := args[0].(*Str)
				s := strings.TrimSpace(str.NativeValue)
				return &Str{s}, nil
			},
		},
	},
}

var CastBool = &Builtin{
	Name: "bool",
	NativeCalls: []*NativeCall{
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Bool{},
				},
				Returns: &Bool{},
			},
			Call: func(args ...Value) (Value, error) {
				return args[0], nil
			},
		},
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Str{},
				},
				Returns: &Bool{},
			},
			Call: func(args ...Value) (Value, error) {
				str := args[0].(*Str)
				return ParseBool(str.NativeValue), nil
			},
		},
	},
}

var CastInt = &Builtin{
	Name: "int",
	NativeCalls: []*NativeCall{
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Int{},
				},
				Returns: &Int{},
			},
			Call: func(args ...Value) (Value, error) {
				return args[0], nil
			},
		},
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Float{},
				},
				Returns: &Int{},
			},
			Call: func(args ...Value) (Value, error) {
				f := args[0].(*Float)
				return &Int{int(f.NativeValue)}, nil
			},
		},
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Str{},
				},
				Returns: &Int{},
			},
			Call: func(args ...Value) (Value, error) {
				str := args[0].(*Str)
				return ParseInt(str.NativeValue), nil
			},
		},
	},
}

var CastFloat = &Builtin{
	Name: "float",
	NativeCalls: []*NativeCall{
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Int{},
				},
				Returns: &Float{},
			},
			Call: func(args ...Value) (Value, error) {
				i := args[0].(*Int)
				return &Float{float64(i.NativeValue)}, nil
			},
		},
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Float{},
				},
				Returns: &Float{},
			},
			Call: func(args ...Value) (Value, error) {
				return args[0], nil
			},
		},
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Str{},
				},
				Returns: &Float{},
			},
			Call: func(args ...Value) (Value, error) {
				str := args[0].(*Str)
				return ParseFloat(str.NativeValue), nil
			},
		},
	},
}

var CastStr = &Builtin{
	Name: "str",
	NativeCalls: []*NativeCall{
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Bool{},
				},
				Returns: &Str{},
			},
			Call: func(args ...Value) (Value, error) {
				b := args[0].(*Bool)
				if b.NativeValue {
					return &Str{"true"}, nil
				}
				return &Str{"false"}, nil
			},
		},
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Int{},
				},
				Returns: &Str{},
			},
			Call: func(args ...Value) (Value, error) {
				i := args[0].(*Int)
				return &Str{fmt.Sprintf("%d", i.NativeValue)}, nil
			},
		},
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Float{},
				},
				Returns: &Str{},
			},
			Call: func(args ...Value) (Value, error) {
				f := args[0].(*Float)
				return &Str{fmt.Sprintf("%f", f.NativeValue)}, nil
			},
		},
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Str{},
				},
				Returns: &Str{},
			},
			Call: func(args ...Value) (Value, error) {
				return args[0], nil
			},
		},
	},
}

var Builtins = []*Builtin{
	Lower,
	Upper,
	Strip,
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
