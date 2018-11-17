package io

import (
	shp "github.com/jonas-p/go-shp"
	"strings"
)

type ShapefileReader struct {
	reader shp.SequentialReader
}

func NewShapefileReader(filename string) (*ShapefileReader, error) {
	var reader shp.SequentialReader
	var err error
	if strings.HasSuffix(filename, ".zip") {
		reader, err = shp.OpenZip(filename)
	} else {
		reader, err = shp.Open(filename)
	}
	if err != nil {
		return nil, err
	}
	return &ShapefileReader{
		reader,
	}, nil
}

func (s *ShapefileReader) Read(out chan map[string]interface{}) error {
	defer s.reader.Close()
	fields := s.reader.Fields()
	for s.reader.Next() {
		_, shape := s.reader.Shape()
		geom := shapeToGeometry(shape)
		properties := make(map[string]interface{})
		for i, field := range fields {
			attr := s.reader.Attribute(i)
			properties[field.String()] = attr
		}
		out <- map[string]interface{}{
			"type":       "Feature",
			"geometry":   geom,
			"properties": properties,
		}
	}
	return nil
}

func shapeToGeometry(shape shp.Shape) map[string]interface{} {
	switch s := shape.(type) {
	case *shp.Polygon:
		coords := buildCoordinates(s.Parts, s.Points)
		return map[string]interface{}{
			"type":        "MultiPolygon",
			"coordinates": [][][][]float64{coords},
		}
	case *shp.PolyLine:
		coords := buildCoordinates(s.Parts, s.Points)
		return map[string]interface{}{
			"type":        "MultiLineString",
			"coordinates": [][][][]float64{coords},
		}
	case *shp.Point:
		return map[string]interface{}{
			"type": "Point",
			"coordinates": []float64{
				s.X,
				s.Y,
			},
		}
	}
	return nil
}

func buildCoordinates(parts []int32, points []shp.Point) [][][]float64 {
	numParts := len(parts)
	var coords [][][]float64
	for i, start := range parts {
		var partCoords [][]float64
		end := len(points)
		if i < numParts-1 {
			end = int(parts[i+1])
		}
		for _, point := range points[start:end] {
			partCoords = append(partCoords, []float64{
				point.X,
				point.Y,
			})
		}
		coords = append(coords, partCoords)
	}
	return coords
}
