package util

func CastProps(m interface{}) map[string]string {
	out := make(map[string]string)
	for k, v := range m.(map[string]interface{}) {
		out[k] = v.(string)
	}
	return out
}
