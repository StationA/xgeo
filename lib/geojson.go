package lib

import (
    "fmt"
    "encoding/json"
    "github.com/bcicen/jstream"
    "os"
)

type GeoJSONReader struct{
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

func (g *GeoJSONReader) Read() error {
    defer g.inFile.Close()
    decoder := jstream.NewDecoder(g.inFile, 2)
    for mv := range decoder.Stream() {
        json, err := json.Marshal(mv.Value)
        if err != nil {
            return err
        }
        fmt.Println(string(json))
    }
    return nil
}
