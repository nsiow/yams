package assets

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// -----------------------------------------------------------------------------------------------
// MANAGED POLICIES
// -----------------------------------------------------------------------------------------------

//go:embed mp.json.gz
var compressedManagedPolicyData []byte
var managedPolicyData map[string]policy.Policy
var managedPolicyDataLoad sync.Once

func ManagedPolicyData() map[string]policy.Policy {
	managedPolicyDataLoad.Do(func() {
		reader, err := gzip.NewReader(bytes.NewReader(compressedManagedPolicyData))
		if err != nil {
			panic(fmt.Sprintf("error unwrapping managed policy data: %s", err.Error()))
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

// -----------------------------------------------------------------------------------------------
// SERVICE AUTHORIZATION REFERENCE
// -----------------------------------------------------------------------------------------------

//go:embed sar.json.gz
var compressedSarData []byte
var sarData map[string][]entities.ApiCall
var sarDataLoad sync.Once

func SarData() map[string][]entities.ApiCall {
	sarDataLoad.Do(func() {
		reader, err := gzip.NewReader(bytes.NewReader(compressedSarData))
		if err != nil {
			panic(fmt.Sprintf("error unwrapping SAR data: %s", err.Error()))
		}

		rawSarData, err := io.ReadAll(reader)
		if err != nil {
			panic(fmt.Sprintf("error decompressing SAR data: %s", err.Error()))
		}

		err = json.Unmarshal(rawSarData, &sarData)
		if err != nil {
			panic(fmt.Sprintf("error decoding SAR data: %s", err.Error()))
		}
	})

	return sarData
}
