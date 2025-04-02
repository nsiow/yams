package assets

import (
	"encoding/base64"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

// -----------------------------------------------------------------------------------------------
// Test Functions
// -----------------------------------------------------------------------------------------------

// TestValidSarDataLoad confirms the successful loading of valid data
func TestValidSarDataLoad(t *testing.T) {
	sar := SAR()
	if len(sar) < MINIMUM_SAR_SIZE {
		t.Fatalf("expected >= %d SAR entries but saw %d",
			MINIMUM_SAR_SIZE,
			len(sar),
		)
	}
}

// TestInvalidSarGzip confirms the failed loading of corrupted gzip data
func TestInvalidSarGzip(t *testing.T) {
	defer testlib.AssertPanicWithText(t, `error unwrapping SAR data: EOF`)
	loadSAR([]byte{})
}

// TestInvalidSarDecode confirms the failed loading of corrupted JSON data
func TestInvalidSarDecode(t *testing.T) {
	defer testlib.AssertPanicWithText(t, `error decoding SAR data: unexpected end of JSON input`)
	loadSAR(fixtureInvalidEncodedSar())

}

// TestInvalidSarEmpty confirms the failed loading of a too-short SAR list
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
