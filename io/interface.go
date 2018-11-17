package io

type FeatureReader interface {
	Read(out chan map[string]interface{}) error
}
