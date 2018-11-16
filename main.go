package main

import (
	"fmt"
	xgeo "github.com/stationa/xgeo/lib"
	"gopkg.in/alecthomas/kingpin.v2"
	"strings"
)

var (
	sourceFiles = kingpin.Arg("source", "Source file").Strings()
)

func main() {
	kingpin.Parse()
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
		fmt.Println("Using standard input")
	}
}
