package sim

import (
	"strings"
)

// fold defines the internal implementation of case-folding used for bag keys
var fold = strings.ToLower

// Bag implements a generic looser map interface which case-folds its string keys
type Bag[T any] struct {
	innerMap map[string]T
}

// NewBag creates and returns a new case-folded bag with the specified value type T
func NewBag[T any]() Bag[T] {
	return Bag[T]{innerMap: map[string]T{}}
}

// NewBagFroMap creates and returns a new case-folded bag with the specified value type T,
// seeded using the folded key/values from the provided map
func NewBagFromMap[T any](other map[string]T) Bag[T] {
	b := Bag[T]{innerMap: make(map[string]T)}

	for k, v := range other {
		b.Put(k, v)
	}

	return b
}

// Get folds the input key and then returns the matched value (or the zero-value for the
// registered type if a match cannot be found)
func (b *Bag[T]) Get(k string) T {
	v := b.innerMap[fold(k)]
	return v
}

// Check folds the input key and then checks the bag for a value matching the provided key
func (b *Bag[T]) Check(k string) (T, bool) {
	v, ok := b.innerMap[fold(k)]
	return v, ok
}

// Put saves the provided value to our Bag after folding the input key
func (b *Bag[T]) Put(k string, v T) {
	b.innerMap[fold(k)] = v
}

// Delete removes the key+value pair
func (b *Bag[T]) Delete(k string) {
	delete(b.innerMap, fold(k))
}
