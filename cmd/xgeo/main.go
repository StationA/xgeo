package main

import (
	"encoding/json"
	"fmt"
	"github.com/stationa/xgeo/gx"
	"github.com/stationa/xgeo/io"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
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
func merge(cs ...chan map[string]interface{}) <-chan map[string]interface{} {
	var wg sync.WaitGroup
	out := make(chan map[string]interface{})

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c chan map[string]interface{}) {
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

	var inputs []chan map[string]interface{}
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
			input := make(chan map[string]interface{})
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
		// For now, just assuming stdin uses JSON lines format
		reader := io.ToJSONLinesReader(os.Stdin)
		input := make(chan map[string]interface{})
		inputs = append(inputs, input)
		go func() {
			defer close(input)
			err := reader.Read(input)
			if err != nil {
				panic(err)
			}
		}()
	}

	input := merge(inputs...)

	var outputs []chan map[string]interface{}
	if *gxFile != nil {
		for i := 0; i < *workers; i++ {
			vm := gx.NewVM((*gxFile).Name())
			outputs = append(outputs, vm.Output)
			if err := vm.Init(); err != nil {
				panic(err)
			}
			if *debug {
				// vm.SetDebug(true)
			}
			if *dumpOnCrash {
				// vm.SetDumpOnCrash(true)
			}
			go func() {
				defer vm.Close()
				for feature := range input {
					err := vm.Run(feature)
					if err != nil {
						panic(err)
					}
				}
			}()
		}
	} else {
		output := make(chan map[string]interface{})
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
