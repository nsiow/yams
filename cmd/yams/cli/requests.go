package cli

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"os"

	json "github.com/bytedance/sonic"
)

func GetReq(url string) {
	slog.Debug("making request", "url", url)
	resp, err := http.Get(url)
	if err != nil {
		Fail("error retrieving URL '%s': %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		Fail("received non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		Fail("error reading body of URL '%s': %v", url, err)
	}

	os.Stdout.Write(body)
}

func PostReq(url string, body any) {
	var buf bytes.Buffer
	err := json.ConfigDefault.NewEncoder(&buf).Encode(body)
	if err != nil {
		Fail("error encoding simulation input: %v", err)
	}

	resp, err := http.Post(url, "application/json", &buf)
	if err != nil {
		Fail("error hitting URL '%s': %v", url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		Fail("error reading body of URL '%s': %v", url, err)
	}

	os.Stdout.Write(respBody)
}
