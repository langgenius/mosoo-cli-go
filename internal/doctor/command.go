package doctor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/langgenius/mosoo-cli-go/internal/buildinfo"
	"github.com/langgenius/mosoo-cli-go/internal/target"
	latheconfig "github.com/lathe-cli/lathe/pkg/config"
	"github.com/spf13/cobra"
)

type Report struct {
	SchemaVersion int            `json:"schemaVersion"`
	Ready         bool           `json:"ready"`
	Target        target.State   `json:"target"`
	Auth          AuthState      `json:"auth"`
	Install       buildinfo.Info `json:"install"`
	Failures      []Failure      `json:"failures"`
	Checks        []Check        `json:"checks"`
}

type AuthState struct {
	Required        bool     `json:"required"`
	Authenticated   bool     `json:"authenticated"`
	CredentialHosts []string `json:"credentialHosts"`
	MissingHosts    []string `json:"missingHosts"`
}

type Failure struct {
	Check   string `json:"check"`
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
	Action  string `json:"action"`
}

type Check struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

func NewCommand() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check Mosoo CLI target resolution and readiness",
		RunE: func(cmd *cobra.Command, _ []string) error {
			report, err := BuildReport(cmd)
			if err != nil {
				return err
			}
			outputFormat, _ := cmd.Root().PersistentFlags().GetString("output")
			if jsonOutput || outputFormat == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(report)
			}
			printHuman(cmd, report)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print machine-readable JSON")
	return cmd
}

func BuildReport(cmd *cobra.Command) (Report, error) {
	resolved, err := target.ResolveFromCommand(cmd)
	if err != nil {
		return Report{}, err
	}

	apiCheck := checkAPI(cmd.Context(), resolved.Hosts[target.SurfaceConsole])
	auth, authCheck := evaluateAuth(resolved)

	return newReport(resolved, apiCheck, auth, authCheck), nil
}

func NewReport(resolved target.Resolution, apiCheck Check, auth AuthState) Report {
	return newReport(resolved, apiCheck, auth, checkFromAuthState(auth))
}

func newReport(resolved target.Resolution, apiCheck Check, auth AuthState, authCheck Check) Report {
	install := buildinfo.Current()
	checks := []Check{
		{Name: "cli", OK: true, Code: "cli_available"},
		{Name: "target", OK: true, Code: "target_resolved", Message: fmt.Sprintf("%s from %s", resolved.Target, resolved.Source)},
		apiCheck,
		authCheck,
		checkFromInstallState(install),
	}

	return Report{
		SchemaVersion: 1,
		Ready:         checksReady(checks),
		Target:        target.StateFromResolution(resolved),
		Auth:          auth,
		Install:       install,
		Failures:      failuresForChecks(checks),
		Checks:        checks,
	}
}

