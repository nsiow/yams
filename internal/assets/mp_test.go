package assets

import (
	"encoding/base64"
	"fmt"
	"testing"
)

// -----------------------------------------------------------------------------------------------
// Test Functions
// -----------------------------------------------------------------------------------------------

// TestValidPolicyDataLoad confirms the successful loading of valid data
func TestValidPolicyDataLoad(t *testing.T) {
	policies := ManagedPolicyData()
	if len(policies) < MINIMUM_POLICYSET_SIZE {
		t.Fatalf(fmt.Sprintf("expected >= %d policies but saw %d",
			MINIMUM_POLICYSET_SIZE,
			len(policies),
		))
	}
}

// TestInvalidPolicyGzip confirms the failed loading of corrupted gzip data
func TestInvalidPolicyGzip(t *testing.T) {
	// Assert panic
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic but observed success")
		} else {
			t.Logf("saw expected panic: %s", r.(string))
		}
	}()

	// Load invalid data
	loadManagedPolicyData([]byte{})
}

// TestInvalidPolicyDecode confirms the failed loading of corrupted JSON data
func TestInvalidPolicyDecode(t *testing.T) {
	// Assert panic
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic but observed success")
		} else {
			t.Logf("saw expected panic: %s", r.(string))
		}
	}()

	// Load invalid data
	loadManagedPolicyData(fixtureInvalidEncodedPolicy())

}

// TestInvalidPolicyEmpty confirms the failed loading of a too-short policy list
func TestInvalidPolicyEmpty(t *testing.T) {
	// Assert panic
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic but observed success")
		} else {
			t.Logf("saw expected panic: %s", r.(string))
		}
	}()

	// Load invalid data
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
