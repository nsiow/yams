package assets

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/nsiow/yams/pkg/policy"
)

//go:embed mp.json.gz
var compressedManagedPolicyData []byte
var managedPolicyData map[string]policy.Policy
var managedPolicyDataLoad sync.Once

// The minimum number of policies expected; used to detect regressions
// Last updated 03-23-2025
var MINIMUM_POLICYSET_SIZE = 1335

// ManagedPolicyData loads the data if it has not been loaded, and returns the result
func ManagedPolicyData() map[string]policy.Policy {
	managedPolicyDataLoad.Do(func() { loadManagedPolicyData(compressedManagedPolicyData) })
	return managedPolicyData
}

// loadManagedPolicyData processes the provided raw compressed data into the structured policy set
func loadManagedPolicyData(compressedData []byte) {
	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		panic(fmt.Sprintf("error unwrapping managed policy data: %s", err.Error()))
	}

	rawManagedPolicyData, _ := io.ReadAll(reader)

	var newData map[string]policy.Policy
	err = json.Unmarshal(rawManagedPolicyData, &newData)
	if err != nil {
		panic(fmt.Sprintf("error decoding managed policy data: %s", err.Error()))
	}

	// basic validation check for successful load
	if len(newData) < MINIMUM_POLICYSET_SIZE {
		panic(fmt.Sprintf("error validating managed policy data, len too small: %d",
			len(managedPolicyData)))
	}

	managedPolicyData = newData
}
