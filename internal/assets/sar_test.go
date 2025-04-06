package assets

import (
	"encoding/base64"
	"sync"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestValidSarDataLoad(t *testing.T) {
	sar := SAR()
	if len(sar) < MINIMUM_SAR_SIZE {
		t.Fatalf("expected >= %d SAR entries but saw %d", MINIMUM_SAR_SIZE, len(sar))
	}
	sarDataLoad = sync.Once{} // reset data loading
}

func TestValidSarIndexLoad(t *testing.T) {
	idx := SARIndex()
	if len(idx) < MINIMUM_SAR_SIZE {
		t.Fatalf("expected >= %d SAR index entries but saw %d", MINIMUM_SAR_SIZE, len(idx))
	}
	sarDataLoad = sync.Once{} // reset data loading
}

func TestInvalidSarGzip(t *testing.T) {
	defer testlib.AssertPanicWithText(t, `error unwrapping SAR data: EOF`)
	loadSAR([]byte{})
}

func TestInvalidSarDecode(t *testing.T) {
	defer testlib.AssertPanicWithText(t, `error decoding SAR data: unexpected end of JSON input`)
	loadSAR(fixtureInvalidEncodedSar())

}

func TestInvalidSarEmpty(t *testing.T) {
	defer testlib.AssertPanicWithText(t, `error validating SAR data, len too small: 0`)
	loadSAR(fixtureInvalidEmptySar())
}

// -----------------------------------------------------------------------------------------------
// Helper Functions
// -----------------------------------------------------------------------------------------------

// A gzip-compressed representation of '{'
func fixtureInvalidEncodedSar() []byte {
	invalidEncoded := `H4sIAAAAAAAAA6sGADlH1RUBAAAA`
	invalidDecoded, err := base64.StdEncoding.DecodeString(invalidEncoded)
	if err != nil {
		panic(err)
	}

	return invalidDecoded
}

// A gzip-compressed representation of '[]'
func fixtureInvalidEmptySar() []byte {
	invalidEncoded := `H4sIAAAAAAAAA4uOBQApu0wNAgAAAA==`
	invalidDecoded, err := base64.StdEncoding.DecodeString(invalidEncoded)
	if err != nil {
		panic(err)
	}

	return invalidDecoded
}
