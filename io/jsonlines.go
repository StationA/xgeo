package io

import (
	"encoding/json"
	"github.com/stationa/xgeo/model"
	"github.com/stationa/xgeo/util"
	"io"
	"os"
)

type JSONLinesReader struct {
	inFile *os.File
}

func NewJSONLinesReader(filename string) (*JSONLinesReader, error) {
	inFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return &JSONLinesReader{
		inFile: inFile,
	}, nil
}

func (j *JSONLinesReader) Read(out chan *model.Feature) error {
	defer j.inFile.Close()
	dec := json.NewDecoder(j.inFile)
	for {
		var val map[string]interface{}
		err := dec.Decode(&val)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		out <- &model.Feature{
			ID:         val["id"],
			Type:       val["type"].(string),
			Properties: util.CastProps(val["properties"]),
			Geometry:   val["geometry"],
		}
	}
	return nil
}
