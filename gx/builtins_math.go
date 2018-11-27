package gx

import (
	"errors"
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
				absFloat := math.Abs(float.NativeValue)
				return &Float{absFloat}, nil
			},
		},
	},
}

var Sqrt = &Builtin{
	Name: "sqrt",
	NativeCalls: []*NativeCall{
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Int{},
				},
				Returns: &Float{},
			},
			Call: func(args ...Value) (Value, error) {
				val := args[0].(*Int)
				sqrt := math.Sqrt(float64(val.NativeValue))
				return &Float{sqrt}, nil
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
				val := args[0].(*Float)
				sqrt := math.Sqrt(val.NativeValue)
				return &Float{sqrt}, nil
			},
		},
	},
}

var Round = &Builtin{
	Name: "round",
	NativeCalls: []*NativeCall{
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Float{},
				},
				Returns: &Int{},
			},
			Call: func(args ...Value) (Value, error) {
				val := args[0].(*Float)
				rounded := math.Round(float64(val.NativeValue))
				return &Int{int(rounded)}, nil
			},
		},
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Float{},
					&Int{},
				},
				Returns: &Float{},
			},
			Call: func(args ...Value) (Value, error) {
				val := args[0].(*Float)
				decimalPlaces := args[1].(*Int)
				if decimalPlaces.NativeValue >= 0 {
					exponent := float64(decimalPlaces.NativeValue)
					baseTen := math.Pow(10.0, exponent)
					roundedInflated := math.Round(float64(val.NativeValue) * baseTen)
					rounded := roundedInflated / baseTen
					return &Float{rounded}, nil
				}
				return nil, errors.New(fmt.Sprintf("Invalid number of decimal places: %v", decimalPlaces))
			},
		},
	},
}
