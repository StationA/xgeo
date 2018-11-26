package main

import (
	"encoding/json"
	"fmt"
	io "github.com/stationa/xgeo/io"
	gx "github.com/stationa/xgeo/lang"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"runtime"
	"strings"
	"sync"
)

const (
	WorkersPerCPU = 2
)

var (
	sourceFiles = kingpin.Arg("source", "Source file").Strings()
	gxFile      = kingpin.Flag("gx", "GX script").Short('g').File()
	debug       = kingpin.Flag("debug", "Debug mode").Short('d').Bool()
	dumpOnCrash = kingpin.Flag("dump-on-crash", "Dumps the VM state on crash").Short('x').Bool()
	parallelism = kingpin.Flag("parallelism", "Number of cores to use").Short('p').Default("0").Int()
	workers     = kingpin.Flag("num-workers", "Number of workers to use").Short('w').Default("0").Int()
)

// Adapted from https://blog.golang.org/pipelines
func merge(cs ...chan interface{}) <-chan interface{} {
	var wg sync.WaitGroup
	out := make(chan interface{})

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c chan interface{}) {
		defer wg.Done()
		for n := range c {
			out <- n
		}
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func main() {
	kingpin.Parse()

	if *parallelism < 1 {
		*parallelism = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(*parallelism)

	if *workers < 1 {
		*workers = *parallelism * WorkersPerCPU
	}

	var inputs []chan interface{}
	if len(*sourceFiles) > 0 {
		for _, sourceFile := range *sourceFiles {
			var reader io.FeatureReader
			var err error
			if strings.HasSuffix(sourceFile, ".zip") || strings.HasSuffix(sourceFile, ".shp") {
				reader, err = io.NewShapefileReader(sourceFile)
			}
			if strings.HasSuffix(sourceFile, ".geojson") {
				reader, err = io.NewGeoJSONReader(sourceFile)
			}
			if strings.HasSuffix(sourceFile, ".jsonlines") {
				reader, err = io.NewJSONLinesReader(sourceFile)
			}
			if err != nil {
				panic(err)
			}
			input := make(chan interface{})
			inputs = append(inputs, input)
			go func() {
				defer close(input)
				err := reader.Read(input)
				if err != nil {
					panic(err)
				}
			}()
		}
	} else {
		fmt.Println("IMPLEMENT STDIN PROCESSING")
	}

	input := merge(inputs...)

	var outputs []chan interface{}
	if *gxFile != nil {
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

		for i := 0; i < *workers; i++ {
			output := make(chan interface{})
			outputs = append(outputs, output)
			vm := compiler.InitVM()
			if *debug {
				vm.SetDebug(true)
			}
			if *dumpOnCrash {
				vm.SetDumpOnCrash(true)
			}
			go func() {
				defer close(output)
				for feature := range input {
					vm.Run(feature, output)
				}
			}()
		}
	} else {
		output := make(chan interface{})
		outputs = append(outputs, output)
		go func() {
			defer close(output)
			for feature := range input {
				output <- feature
			}
		}()
	}

	output := merge(outputs...)

	for feature := range output {
		if feature == nil {
			continue
		}
		json, err := json.Marshal(feature)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(json))
	}
}
