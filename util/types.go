package util

import (
	"github.com/stationa/xgeo/vm"
	"strconv"
)

func ParseInt(val string) vm.Value {
	v, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		panic(err)
	}
	return &vm.Int{int(v)}
}

func ParseFloat(val string) vm.Value {
	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		panic(err)
	}
	return &vm.Float{v}
}

func ParseBool(val string) vm.Value {
	v, err := strconv.ParseBool(val)
	if err != nil {
		panic(err)
	}
	return &vm.Bool{v}
}
