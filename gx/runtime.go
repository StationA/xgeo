package gx

import (
	"github.com/Shopify/go-lua"
	"github.com/stationa/xgeo/gx/util"
)

type XGeoVM struct {
	ScriptFilename string
	Output         chan map[string]interface{}
	L              *lua.State
}

func NewVM(filename string) *XGeoVM {
	output := make(chan map[string]interface{})
	return &XGeoVM{ScriptFilename: filename, Output: output}
}

func (vm *XGeoVM) Init() error {
	vm.L = lua.NewState()
	lua.OpenLibraries(vm.L)
	RegisterBuiltins(vm.L)
	return lua.DoFile(vm.L, vm.ScriptFilename)
}

func (vm *XGeoVM) RegisterBuiltins() {
	vm.L.Register("emit", func(L *lua.State) int {
		feature, _ := util.PullTable(vm.L, -1)
		if feature != nil {
			vm.Output <- feature.(map[string]interface{})
		}
		return 0
	})
	RegisterBuiltins(vm.L)
}

func (vm *XGeoVM) Run(input map[string]interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			// TODO: Enter debug mode?
			panic(r)
		}
	}()
	vm.RegisterBuiltins()
	vm.L.Global("process")
	util.DeepPush(vm.L, input)
	vm.L.Call(1, 0)
	return nil
}

func (vm *XGeoVM) Close() {
	close(vm.Output)
}
