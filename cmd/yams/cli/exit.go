package cli

import (
	"fmt"
	"os"

	json "github.com/bytedance/sonic"
)

func Fail(format string, a ...any) {
	errblob, err := json.MarshalIndent(
		map[string]string{
			"error": fmt.Sprintf(format, a...),
		},
		"",
		"  ",
	)
	if err != nil {
		panic(err) // should never really get here
	}

	fmt.Fprintf(os.Stderr, "%s\n", errblob)
	os.Exit(2)
}
