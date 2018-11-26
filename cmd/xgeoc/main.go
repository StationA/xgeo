package main

import (
	"fmt"
	gx "github.com/stationa/xgeo/vm"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
)

var (
	gxFile = kingpin.Arg("gx", "GX script").Required().File()
)

func main() {
	kingpin.Parse()

	data, err := ioutil.ReadAll(*gxFile)
	if err != nil {
		panic(err)
	}
	compiler := &gx.XGeoCompiler{Buffer: string(data)}
	compiler.Prepare()
	if err := compiler.Parse(); err != nil {
		fmt.Println(err)
		panic(err)
	}
	if err := compiler.Compile(); err != nil {
		fmt.Println(err)
		panic(err)
	}
	vm := compiler.InitVM()
	vm.DumpConstants()
	vm.DumpCode()
}
