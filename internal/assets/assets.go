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

func ManagedPolicyData() map[string]policy.Policy {
	managedPolicyDataLoad.Do(func() {
		reader, err := gzip.NewReader(bytes.NewReader(compressedManagedPolicyData))
		if err != nil {
			panic(fmt.Sprintf("error wrapping managed policy data: %s", err.Error()))
		}

		rawManagedPolicyData, err := io.ReadAll(reader)
		if err != nil {
			panic(fmt.Sprintf("error decompressing managed policy data: %s", err.Error()))
		}

		err = json.Unmarshal(rawManagedPolicyData, &managedPolicyData)
		if err != nil {
			panic(fmt.Sprintf("error decoding managed policy data: %s", err.Error()))
		}
	})

	return managedPolicyData
}
