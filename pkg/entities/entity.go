package entities

type Entity interface {
	Key() string
	Repr() (any, error)
}
