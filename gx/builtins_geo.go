package gx

import (
	"github.com/Shopify/go-lua"
	"github.com/google/open-location-code/go"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/paulmach/orb/planar"
	"github.com/paulmach/orb/project"
	"github.com/paulmach/orb/simplify"
	"github.com/stationa/xgeo/gx/util"
	"math"
	"strings"
)

const (
	DefaultPlusCodeLength = 10
	FeetPerMeter          = 3.2808399
)

func Area(L *lua.State) int {
	geometry := util.CheckGeometry(L, 1)
	area := geo.Area(geometry)
	L.PushNumber(area)
	return 1
}

func Perimeter(L *lua.State) int {
	geometry := util.CheckGeometry(L, 1)
	perim := geo.Length(geometry)
	L.PushNumber(perim)
	return 1
}

func BoundingBox(L *lua.State) int {
	geometry := util.CheckGeometry(L, 1)
	bbox := geometry.Bound()
	util.PushGeometry(L, bbox)
	return 1
}

func Centroid(L *lua.State) int {
	geometry := util.CheckGeometry(L, 1)
	centroid, _ := planar.CentroidArea(geometry)
	util.PushGeometry(L, centroid)
	return 1
}

func getProjection(proj string) orb.Projection {
	switch strings.ToLower(proj) {
	case "wgs84":
		return func(p orb.Point) orb.Point { return p }
	case "mercator":
		return project.Mercator.ToWGS84
	default:
		panic("Unsupported projection: " + proj)
	}
}

func Project(L *lua.State) int {
	geometry := util.CheckGeometry(L, 1)
	fromProj := lua.CheckString(L, 2)
	projection := getProjection(fromProj)
	projected := project.Geometry(geometry, projection)
	util.PushGeometry(L, projected)
	return 1
}

func Simplify(L *lua.State) int {
	geometry := util.CheckGeometry(L, 1)
	threshold := lua.CheckNumber(L, 2)
	s := simplify.DouglasPeucker(threshold)
	simplified := s.Simplify(geometry)
	util.PushGeometry(L, simplified)
	return 1
}

func EncodePlusCode(L *lua.State) int {
	point := util.CheckGeometry(L, 1).(orb.Point)
	length := lua.OptInteger(L, 2, DefaultPlusCodeLength)
	pluscode := olc.Encode(point[1], point[0], length)
	L.PushString(pluscode)
	return 1
}

func DecodePlusCode(L *lua.State) int {
	pluscode := lua.CheckString(L, 1)
	decoded, _ := olc.Decode(pluscode)
	lat, lng := decoded.Center()
	centroid := orb.Point{lng, lat}
	util.PushGeometry(L, centroid)
	return 1
}

func M2Ft(L *lua.State) int {
	meters := lua.CheckNumber(L, 1)
	feet := meters * FeetPerMeter
	L.PushNumber(feet)
	return 1
}

func Ft2M(L *lua.State) int {
	feet := lua.CheckNumber(L, 1)
	meters := feet / FeetPerMeter
	L.PushNumber(meters)
	return 1
}

func Sqm2Sqft(L *lua.State) int {
	sqm := lua.CheckNumber(L, 1)
	sqft := sqm * math.Pow(FeetPerMeter, 2)
	L.PushNumber(sqft)
	return 1
}

func Sqft2Sqm(L *lua.State) int {
	sqft := lua.CheckNumber(L, 1)
	sqm := sqft / math.Pow(FeetPerMeter, 2)
	L.PushNumber(sqm)
	return 1
}
