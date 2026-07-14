package cli

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/opp-13/nhn_iac/resource_checker/config"
	"github.com/opp-13/nhn_iac/resource_checker/output"
)

// maskedPassword is what "configure show" always prints for the password
// field, matching resource_checker/README.md's documented behavior exactly.
const maskedPassword = "******"

func runConfigure(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, configureHelp)
		return 0
	}

	switch args[0] {
	case "-h", "--help":
		fmt.Fprint(stdout, configureHelp)
		return 0
	case "set":
		return runConfigureSet(args[1:], stdout, stderr)
	case "show":
		return runConfigureShow(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "rescheck configure: 알 수 없는 명령어 %q\n\n", args[0])
		fmt.Fprint(stderr, configureHelp)
		return 1
	}
}

func runConfigureSet(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("configure set", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var tenantID, region, username, password, configPath string
	fs.StringVar(&tenantID, "tenant-id", "", "")
	fs.StringVar(&region, "region", "", "")
	fs.StringVar(&username, "username", "", "")
	fs.StringVar(&password, "password", "", "")
	fs.StringVar(&configPath, "config", config.FindConfigPath(), "")

	if err := fs.Parse(args); err != nil {
		return helpOrError(err, stdout, stderr, configureSetHelp)
	}

	configPathProvided, tenantIDProvided, regionProvided, usernameProvided, passwordProvided := false, false, false, false, false
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "config":
			configPathProvided = true
		case "tenant-id":
			tenantIDProvided = true
		case "region":
			regionProvided = true
		case "username":
			usernameProvided = true
		case "password":
			passwordProvided = true
		}
	})

	// One shared reader across every prompt: bufio.Reader reads ahead, so a
	// fresh bufio.NewReader(os.Stdin) per prompt would swallow the next
	// line's input into a buffer that gets discarded when that call returns.
	stdin := bufio.NewReader(os.Stdin)

	if !configPathProvided {
		// configPath already holds config.FindConfigPath()'s result (the
		// flag's default); an empty answer here just keeps it, so pressing
		// Enter means "use the default" rather than "no config file".
		line, err := promptLine(stdout, stdin, fmt.Sprintf("Config file path (default: %s): ", configPath))
		if err != nil {
			fmt.Fprintf(stderr, "rescheck configure set: config 경로를 입력받을 수 없습니다 (--config를 직접 전달하세요): %v\n", err)
			return 1
		}
		if line != "" {
			configPath = line
		}
	}

	if !tenantIDProvided {
		line, err := promptLine(stdout, stdin, "Tenant ID: ")
		if err != nil {
			fmt.Fprintf(stderr, "rescheck configure set: tenant ID를 입력받을 수 없습니다 (--tenant-id를 직접 전달하세요): %v\n", err)
			return 1
		}
		tenantID = line
	}
	if !usernameProvided {
		line, err := promptLine(stdout, stdin, "Username: ")
		if err != nil {
			fmt.Fprintf(stderr, "rescheck configure set: username을 입력받을 수 없습니다 (--username을 직접 전달하세요): %v\n", err)
			return 1
		}
		username = line
	}

	if tenantID == "" || username == "" {
		fmt.Fprintln(stderr, "rescheck configure set: --tenant-id와 --username은 필수입니다")
		fmt.Fprint(stderr, configureSetHelp)
		return 1
	}

	if !passwordProvided {
		fmt.Fprint(stdout, "Password: ")
		bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(stdout)
		if err != nil {
			fmt.Fprintf(stderr, "rescheck configure set: 비밀번호를 입력받을 수 없습니다 (터미널이 아닌 경우 --password를 직접 전달하세요): %v\n", err)
			return 1
		}
		password = string(bytes)
	}

	cfg, raw, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if !regionProvided {
		// Keep whatever region is already saved (or the default, if this is
		// the first save) so a credential-only re-run doesn't silently
		// revert or drop a previously chosen region.
		region = cfg.Nhn.Auth.Region
	}
	if err := config.SaveAuth(configPath, raw, tenantID, region, username, password); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	fmt.Fprintf(stdout, "credentials saved to %s\n", configPath)
	return 0
}

func runConfigureShow(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("configure show", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var format, configPath string
	fs.StringVar(&format, "o", output.FormatTable, "")
	fs.StringVar(&format, "output", output.FormatTable, "")
	fs.StringVar(&configPath, "config", config.FindConfigPath(), "")

	if err := fs.Parse(args); err != nil {
		return helpOrError(err, stdout, stderr, configureShowHelp)
	}
	if !output.Valid(format) {
		fmt.Fprintf(stderr, "rescheck configure show: 알 수 없는 --output %q\n", format)
		return 1
	}

	cfg, _, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	if format == output.FormatJSON {
		view := struct {
			TenantID               string `json:"tenant_id"`
			Region                 string `json:"region"`
			Username               string `json:"username"`
			Password               string `json:"password"`
			ResourcecheckerEnabled bool   `json:"resourcechecker_enabled"`
			ResourcecheckerMode    string `json:"resourcechecker_mode"`
		}{
			TenantID:               cfg.Nhn.Auth.TenantID,
			Region:                 cfg.Nhn.Auth.Region,
			Username:               cfg.Nhn.Auth.Username,
			Password:               maskedPassword,
			ResourcecheckerEnabled: cfg.Nhn.Resourcechecker.Enabled,
			ResourcecheckerMode:    cfg.Nhn.Resourcechecker.Mode,
		}
		if err := output.RenderJSON(stdout, view); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	}

	headers := []string{"TENANT ID", "REGION", "USERNAME", "PASSWORD", "RESOURCECHECKER ENABLED", "MODE"}
	row := []string{
		cfg.Nhn.Auth.TenantID,
		cfg.Nhn.Auth.Region,
		cfg.Nhn.Auth.Username,
		maskedPassword,
		fmt.Sprintf("%t", cfg.Nhn.Resourcechecker.Enabled),
		cfg.Nhn.Resourcechecker.Mode,
	}
	if err := output.Render(stdout, format, headers, [][]string{row}); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

// promptLine writes label to stdout and reads back one line from stdin via
// the caller-owned reader, trimmed. Used for --tenant-id/--username when
// omitted; unlike the password prompt, input is shown as typed since these
// aren't secrets. Callers making multiple prompts must reuse the same
// *bufio.Reader (see runConfigureSet) so read-ahead buffering doesn't
// swallow later lines.
func promptLine(stdout io.Writer, stdin *bufio.Reader, label string) (string, error) {
	fmt.Fprint(stdout, label)
	line, err := stdin.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// helpOrError prints helpText and returns 0 when err is flag.ErrHelp
// (triggered by -h/--help), otherwise prints the parse error plus helpText
// to stderr and returns 1.
func helpOrError(err error, stdout, stderr io.Writer, helpText string) int {
	if errors.Is(err, flag.ErrHelp) {
		fmt.Fprint(stdout, helpText)
		return 0
	}
	fmt.Fprintln(stderr, err)
	fmt.Fprint(stderr, helpText)
	return 1
}
