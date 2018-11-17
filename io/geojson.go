package io

import (
	"github.com/json-iterator/go"
	"os"
)

type GeoJSONReader struct {
	inFile *os.File
}

func NewGeoJSONReader(filename string) (*GeoJSONReader, error) {
	inFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return &GeoJSONReader{
		inFile: inFile,
	}, nil
}

func (g *GeoJSONReader) Read(out chan map[string]interface{}) error {
	defer g.inFile.Close()
	cfg := jsoniter.Config{}
	dec := jsoniter.Parse(cfg.Froze(), g.inFile, 4096)
	skipToFeatures(dec)
	for dec.ReadArray() {
		feature := dec.Read().(map[string]interface{})
		out <- feature
	}
	return nil
}

func skipToFeatures(dec *jsoniter.Iterator) {
	for {
		field := dec.ReadObject()
		if field == "" {
			panic("GeoJSON object must have a \"features\" array")
		}
		if field == "features" {
			break
		}
		// Skip whatever the value is
		dec.Skip()
	}
}
