package lib

type Feature struct {
    Type string `json:"type"`
    Geometry Geometry `json:"geometry"`
    Properties map[string]string `json:"properties"`
}

type Geometry struct {
    Type        string `json:"type"`
    Coordinates interface{} `json:"coordinates"`
}
