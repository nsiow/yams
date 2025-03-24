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
)

//go:embed sar.json.gz
var compressedSarData []byte
var sarData map[string][]entities.ApiCall
var sarDataLoad sync.Once

// The minimum number of API calls expected; used to detect regressions
// Last updated 03-23-2025
var MINIMUM_SAR_SIZE = 415

// SarData loads the data if it has not been loaded, and returns the result
func SarData() map[string][]entities.ApiCall {
	sarDataLoad.Do(func() { loadSarData(compressedSarData) })
	return sarData
}

// loadSarData processes the provided raw compressed data into the structured policy set
func loadSarData(compressedData []byte) {
	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		panic(fmt.Sprintf("error unwrapping SAR data: %s", err.Error()))
	}

	rawSarData, _ := io.ReadAll(reader)

	var newData map[string][]entities.ApiCall
	err = json.Unmarshal(rawSarData, &newData)
	if err != nil {
		panic(fmt.Sprintf("error decoding SAR data: %s", err.Error()))
	}

	// basic validation check for successful load
	if len(newData) < MINIMUM_SAR_SIZE {
		panic(fmt.Sprintf("error validating SAR data, len too small: %d",
			len(newData)))
	}

	sarData = newData
}
