package entities

// Tag defines the general shape of an AWS metadata tag
type Tag struct {
	// Key refers to the tag key
	Key string

	// Value refers to the tag value
	Value string
}
