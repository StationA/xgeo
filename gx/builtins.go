package gx

import (
	"github.com/Shopify/go-lua"
)

var Builtins = map[string]lua.Function{
	// Geo builtins
	"area":       Area,
	"centroid":   Centroid,
	"bbox":       BoundingBox,
	"perimeter":  Perimeter,
	"project":    Project,
	"simplify":   Simplify,
	"pluscode":   EncodePlusCode,
	"unpluscode": DecodePlusCode,
	"ft2m":       Ft2M,
	"m2ft":       M2Ft,
	"sqft2sqm":   Sqft2Sqm,
	"sqm2sqft":   Sqm2Sqft,
}

func RegisterBuiltins(L *lua.State) {
	for name, f := range Builtins {
		L.Register(name, f)
	}
}
