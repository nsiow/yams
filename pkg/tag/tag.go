package tag

// Tag defines a standard AWS resource/principal tag
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
