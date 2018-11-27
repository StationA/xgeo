package gx

import (
	"fmt"
	"github.com/stationa/xgeo/model"
)

func (vm *XGeoVM) access(feature *model.Feature, key string) Value {
	switch key {
	case "id":
		return &Raw{feature.ID}
	case "type":
		return &Str{feature.Type}
	case "properties":
		return &Raw{feature.Properties}
	case "geometry":
		return &Raw{feature.Geometry}
	}
	return nil
}

func (vm *XGeoVM) mutate(feature *model.Feature, key string, val Value) error {
	switch key {
	case "id":
		feature.ID = val.Raw().(string)
	case "type":
		feature.Type = val.Raw().(string)
	default:
		panic(fmt.Errorf("Cannot mutate feature attribute: %s", key))
	}
	return nil
}
