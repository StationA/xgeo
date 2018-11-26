package vm

import (
	"fmt"
	"strings"
)

func typeError(val Value, funcSig string) error {
	return fmt.Errorf("Invalid type %T for function %s", val, funcSig)
}

func Lower(args ...Value) (Value, error) {
	val := args[0]
	str, isStr := val.(*Str)
	if !isStr {
		return nil, typeError(val, "lower(str) -> str")
	}
	s := strings.ToLower(str.NativeValue)
	return &Str{s}, nil
}

func Upper(args ...Value) (Value, error) {
	val := args[0]
	str, isStr := val.(*Str)
	if !isStr {
		return nil, typeError(val, "upper(str) -> str")
	}
	s := strings.ToUpper(str.NativeValue)
	return &Str{s}, nil
}

func Title(args ...Value) (Value, error) {
	val := args[0]
	str, isStr := val.(*Str)
	if !isStr {
		return nil, typeError(val, "title(str) -> str")
	}
	s := strings.ToTitle(str.NativeValue)
	return &Str{s}, nil
}

func Strip(args ...Value) (Value, error) {
	val := args[0]
	str, isStr := val.(*Str)
	if !isStr {
		return nil, typeError(val, "strip(str) -> str")
	}
	s := strings.TrimSpace(str.NativeValue)
	return &Str{s}, nil
}

func CastBool(args ...Value) (Value, error) {
	val := args[0]
	switch t := val.(type) {
	case *Bool:
		return t, nil
	case *Str:
		return ParseBool(t.NativeValue), nil
	default:
		return nil, typeError(val, "bool(bool|str) -> bool")
	}
}

func CastInt(args ...Value) (Value, error) {
	val := args[0]
	switch t := val.(type) {
	case *Int:
		return t, nil
	case *Float:
		return &Int{int(t.NativeValue)}, nil
	case *Str:
		return ParseInt(t.NativeValue), nil
	default:
		return nil, typeError(val, "int(int|float|str) -> int")
	}
}

func CastFloat(args ...Value) (Value, error) {
	val := args[0]
	switch t := val.(type) {
	case *Int:
		return &Float{float64(t.NativeValue)}, nil
	case *Float:
		return t, nil
	case *Str:
		return ParseFloat(t.NativeValue), nil
	default:
		return nil, typeError(val, "float(int|float|str) -> float")
	}
}

func CastStr(args ...Value) (Value, error) {
	val := args[0]
	switch t := val.(type) {
	case *Bool:
		if t.NativeValue {
			return &Str{"true"}, nil
		}
		return &Str{"false"}, nil
	case *Int:
		return &Str{fmt.Sprintf("%d", t.NativeValue)}, nil
	case *Float:
		return &Str{fmt.Sprintf("%f", t.NativeValue)}, nil
	case *Str:
		return t, nil
	default:
		return nil, typeError(val, "str(bool|int|float|str) -> str")
	}
}
