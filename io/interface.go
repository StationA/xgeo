package io

type FeatureReader interface {
	Read(out chan interface{}) error
}
