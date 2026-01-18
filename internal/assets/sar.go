package assets

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"sync"

	json "github.com/bytedance/sonic"
	"github.com/nsiow/yams/pkg/aws/sar/types"
)

//go:embed sar.json.gz
var compressedSarData []byte
var sarData []types.Service
var sarIndex map[string]map[string]types.Action // map[service]map[action]action
var sarDataLoad sync.Once

// The minimum number of documented services expected; used to detect regressions
// Last updated 03-23-2025
var MINIMUM_SAR_SIZE = 415

// SAR loads the data if it has not been loaded, and returns the result
func SAR() []types.Service {
	sarDataLoad.Do(func() { loadSAR(compressedSarData) })
	return sarData
}

// SARIndex loads the data if it has not been loaded, and returns an indexed version
func SARIndex() map[string]map[string]types.Action {
	sarDataLoad.Do(func() { loadSAR(compressedSarData) })
	return sarIndex
}

// loadSAR processes the provided raw compressed data into the structured policy set
func loadSAR(compressedData []byte) {
	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		panic(fmt.Sprintf("error unwrapping SAR data: %s", err.Error()))
	}

	rawSarData, _ := io.ReadAll(reader)

	var newData []types.Service
	err = json.Unmarshal(rawSarData, &newData)
	if err != nil {
		panic(fmt.Sprintf("error decoding SAR data: %s", err.Error()))
	}

	// basic validation check for successful load
	if len(newData) < MINIMUM_SAR_SIZE {
		panic(fmt.Sprintf("error validating SAR data, len too small: %d",
			len(newData)))
	}

	// build the index, once
	newIndex := make(map[string]map[string]types.Action)
	for _, service := range newData {
		serviceName := strings.ToLower(service.Name)
		if _, exists := newIndex[service.Name]; !exists {
			newIndex[serviceName] = make(map[string]types.Action)
		}

		for _, action := range service.Actions {
			actionName := strings.ToLower(action.Name)
			newIndex[serviceName][actionName] = action
		}
	}

	sarData = newData
	sarIndex = newIndex
}
