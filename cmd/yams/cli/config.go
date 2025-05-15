package cli

import (
	urllib "net/url"
	"strings"
)

func ApiUrl(addr string, elem ...string) string {
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}

	url, err := urllib.JoinPath(addr+"/api/v1/", elem...)
	if err != nil {
		panic(err.Error())
	}

	return url
}
