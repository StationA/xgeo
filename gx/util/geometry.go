package util

import (
	"github.com/Shopify/go-lua"
	"github.com/paulmach/orb"
)

func CheckGeometry(L *lua.State, index int) orb.Geometry {
	val, err := PullTable(L, index)
	if err != nil {
		panic(err)
	}
	geom := val.(map[string]interface{})
	return orb.Clone(ToGeometry(geom))
}

func PushGeometry(L *lua.State, g orb.Geometry) {
	if g == nil {
		L.PushNil()
		return
	}
	geom := map[string]interface{}{
		"type":        g.GeoJSONType(),
		"coordinates": g,
	}
	DeepPush(L, geom)
}

func ToGeometry(geom map[string]interface{}) orb.Geometry {
	geomType := geom["type"].(string)
	coords := geom["coordinates"].([]interface{})
	switch geomType {
	case "Point":
		return toPoint(coords)
	case "MultiPoint":
		return toMultiPoint(coords)
	case "LineString":
		return toLineString(coords)
	case "MultiLineString":
		return toMultiLineString(coords)
	case "Polygon":
		return toPolygon(coords)
	case "MultiPolygon":
		return toMultiPolygon(coords)
	default:
		panic("Unsupported geometry: " + geomType)
	}
}

func toPoint(c []interface{}) orb.Point {
	var p orb.Point
	for i, l := range c {
		p[i] = l.(float64)
	}
	return p
}

func toMultiPoint(c []interface{}) orb.MultiPoint {
	var mp orb.MultiPoint
	for _, p := range c {
		mp = append(mp, toPoint(p.([]interface{})))
	}
	return mp
}

func toLineString(c []interface{}) orb.LineString {
	var ls orb.LineString
	for _, p := range c {
		ls = append(ls, toPoint(p.([]interface{})))
	}
	return ls
}

func toMultiLineString(c []interface{}) orb.MultiLineString {
	var mls orb.MultiLineString
	for _, l := range c {
		mls = append(mls, toLineString(l.([]interface{})))
	}
	return mls
}

func toRing(c []interface{}) orb.Ring {
	var r orb.Ring
	for _, p := range c {
		r = append(r, toPoint(p.([]interface{})))
	}
	return r
}

func toPolygon(c []interface{}) orb.Polygon {
	var p orb.Polygon
	for _, r := range c {
		p = append(p, toRing(r.([]interface{})))
	}
	return p
}

func toMultiPolygon(c []interface{}) orb.MultiPolygon {
	var mp orb.MultiPolygon
	for _, p := range c {
		mp = append(mp, toPolygon(p.([]interface{})))
	}
	return mp
}
