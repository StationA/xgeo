package vm

import (
	"fmt"
)

type Value interface {
	Add(Value) (Value, error)
	Sub(Value) (Value, error)
	Mul(Value) (Value, error)
	Div(Value) (Value, error)
	Eq(Value) (*Bool, error)
	Raw() interface{}
	Type() string
	String() string
}

func InvalidOperation(op string, left, right Value) error {
	return fmt.Errorf("Invalid operation '%s %s %s'\n", left.Type(), op, right.Type())
}

type Int struct {
	NativeValue int
}

func (i *Int) Raw() interface{} {
	return i.NativeValue
}

func (i *Int) Type() string {
	return "int"
}

func (i *Int) String() string {
	return fmt.Sprintf("{int %d}", i.NativeValue)
}

func (i *Int) Add(val Value) (Value, error) {
	switch val := val.(type) {
	case *Int:
		return &Int{i.NativeValue + val.NativeValue}, nil
	case *Float:
		return &Float{float64(i.NativeValue) + val.NativeValue}, nil
	}
	return nil, InvalidOperation("+", i, val)
}

func (i *Int) Sub(val Value) (Value, error) {
	switch val := val.(type) {
	case *Int:
		return &Int{i.NativeValue - val.NativeValue}, nil
	case *Float:
		return &Float{float64(i.NativeValue) - val.NativeValue}, nil
	}
	return nil, InvalidOperation("-", i, val)
}

func (i *Int) Mul(val Value) (Value, error) {
	switch val := val.(type) {
	case *Int:
		return &Int{i.NativeValue * val.NativeValue}, nil
	case *Float:
		return &Float{float64(i.NativeValue) * val.NativeValue}, nil
	}
	return nil, InvalidOperation("*", i, val)
}

func (i *Int) Div(val Value) (Value, error) {
	switch val := val.(type) {
	case *Int:
		return &Int{i.NativeValue / val.NativeValue}, nil
	case *Float:
		return &Float{float64(i.NativeValue) / val.NativeValue}, nil
	}
	return nil, InvalidOperation("/", i, val)
}

func (i *Int) Eq(val Value) (*Bool, error) {
	switch val := val.(type) {
	case *Int:
		return &Bool{i.NativeValue == val.NativeValue}, nil
	case *Float:
		return &Bool{float64(i.NativeValue) == val.NativeValue}, nil
	}
	return &Bool{false}, nil
}

type Float struct {
	NativeValue float64
}

func (f *Float) Raw() interface{} {
	return f.NativeValue
}

func (f *Float) Type() string {
	return "float"
}

func (f *Float) String() string {
	return fmt.Sprintf("{float %f}", f.NativeValue)
}

func (f *Float) Add(val Value) (Value, error) {
	switch val := val.(type) {
	case *Int:
		return &Float{f.NativeValue + float64(val.NativeValue)}, nil
	case *Float:
		return &Float{f.NativeValue + val.NativeValue}, nil
	}
	return nil, InvalidOperation("+", f, val)
}

func (f *Float) Sub(val Value) (Value, error) {
	switch val := val.(type) {
	case *Int:
		return &Float{f.NativeValue - float64(val.NativeValue)}, nil
	case *Float:
		return &Float{f.NativeValue - val.NativeValue}, nil
	}
	return nil, InvalidOperation("-", f, val)
}

func (f *Float) Mul(val Value) (Value, error) {
	switch val := val.(type) {
	case *Int:
		return &Float{f.NativeValue * float64(val.NativeValue)}, nil
	case *Float:
		return &Float{f.NativeValue * val.NativeValue}, nil
	}
	return nil, InvalidOperation("*", f, val)
}

func (f *Float) Div(val Value) (Value, error) {
	switch val := val.(type) {
	case *Int:
		return &Float{f.NativeValue / float64(val.NativeValue)}, nil
	case *Float:
		return &Float{f.NativeValue / val.NativeValue}, nil
	}
	return nil, InvalidOperation("/", f, val)
}

func (f *Float) Eq(val Value) (*Bool, error) {
	switch val := val.(type) {
	case *Int:
		return &Bool{f.NativeValue == float64(val.NativeValue)}, nil
	case *Float:
		return &Bool{f.NativeValue == val.NativeValue}, nil
	}
	return &Bool{false}, nil
}

type Bool struct {
	NativeValue bool
}

func (b *Bool) Raw() interface{} {
	return b.NativeValue
}

func (b *Bool) Type() string {
	return "bool"
}

func (b *Bool) String() string {
	return fmt.Sprintf("{bool %v}", b.NativeValue)
}

func (b *Bool) Add(val Value) (Value, error) {
	return nil, InvalidOperation("+", b, val)
}

func (b *Bool) Sub(val Value) (Value, error) {
	return nil, InvalidOperation("-", b, val)
}

func (b *Bool) Mul(val Value) (Value, error) {
	return nil, InvalidOperation("*", b, val)
}

func (b *Bool) Div(val Value) (Value, error) {
	return nil, InvalidOperation("/", b, val)
}

func (b *Bool) Eq(val Value) (*Bool, error) {
	switch val := val.(type) {
	case *Bool:
		return &Bool{b.NativeValue == val.NativeValue}, nil
	}
	return &Bool{false}, nil
}

type Str struct {
	NativeValue string
}

func (s *Str) Raw() interface{} {
	return s.NativeValue
}

func (s *Str) Type() string {
	return "str"
}

func (s *Str) String() string {
	return fmt.Sprintf("{str %s}", s.NativeValue)
}

func (s *Str) Add(val Value) (Value, error) {
	switch val := val.(type) {
	case *Str:
		return &Str{s.NativeValue + val.NativeValue}, nil
	}
	return nil, InvalidOperation("+", s, val)
}

func (s *Str) Sub(val Value) (Value, error) {
	return nil, InvalidOperation("-", s, val)
}

func (s *Str) Mul(val Value) (Value, error) {
	return nil, InvalidOperation("*", s, val)
}

func (s *Str) Div(val Value) (Value, error) {
	return nil, InvalidOperation("/", s, val)
}

func (s *Str) Eq(val Value) (*Bool, error) {
	switch val := val.(type) {
	case *Str:
		return &Bool{s.NativeValue == val.NativeValue}, nil
	}
	return &Bool{false}, nil
}

type Raw struct {
	NativeValue interface{}
}

func (r *Raw) Type() string {
	return "raw"
}

func (r *Raw) Raw() interface{} {
	return r.NativeValue
}

func (r *Raw) String() string {
	return fmt.Sprintf("{raw %+v}", r.Raw())
}

func (r *Raw) Add(val Value) (Value, error) {
	return nil, InvalidOperation("+", r, val)
}

func (r *Raw) Sub(val Value) (Value, error) {
	return nil, InvalidOperation("-", r, val)
}

func (r *Raw) Mul(val Value) (Value, error) {
	return nil, InvalidOperation("*", r, val)
}

func (r *Raw) Div(val Value) (Value, error) {
	return nil, InvalidOperation("/", r, val)
}

func (r *Raw) Eq(val Value) (*Bool, error) {
	switch val := val.(type) {
	case *Raw:
		return &Bool{r.Raw() == val.Raw()}, nil
	}
	return &Bool{false}, nil
}
