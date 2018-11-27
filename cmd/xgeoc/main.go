package main

import (
	"encoding/json"
	"fmt"
	gx "github.com/stationa/xgeo/vm"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
)

var (
	debug       = kingpin.Flag("debug", "Debug mode").Short('d').Bool()
	dumpOnCrash = kingpin.Flag("dump-on-crash", "Dumps the VM state on crash").Short('x').Bool()
	test        = kingpin.Flag("test", "Test the GX script with no input").Short('t').Bool()
	gxFile      = kingpin.Arg("gx", "GX script").Required().File()
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

	if *debug {
		compiler.PrintSyntaxTree()
		vm.SetDebug(true)
		if !*test {
			vm.DumpConstants()
			vm.DumpCode()
		}
	}
	if *dumpOnCrash {
		vm.SetDumpOnCrash(true)
	}

	if *test {
		output := make(chan interface{})
		go func() {
			defer close(output)
			vm.Run(nil, output)
		}()
		for o := range output {
			json, err := json.Marshal(o)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(json))
		}
	}
}
