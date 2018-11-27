package gx

import (
	"strings"
)

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

var Substr = &Builtin{
	Name: "substr",
	NativeCalls: []*NativeCall{
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Str{},
					&Int{},
				},
				Returns: &Str{},
			},
			Call: func(args ...Value) (Value, error) {
				str := args[0].(*Str)
				start := args[1].(*Int)
				return &Str{str.NativeValue[start.NativeValue:]}, nil
			},
		},
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Str{},
					&Int{},
					&Int{},
				},
				Returns: &Str{},
			},
			Call: func(args ...Value) (Value, error) {
				str := args[0].(*Str)
				start := args[1].(*Int)
				end := args[2].(*Int)
				substr := str.NativeValue[start.NativeValue:end.NativeValue]
				return &Str{substr}, nil
			},
		},
	},
}

var Replace = &Builtin{
	Name: "replace",
	NativeCalls: []*NativeCall{
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Str{},
					&Str{},
					&Str{},
				},
				Returns: &Str{},
			},
			Call: func(args ...Value) (Value, error) {
				str := args[0].(*Str)
				oldStr := args[1].(*Str)
				newStr := args[2].(*Str)
				replaced := strings.Replace(str.NativeValue, oldStr.NativeValue, newStr.NativeValue, -1)
				return &Str{replaced}, nil
			},
		},
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Str{},
					&Str{},
					&Str{},
					&Int{},
				},
				Returns: &Str{},
			},
			Call: func(args ...Value) (Value, error) {
				str := args[0].(*Str)
				oldStr := args[1].(*Str)
				newStr := args[2].(*Str)
				numReplacements := args[3].(*Int)
				replaced := strings.Replace(str.NativeValue, oldStr.NativeValue, newStr.NativeValue, numReplacements.NativeValue)
				return &Str{replaced}, nil
			},
		},
	},
}
