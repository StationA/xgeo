package main

import (
	"fmt"
	gx "github.com/stationa/xgeo/lang"
	xgeo "github.com/stationa/xgeo/lib"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"strings"
)

var (
	sourceFiles = kingpin.Arg("source", "Source file").Strings()
	gxFile      = kingpin.Flag("gx", "GX script").Short('g').String()
)

func main() {
	kingpin.Parse()

	if *gxFile != "" {
		data, err := ioutil.ReadFile(*gxFile)
		if err != nil {
			panic(err)
		}
		parser := &gx.XGeoParser{Buffer: string(data)}
		parser.Init()
		err = parser.Parse()
		if err != nil {
			fmt.Print(err)
			panic(err)
		}
		parser.PrintSyntaxTree()
	}
	if len(*sourceFiles) > 0 {
		for _, sourceFile := range *sourceFiles {
			if strings.HasSuffix(sourceFile, ".zip") || strings.HasSuffix(sourceFile, ".shp") {
				reader, _ := xgeo.NewShapefileReader(sourceFile)
				reader.Read()
			}
			if strings.HasSuffix(sourceFile, ".geojson") {
				reader, _ := xgeo.NewGeoJSONReader(sourceFile)
				reader.Read()
			}
		}
	} else {
		fmt.Println("IMPLEMENT STDIN PROCESSING")
	}
}
