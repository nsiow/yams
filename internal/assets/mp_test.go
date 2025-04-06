package assets

import (
	"encoding/base64"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestValidPolicyDataLoad(t *testing.T) {
	policies := ManagedPolicyData()
	if len(policies) < MINIMUM_POLICYSET_SIZE {
		t.Fatalf("expected >= %d policies but saw %d",
			MINIMUM_POLICYSET_SIZE,
			len(policies),
		)
	}
}

func TestInvalidPolicyGzip(t *testing.T) {
	defer testlib.AssertPanicWithText(t,
		`error unwrapping managed policy data: EOF`)
	loadManagedPolicyData([]byte{})
}

func TestInvalidPolicyDecode(t *testing.T) {
	defer testlib.AssertPanicWithText(t,
		`error decoding managed policy data: unexpected end of JSON input`)
	loadManagedPolicyData(fixtureInvalidEncodedPolicy())

}

func TestInvalidPolicyEmpty(t *testing.T) {
	defer testlib.AssertPanicWithText(t,
		`error validating managed policy data, len too small: 0`)
	loadManagedPolicyData(fixtureInvalidEmptyPolicy())
}

// -----------------------------------------------------------------------------------------------
// Helper Functions
// -----------------------------------------------------------------------------------------------

// A gzip-compressed representation of '{'
func fixtureInvalidEncodedPolicy() []byte {
	invalidEncoded := `H4sICLVr4GcAA2V4YW1wbGUuanNvbgCr5gIA3FvHbQIAAAA=`
	invalidDecoded, err := base64.StdEncoding.DecodeString(invalidEncoded)
	if err != nil {
		panic(err)
	}

	return invalidDecoded
}

// A gzip-compressed representation of '{}'
func fixtureInvalidEmptyPolicy() []byte {
	invalidEncoded := `H4sICPlu4GcAA2V4YW1wbGUuanNvbgCrruUCAAawod0DAAAA`
	invalidDecoded, err := base64.StdEncoding.DecodeString(invalidEncoded)
	if err != nil {
		panic(err)
	}

	return invalidDecoded
}
