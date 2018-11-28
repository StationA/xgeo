package model

type Feature struct {
	ID         interface{}       `json:"id,omitempty"`
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties"`
	Geometry   interface{}       `json:"geometry"`
}

func CastProps(m interface{}) map[string]string {
	out := make(map[string]string)
	for k, v := range m.(map[string]interface{}) {
		out[k] = v.(string)
	}
	return out
}
