package io

import (
	"encoding/json"
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

func ToJSONLinesReader(file *os.File) *JSONLinesReader {
	return &JSONLinesReader{
		inFile: file,
	}
}

func (j *JSONLinesReader) Read(out chan map[string]interface{}) error {
	defer j.inFile.Close()
	dec := json.NewDecoder(j.inFile)
	for {
		var feature map[string]interface{}
		err := dec.Decode(&feature)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		out <- feature
	}
	return nil
}
