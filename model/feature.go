package model

type Feature struct {
	ID         interface{}       `json:"id"`
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties"`
	Geometry   interface{}       `json:"geometry"`
}
