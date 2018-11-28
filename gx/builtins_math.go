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
					&Int{},
				},
				Returns: &Int{},
			},
			Call: func(args ...Value) (Value, error) {
				val := args[0].(*Int)
				return &Int{val.NativeValue}, nil
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

var Ceil = &Builtin{
	Name: "ceil",
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
				rounded := math.Ceil(float64(val.NativeValue))
				return &Int{int(rounded)}, nil
			},
		},
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Int{},
				},
				Returns: &Int{},
			},
			Call: func(args ...Value) (Value, error) {
				val := args[0].(*Int)
				return &Int{val.NativeValue}, nil
			},
		},
	},
}

var Floor = &Builtin{
	Name: "floor",
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
				rounded := math.Floor(float64(val.NativeValue))
				return &Int{int(rounded)}, nil
			},
		},
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Int{},
				},
				Returns: &Int{},
			},
			Call: func(args ...Value) (Value, error) {
				val := args[0].(*Int)
				return &Int{val.NativeValue}, nil
			},
		},
	},
}

var M2Ft = &Builtin {
	Name: "m2ft",
	NativeCalls: []*NativeCall{
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Float{},
				},
				Returns: &Float{},
			},
			Call: func(args ...Value) (Value, error) {
				meters := args[0].(*Float)
				feet := meters.NativeValue * 3.2808399
				return &Float{feet}, nil
			},
		},
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Int{},
				},
				Returns: &Float{},
			},
			Call: func(args ...Value) (Value, error) {
				meters := args[0].(*Int)
				feet := float64(meters.NativeValue ) * 3.2808399
				return &Float{feet}, nil
			},
		},
	},
}

var Ft2M = &Builtin {
	Name: "ft2m",
	NativeCalls: []*NativeCall{
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Float{},
				},
				Returns: &Float{},
			},
			Call: func(args ...Value) (Value, error) {
				meters := args[0].(*Float)
				feet := meters.NativeValue / 3.2808399
				return &Float{feet}, nil
			},
		},
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Int{},
				},
				Returns: &Float{},
			},
			Call: func(args ...Value) (Value, error) {
				meters := args[0].(*Int)
				feet := float64(meters.NativeValue ) / 3.2808399
				return &Float{feet}, nil
			},
		},
	},
}

var Sqm2Sqft = &Builtin {
	Name: "sqm2sqft",
	NativeCalls: []*NativeCall{
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Float{},
				},
				Returns: &Float{},
			},
			Call: func(args ...Value) (Value, error) {
				meters := args[0].(*Float)
				feet := meters.NativeValue * math.Pow(3.2808399, 2)
				return &Float{feet}, nil
			},
		},
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Int{},
				},
				Returns: &Float{},
			},
			Call: func(args ...Value) (Value, error) {
				meters := args[0].(*Int)
				feet := float64(meters.NativeValue ) * math.Pow(3.2808399, 2)
				return &Float{feet}, nil
			},
		},
	},
}

var Sqft2Sqm = &Builtin {
	Name: "sqft2sqm",
	NativeCalls: []*NativeCall{
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Float{},
				},
				Returns: &Float{},
			},
			Call: func(args ...Value) (Value, error) {
				meters := args[0].(*Float)
				feet := meters.NativeValue / math.Pow(3.2808399, 2)
				return &Float{feet}, nil
			},
		},
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Int{},
				},
				Returns: &Float{},
			},
			Call: func(args ...Value) (Value, error) {
				meters := args[0].(*Int)
				feet := float64(meters.NativeValue ) / math.Pow(3.2808399, 2)
				return &Float{feet}, nil
			},
		},
	},
}
