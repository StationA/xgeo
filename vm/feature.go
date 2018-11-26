package vm

import (
	"github.com/stationa/xgeo/model"
)

func (vm *XGeoVM) Access(feature *model.Feature, key string) Value {
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
