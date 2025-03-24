package assets

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

// -----------------------------------------------------------------------------------------------
// Test Functions
// -----------------------------------------------------------------------------------------------

// TestValidSarDataLoad confirms the successful loading of valid data
func TestValidSarDataLoad(t *testing.T) {
	sar := SarData()
	if len(sar) < MINIMUM_SAR_SIZE {
		t.Fatalf(fmt.Sprintf("expected >= %d SAR entries but saw %d",
			MINIMUM_SAR_SIZE,
			len(sar),
		))
	}
}

// TestInvalidSarGzip confirms the failed loading of corrupted gzip data
func TestInvalidSarGzip(t *testing.T) {
	defer testlib.AssertPanicWithText(t,
		`error unwrapping SAR data: EOF`)
	loadSarData([]byte{})
}

// TestInvalidSarDecode confirms the failed loading of corrupted JSON data
func TestInvalidSarDecode(t *testing.T) {
	defer testlib.AssertPanicWithText(t,
		`error decoding SAR data: unexpected end of JSON input`)
	loadSarData(fixtureInvalidEncodedSar())

}

// TestInvalidSarEmpty confirms the failed loading of a too-short SAR list
func TestInvalidSarEmpty(t *testing.T) {
	defer testlib.AssertPanicWithText(t,
		`error validating SAR data, len too small: 0`)
	loadSarData(fixtureInvalidEmptySar())
}

// -----------------------------------------------------------------------------------------------
// Helper Functions
// -----------------------------------------------------------------------------------------------

// A gzip-compressed representation of '{'
func fixtureInvalidEncodedSar() []byte {
	invalidEncoded := `H4sICLVr4GcAA2V4YW1wbGUuanNvbgCr5gIA3FvHbQIAAAA=`
	invalidDecoded, err := base64.StdEncoding.DecodeString(invalidEncoded)
	if err != nil {
		panic(err)
	}

	return invalidDecoded
}

// A gzip-compressed representation of '{}'
func fixtureInvalidEmptySar() []byte {
	invalidEncoded := `H4sICPlu4GcAA2V4YW1wbGUuanNvbgCrruUCAAawod0DAAAA`
	invalidDecoded, err := base64.StdEncoding.DecodeString(invalidEncoded)
	if err != nil {
		panic(err)
	}

	return invalidDecoded
}
