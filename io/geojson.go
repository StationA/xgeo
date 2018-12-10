package io

import (
	"github.com/json-iterator/go"
	"io"
)

const ParseBufferSize = 16 * 1024

type GeoJSONReader struct {
	input io.Reader
}

func NewGeoJSONReader(input io.Reader) (*GeoJSONReader, error) {
	return &GeoJSONReader{
		input,
	}, nil
}

func (g *GeoJSONReader) Read(out chan map[string]interface{}) error {
	cfg := jsoniter.Config{}
	dec := jsoniter.Parse(cfg.Froze(), g.input, ParseBufferSize)
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
