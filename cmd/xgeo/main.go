package main

import (
	"compress/bzip2"
	"compress/gzip"
	"encoding/json"
	"fmt"
	gio "github.com/stationa/xgeo/io"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"strings"
)

var (
	src = kingpin.Arg("source", "Source file").File()
)

func main() {
	kingpin.Parse()

	var reader gio.FeatureReader
	var err error

	filename := (*src).Name()
	if strings.HasSuffix(filename, ".zip") || strings.HasSuffix(filename, ".shp") {
		reader, err = gio.NewShapefileReader(filename)
	} else {
		var input io.Reader = *src
		if strings.HasSuffix(filename, ".gz") {
			input, _ = gzip.NewReader(input)
			filename = filename[:len(filename)-len(".gz")]
		}
		if strings.HasSuffix(filename, ".bz2") {
			input = bzip2.NewReader(input)
			filename = filename[:len(filename)-len(".bz2")]
		}
		if strings.HasSuffix(filename, ".geojson") {
			reader, err = gio.NewGeoJSONReader(input)
		}
	}
	if err != nil {
		panic(err)
	}
	features := make(chan map[string]interface{})
	go func() {
		defer close(features)
		err := reader.Read(features)
		if err != nil {
			panic(err)
		}
	}()

	for feature := range features {
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
