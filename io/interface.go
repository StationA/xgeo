package io

import (
	"github.com/stationa/xgeo/model"
)

type FeatureReader interface {
	Read(out chan *model.Feature) error
}
