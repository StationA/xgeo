package gx

import (
	"fmt"
)

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
