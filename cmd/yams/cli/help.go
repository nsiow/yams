package cli

import (
	"fmt"
	"os"
)

const banner = `
 _   _  ___  __  __ ___
| | | |/   \|  \/  / __|
| |_| | |_| | |\/| \__ \
 \__, |\___/|_|  |_|___/
 |___/
`

// CommandInfo describes a CLI command
type CommandInfo struct {
	Name        string
	Description string
	Aliases     []string
}

var Commands = []CommandInfo{
	{Name: "status", Description: "Show server status and loaded data sources"},
	{Name: "server", Description: "Start the yams API server"},
	{Name: "dump", Description: "Export AWS organization or config data"},
	{Name: "sim", Description: "Simulate IAM permission checks"},
	{Name: "principals", Description: "List or search IAM principals (roles, users)", Aliases: []string{"p"}},
	{Name: "resources", Description: "List or search AWS resources", Aliases: []string{"r"}},
	{Name: "actions", Description: "List or search IAM actions", Aliases: []string{"a"}},
	{Name: "accounts", Description: "List or search AWS accounts", Aliases: []string{"acc"}},
	{Name: "policies", Description: "List or search IAM policies", Aliases: []string{"pol"}},
}

// PrintHelp prints the main help message
func PrintHelp() {
	if StderrIsTTY() {
		fmt.Fprint(os.Stderr, banner)
	}
	fmt.Fprintln(os.Stderr, "Usage: yams <command> [options]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Commands:")

	// Calculate max command name width for alignment
	maxWidth := 0
	for _, cmd := range Commands {
		if len(cmd.Name) > maxWidth {
			maxWidth = len(cmd.Name)
		}
	}

	for _, cmd := range Commands {
		aliasStr := ""
		if len(cmd.Aliases) > 0 {
			aliasStr = fmt.Sprintf(" (alias: %s)", cmd.Aliases[0])
		}
		fmt.Fprintf(os.Stderr, "  %-*s  %s%s\n", maxWidth, cmd.Name, cmd.Description, aliasStr)
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Global options:")
	fmt.Fprintln(os.Stderr, "  -h, --help       Show this help message")
	fmt.Fprintln(os.Stderr, "  -v, --version    Show version information")
	fmt.Fprintln(os.Stderr, "  -V, --verbose    Enable debug logging")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Use 'yams <command> -h' for help on a specific command.")
}

// ResolveAlias returns the canonical command name for an alias
func ResolveAlias(cmd string) string {
	for _, c := range Commands {
		if c.Name == cmd {
			return cmd
		}
		for _, alias := range c.Aliases {
			if alias == cmd {
				return c.Name
			}
		}
	}
	return cmd
}
