package entities

// Tag defines a standard AWS resource/principal tag
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Tag   string `json:"tag"`
}

// Tags defines a collection around a slice of Tag structs
type Tags = []Tag
