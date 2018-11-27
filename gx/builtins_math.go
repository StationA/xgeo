package gx

import (
	"fmt"
	"math"
)

var Abs = &Builtin{
	Name: "abs",
	NativeCalls: []*NativeCall{
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Int{},
				},
				Returns: &Int{},
			},
			Call: func(args ...Value) (Value, error) {
				integer := args[0].(*Int)
				fmt.Println("Treating as Int")
				if integer.NativeValue >= 0 {
					return integer, nil
				}
				absInteger := -1 * integer.NativeValue
				return &Int{absInteger}, nil
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
				float := args[0].(*Float)
				fmt.Println("Treating as Float")
				absFloat := math.Abs(float.NativeValue)
				return &Float{absFloat}, nil
			},
		},
	},
}
