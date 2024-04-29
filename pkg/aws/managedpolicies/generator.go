//go:build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/nsiow/yams/pkg/policy"
)

const (
	MANAGED_POLICY_TEMPLATE_FILE = "managed_policies.go.template"
)

func main() {
	// Print some debugging
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error retrieving working directory: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("-> running code generation for: managed policies\n")
	fmt.Printf("-> args = %v\n", os.Args)
	fmt.Printf("-> cwd = %s\n", cwd)
	var govars []string
	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "GO") {
			govars = append(govars, v)
		}
	}
	fmt.Printf("-> env = %v\n", govars)

	// Parse arguments
	datafile := flag.String("data", "", "path to data file")
	flag.Parse()

	// Read data file
	data, err := os.ReadFile(*datafile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading from file '%s': %v\n", *datafile, err)
		os.Exit(1)
	}

	// Construct policies
	var policies []ManagedPolicyEntry
	err = json.Unmarshal(data, &policies)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing managed policies: %v\n", err)
		os.Exit(1)
	}

	// Generated managed policies
	data, err = os.ReadFile(MANAGED_POLICY_TEMPLATE_FILE)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading template file '%s': %v\n", MANAGED_POLICY_TEMPLATE_FILE, err)
		os.Exit(1)
	}
	tmpl, err := template.New("mp").Parse(string(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing template file '%s': %v\n", MANAGED_POLICY_TEMPLATE_FILE, err)
		os.Exit(1)
	}
	for _, policy := range policies {
		fn := fmt.Sprintf("%s.go", policy.EscapedName())
		f, err := os.Create(fn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening file for writing '%s': %v\n", fn, err)
			os.Exit(1)
		}
		err = tmpl.Execute(f, policies)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error rendering template file '%s': %v\n", MANAGED_POLICY_TEMPLATE_FILE, err)
			os.Exit(1)
		}
		fmt.Printf("-> successfully rendered template to '%s'\n", fn)

	}

	fmt.Printf("-> generated code from %d managed policies\n", len(policies))
}

type ManagedPolicyEntry struct {
	Arn      string        `json:"arn"`
	Document policy.Policy `json:"document"`
}

func (m *ManagedPolicyEntry) EscapedName() string {
	fragments := strings.Split(m.Arn, "/")
	name := fragments[len(fragments)-1]
	return strings.ReplaceAll(name, "-", "_")
}

func (m *ManagedPolicyEntry) VarName() string {
	return fmt.Sprintf("AWS_MANAGED_POLICY_%s", m.EscapedName())
}

func (m *ManagedPolicyEntry) Struct() string {
	return fmt.Sprintf("%#v", m.Document)
}
