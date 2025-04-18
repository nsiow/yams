package sim

import (
	"testing"
)

func TestNewPropertyBag(t *testing.T) {
	b := NewBag[string]()
	if len(b.innerMap) != 0 {
		t.Fatalf("unexpected non-empty contents for new bag: %+v", b.innerMap)
	}
}

func TestNewPropertyBagFromMap(t *testing.T) {
	b := NewBagFromMap(map[string]string{
		"the":      "quick",
		"brown":    "fox",
		"jumped":   "over",
		"the lazy": "dog",
	})

	if len(b.innerMap) != 4 {
		t.Fatalf("unexpected contents for new bag: %+v", b.innerMap)
	}

	if v := b.Get("the"); v != "quick" {
		t.Fatalf("unexpected value for key 'the': %s", v)
	}
	if v := b.Get("brown"); v != "fox" {
		t.Fatalf("unexpected value for key 'brown': %s", v)
	}
	if v := b.Get("jumped"); v != "over" {
		t.Fatalf("unexpected value for key 'jumped': %s", v)
	}
	if v := b.Get("the lazy"); v != "dog" {
		t.Fatalf("unexpected value for key 'the lazy': %s", v)
	}
}

func TestPropertyBagGet(t *testing.T) {
	b := NewBag[string]()

	b.Put("this", "that")
	b.Put("HONDA", "civic")

	// Check key exactly as saved
	v := b.Get("this")
	if v != "that" {
		t.Fatalf("unexpected value for key 'this': %s", v)
	}

	// Check key and rely on case folding
	v = b.Get("honda")
	if v != "civic" {
		t.Fatalf("unexpected value for key 'honda': %s", v)
	}

	// Check nonexistent key
	v = b.Get("404")
	if v != "" {
		t.Fatalf("unexpected value for key '404': %s", v)
	}
}

func TestPropertyBagCheck(t *testing.T) {
	b := NewBag[string]()

	b.Put("foo", "bar")
	b.Put("coloR", "bluE")

	// Check key exactly as saved
	v, ok := b.Check("foo")
	if !ok {
		t.Fatalf("unable to retrieve expected key: 'foo'")
	}
	if v != "bar" {
		t.Fatalf("unexpected value for key 'foo': %s", v)
	}

	// Check key and rely on case folding
	v, ok = b.Check("COLOR")
	if !ok {
		t.Fatalf("unable to retrieve expected key: 'COLOR'")
	}
	if v != "bluE" {
		t.Fatalf("unexpected value for key 'COLOR': %s", v)
	}

	// Check nonexistent key
	_, ok = b.Check("404")
	if ok {
		t.Fatalf("somehow able to retrieve unexpected key: '404'")
	}
}

func TestPropertyBagPut(t *testing.T) {
	b := NewBag[string]()

	// Save initial key
	b.Put("foo", "bar")

	// Check key exactly as saved
	v, ok := b.Check("foo")
	if !ok {
		t.Fatalf("unable to retrieve expected key: 'foo'")
	}
	if v != "bar" {
		t.Fatalf("unexpected value for key 'foo': %s", v)
	}

	// Overwrite with new value
	b.Put("fOo", "baz")

	// Check key again
	v, ok = b.Check("foo")
	if !ok {
		t.Fatalf("unable to retrieve expected key: 'foo'")
	}
	if v != "baz" {
		t.Fatalf("unexpected value for key 'foo': %s", v)
	}
}

func TestPropertyBagDelete(t *testing.T) {
	b := NewBag[string]()

	// Save initial key
	b.Put("foo", "bar")

	// Check key exactly as saved
	v, ok := b.Check("foo")
	if !ok {
		t.Fatalf("unable to retrieve expected key: 'foo'")
	}
	if v != "bar" {
		t.Fatalf("unexpected value for key 'foo': %s", v)
	}

	// Delete key
	b.Delete("fOo")

	// Check key again
	_, ok = b.Check("foo")
	if ok {
		t.Fatalf("somehow able to retrieve unexpected key: 'foo'")
	}
}
