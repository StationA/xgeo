package io

import (
	"github.com/bcicen/jstream"
	"github.com/stationa/xgeo/model"
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

func (g *GeoJSONReader) Read(out chan interface{}) error {
	defer g.inFile.Close()
	decoder := jstream.NewDecoder(g.inFile, 2)
	for mv := range decoder.Stream() {
		f := mv.Value.(map[string]interface{})
		out <- &model.Feature{
			ID:         f["id"],
			Type:       f["type"].(string),
			Properties: f["properties"].(map[string]string),
			Geometry:   f["geometry"],
		}
	}
	return nil
}
