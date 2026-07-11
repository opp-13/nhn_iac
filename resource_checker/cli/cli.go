// Package cli implements the rescheck command-line dispatch: configure,
// compute, and network, plus their subcommands. Every --help text here is
// kept in sync with resource_checker/README.md's "## CLI Mode" section
// (see help.go).
package cli

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// timeLayout formats timestamps for table/text output.
const timeLayout = "2006-01-02T15:04:05Z"

// formatTime renders t for table/text display, or "-" if t is the zero
// value (some NHN Cloud/gophercloud resources don't populate a creation
// timestamp — see the comments in nhncloud/networks.go and
// nhncloud/floatingips.go).
func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format(timeLayout)
}

// expandShortBoolFlags rewrites combined short boolean flags like "-al" into
// separate "-a" "-l" tokens, since the stdlib flag package doesn't support
// bundled short options. Only a token made up entirely of single-letter
// flags in shortBoolFlags gets expanded (e.g. shortBoolFlags="al" matches
// "-a", "-l", "-al", "-la" but not "-o" or "-ao" — those pass through
// unchanged and fall to flag.Parse's normal "not defined" error). "--",
// "--long"-style double-dash flags, and anything after a bare "--"
// separator are left untouched.
func expandShortBoolFlags(args []string, shortBoolFlags string) []string {
	expanded := make([]string, 0, len(args))
	for i, arg := range args {
		if arg == "--" {
			expanded = append(expanded, args[i:]...)
			break
		}
		if len(arg) > 2 && arg[0] == '-' && arg[1] != '-' && onlyContains(arg[1:], shortBoolFlags) {
			for _, c := range arg[1:] {
				expanded = append(expanded, "-"+string(c))
			}
			continue
		}
		expanded = append(expanded, arg)
	}
	return expanded
}

func onlyContains(s, allowed string) bool {
	for _, c := range s {
		if !strings.ContainsRune(allowed, c) {
			return false
		}
	}
	return true
}

// Run dispatches args (os.Args[1:]) and returns the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, rootHelp)
		return 0
	}

	switch args[0] {
	case "-h", "--help", "help":
		fmt.Fprint(stdout, rootHelp)
		return 0
	case "-v", "--version":
		fmt.Fprintf(stdout, "rescheck %s\n", version)
		return 0
	case "configure", "conf":
		return runConfigure(args[1:], stdout, stderr)
	case "compute", "com":
		return runCompute(args[1:], stdout, stderr)
	case "network", "net":
		return runNetwork(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "rescheck: 알 수 없는 명령어 %q\n\n", args[0])
		fmt.Fprint(stderr, rootHelp)
		return 1
	}
}