func checkAPI(ctx context.Context, consoleHost string) Check {
	if consoleHost == "" {
		return Check{Name: "api", OK: false, Code: "api_console_host_empty", Message: "console host is empty"}
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	endpoint := strings.TrimRight(consoleHost, "/") + "/access-tokens"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Check{Name: "api", OK: false, Code: "api_request_invalid", Message: err.Error()}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Check{Name: "api", OK: false, Code: "api_unreachable", Message: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode >= http.StatusInternalServerError {
		return Check{Name: "api", OK: false, Code: "api_unready_status", Message: fmt.Sprintf("GET %s returned %s", endpoint, resp.Status)}
	}
	return Check{Name: "api", OK: true, Code: "api_reachable", Message: fmt.Sprintf("GET %s returned %s", endpoint, resp.Status)}
}

func checkAuth(resolved target.Resolution) (bool, bool, Check) {
	auth, check := evaluateAuth(resolved)
	return auth.Authenticated, auth.Required, check
}

func evaluateAuth(resolved target.Resolution) (AuthState, Check) {
	authRequired := requiresAuth(resolved)
	auth := AuthState{
		Required:        authRequired,
		Authenticated:   false,
		CredentialHosts: []string{},
		MissingHosts:    []string{},
	}
	if !authRequired {
		return auth, Check{Name: "auth", OK: true, Code: "auth_not_required", Message: "not required for local target"}
	}

	candidateHosts := authCandidateHosts(resolved)
	hosts, err := latheconfig.LoadHosts()
	if err != nil {
		auth.MissingHosts = candidateHosts
		return auth, Check{Name: "auth", OK: false, Code: "auth_store_unavailable", Message: err.Error()}
	}

	for _, host := range candidateHosts {
		if _, ok := hosts.Get(host); !ok {
			auth.MissingHosts = append(auth.MissingHosts, host)
			continue
		}
		auth.CredentialHosts = append(auth.CredentialHosts, host)
	}
	if len(auth.MissingHosts) > 0 {
		return auth, Check{Name: "auth", OK: false, Code: "auth_missing_credentials", Message: "not authenticated to " + strings.Join(auth.MissingHosts, ", ")}
	}
	auth.Authenticated = true
	return auth, Check{Name: "auth", OK: true, Code: "auth_credentials_present"}
}

func requiresAuth(resolved target.Resolution) bool {
	return resolved.Target != target.LocalTarget && !target.IsLocalBaseURL(resolved.BaseURL)
}

func authCandidateHosts(resolved target.Resolution) []string {
	seen := map[string]bool{}
	hosts := make([]string, 0, 2)
	for _, surface := range []string{target.SurfaceConsole, target.SurfacePublicThreadAPI} {
		host := resolved.Hosts[surface]
		if host == "" || seen[host] {
			continue
		}
		seen[host] = true
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)
	return hosts
}

func checkFromAuthState(auth AuthState) Check {
	if !auth.Required {
		return Check{Name: "auth", OK: true, Code: "auth_not_required", Message: "not required for local target"}
	}
	if auth.Authenticated {
		return Check{Name: "auth", OK: true, Code: "auth_credentials_present"}
	}
	if len(auth.MissingHosts) > 0 {
		return Check{Name: "auth", OK: false, Code: "auth_missing_credentials", Message: "not authenticated to " + strings.Join(auth.MissingHosts, ", ")}
	}
	return Check{Name: "auth", OK: false, Code: "auth_missing_credentials", Message: "authentication is required but no credential hosts were found"}
}

func checkFromInstallState(install buildinfo.Info) Check {
	if install.Complete {
		return Check{Name: "install", OK: true, Code: "build_metadata_present", Message: fmt.Sprintf("%s (%s, %s)", install.Version, install.Commit, install.Date)}
	}
	return Check{Name: "install", OK: false, Code: "build_metadata_missing", Message: fmt.Sprintf("%s (%s, %s)", install.Version, install.Commit, install.Date)}
}

func checksReady(checks []Check) bool {
	for _, check := range checks {
		if !check.OK {
			return false
		}
	}
	return true
}

func failuresForChecks(checks []Check) []Failure {
	failures := make([]Failure, 0)
	for _, check := range checks {
		if check.OK {
			continue
		}
		failures = append(failures, Failure{
			Check:   check.Name,
			Code:    check.Code,
			Message: check.Message,
			Action:  actionForCode(check.Code),
		})
	}
	return failures
}

func actionForCode(code string) string {
	switch code {
	case "api_console_host_empty":
		return "Choose a target or base URL that resolves a console API host."
	case "api_request_invalid":
		return "Check the configured target base URL."
	case "api_unreachable":
		return "Start the Mosoo service or choose a reachable target with --target or --base-url."
	case "api_unready_status":
		return "Verify the selected Mosoo API is healthy before using generated API commands."
	case "auth_store_unavailable":
		return "Check the local Mosoo credential store and retry auth login."
	case "auth_missing_credentials":
		return "Run mosoo auth login for every host listed in auth.missingHosts."
	case "build_metadata_missing":
		return "Build the CLI through the Makefile or inject Lathe Version, Commit, and Date with Go ldflags."
	default:
		return "Inspect the failed check message and correct the environment before retrying."
	}
}

func printHuman(cmd *cobra.Command, report Report) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "target: %s\n", report.Target.Name)
	fmt.Fprintf(out, "source: %s\n", report.Target.Source)
	fmt.Fprintf(out, "baseUrl: %s\n", report.Target.BaseURL)
	if report.Target.ConfigPath != "" {
		fmt.Fprintf(out, "configPath: %s\n", report.Target.ConfigPath)
	}
	if report.Target.ProjectRoot != "" {
		fmt.Fprintf(out, "projectRoot: %s\n", report.Target.ProjectRoot)
	}
	fmt.Fprintf(out, "authRequired: %t\n", report.Auth.Required)
	fmt.Fprintf(out, "authenticated: %t\n", report.Auth.Authenticated)
	fmt.Fprintf(out, "install: %s (%s, %s)\n", report.Install.Version, report.Install.Commit, report.Install.Date)
	fmt.Fprintf(out, "ready: %t\n", report.Ready)
	fmt.Fprintln(out, "checks:")
	for _, check := range report.Checks {
		status := "ok"
		if !check.OK {
			status = "fail"
		}
		if check.Message == "" {
			fmt.Fprintf(out, "  [%s] %s\n", status, check.Name)
		} else {
			fmt.Fprintf(out, "  [%s] %s: %s\n", status, check.Name, check.Message)
		}
	}
}
