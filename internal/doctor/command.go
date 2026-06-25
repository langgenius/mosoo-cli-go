package doctor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/langgenius/mosoo-cli-go/internal/target"
	latheconfig "github.com/lathe-cli/lathe/pkg/config"
	"github.com/spf13/cobra"
)

type Report struct {
	Target        string            `json:"target"`
	Source        string            `json:"source"`
	BaseURL       string            `json:"baseUrl"`
	Hosts         map[string]string `json:"hosts"`
	ConfigPath    string            `json:"configPath,omitempty"`
	ProjectRoot   string            `json:"projectRoot,omitempty"`
	AuthRequired  bool              `json:"authRequired"`
	Authenticated bool              `json:"authenticated"`
	Ready         bool              `json:"ready"`
	Checks        []Check           `json:"checks"`
}

type Check struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
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

	checks := []Check{
		{Name: "cli", OK: true},
		{Name: "target", OK: true, Message: fmt.Sprintf("%s from %s", resolved.Target, resolved.Source)},
	}

	apiCheck := checkAPI(cmd.Context(), resolved.Hosts[target.SurfaceConsole])
	checks = append(checks, apiCheck)

	authenticated, authRequired, authCheck := checkAuth(resolved)
	checks = append(checks, authCheck)

	ready := true
	for _, check := range checks {
		if !check.OK {
			ready = false
			break
		}
	}

	return Report{
		Target:        resolved.Target,
		Source:        resolved.Source,
		BaseURL:       resolved.BaseURL,
		Hosts:         resolved.Hosts,
		ConfigPath:    resolved.ConfigPath,
		ProjectRoot:   resolved.ProjectRoot,
		AuthRequired:  authRequired,
		Authenticated: authenticated,
		Ready:         ready,
		Checks:        checks,
	}, nil
}

func checkAPI(ctx context.Context, consoleHost string) Check {
	if consoleHost == "" {
		return Check{Name: "api", OK: false, Message: "console host is empty"}
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	endpoint := strings.TrimRight(consoleHost, "/") + "/access-tokens"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Check{Name: "api", OK: false, Message: err.Error()}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Check{Name: "api", OK: false, Message: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode >= http.StatusInternalServerError {
		return Check{Name: "api", OK: false, Message: fmt.Sprintf("GET %s returned %s", endpoint, resp.Status)}
	}
	return Check{Name: "api", OK: true, Message: fmt.Sprintf("GET %s returned %s", endpoint, resp.Status)}
}

func checkAuth(resolved target.Resolution) (bool, bool, Check) {
	authRequired := requiresAuth(resolved)
	if !authRequired {
		return false, false, Check{Name: "auth", OK: true, Message: "not required for local target"}
	}

	hosts, err := latheconfig.LoadHosts()
	if err != nil {
		return false, authRequired, Check{Name: "auth", OK: false, Message: err.Error()}
	}

	missing := make([]string, 0, 2)
	for _, surface := range []string{target.SurfaceConsole, target.SurfacePublicThreadAPI} {
		host := resolved.Hosts[surface]
		if host == "" {
			continue
		}
		if _, ok := hosts.Get(host); !ok {
			missing = append(missing, host)
		}
	}
	if len(missing) > 0 {
		return false, authRequired, Check{Name: "auth", OK: false, Message: "not authenticated to " + strings.Join(missing, ", ")}
	}
	return true, authRequired, Check{Name: "auth", OK: true}
}

func requiresAuth(resolved target.Resolution) bool {
	return resolved.Target != target.LocalTarget && !target.IsLocalBaseURL(resolved.BaseURL)
}

func printHuman(cmd *cobra.Command, report Report) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "target: %s\n", report.Target)
	fmt.Fprintf(out, "source: %s\n", report.Source)
	fmt.Fprintf(out, "baseUrl: %s\n", report.BaseURL)
	if report.ConfigPath != "" {
		fmt.Fprintf(out, "configPath: %s\n", report.ConfigPath)
	}
	if report.ProjectRoot != "" {
		fmt.Fprintf(out, "projectRoot: %s\n", report.ProjectRoot)
	}
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
