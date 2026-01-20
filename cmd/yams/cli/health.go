package cli

import (
	"log/slog"
	"net/http"
	"time"
)

// CheckServerHealth verifies that the yams server is reachable
// Returns an error description if the server is not reachable
func CheckServerHealth(server string) error {
	url := ApiUrl(server, "status")
	slog.Debug("checking server health", "url", url)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &serverError{statusCode: resp.StatusCode}
	}

	return nil
}

type serverError struct {
	statusCode int
}

func (e *serverError) Error() string {
	return "server returned non-200 status"
}

// RequireServer checks that the server is reachable, failing with a helpful message if not
func RequireServer(server string) {
	if err := CheckServerHealth(server); err != nil {
		Fail("cannot connect to yams server at '%s': %v", server, err)
	}
}
