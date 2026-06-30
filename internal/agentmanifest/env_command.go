package agentmanifest

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/langgenius/mosoo-connector/internal/target"
	latheconfig "github.com/lathe-cli/lathe/pkg/config"
	"github.com/spf13/cobra"
)

const (
	apiBaseEnv  = "MOSOO_API_BASE"
	apiTokenEnv = "MOSOO_API_TOKEN"
	agentIDEnv  = "MOSOO_AGENT_ID"
)

type envCommandOptions struct {
	apiBase  string
	apiToken string
	agentID  string
	file     string
	json     bool
}

type envValues struct {
	APIBase  string
	APIToken string
	AgentID  string
}

type envResult struct {
	File     string `json:"file,omitempty"`
	APIBase  string `json:"apiBase"`
	AgentID  string `json:"agentId"`
	APIToken string `json:"apiToken"`
}

func newEnvCommand() *cobra.Command {
	envCmd := &cobra.Command{
		Use:   "env",
		Short: "Export or write public Agent API environment values",
		Long:  "Export or write MOSOO_API_BASE, MOSOO_AGENT_ID, and MOSOO_API_TOKEN for backend and Worker integrations. If no token is provided, the logged-in Public API host token is used. Token values are redacted in terminal output.",
	}

	exportOpts := &envCommandOptions{}
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Print redacted shell exports or write raw exports to a file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			values, err := resolveEnvValues(cmd, exportOpts)
			if err != nil {
				return err
			}
			if exportOpts.file != "" {
				if err := writeShellExportFile(exportOpts.file, values); err != nil {
					return err
				}
				return printEnvResult(cmd, exportOpts, values, exportOpts.file)
			}
			if wantsEnvJSON(cmd, exportOpts) {
				return printEnvResult(cmd, exportOpts, values, "")
			}
			return printRedactedShellExports(cmd, values)
		},
	}
	addEnvValueFlags(exportCmd, exportOpts)
	exportCmd.Flags().StringVar(&exportOpts.file, "file", "", "Write raw shell exports to this file instead of printing redacted exports")

	writeOpts := &envCommandOptions{}
	writeCmd := &cobra.Command{
		Use:   "write",
		Short: "Write public Agent API values to a dotenv file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(writeOpts.file) == "" {
				return fmt.Errorf("--file is required")
			}
			values, err := resolveEnvValues(cmd, writeOpts)
			if err != nil {
				return err
			}
			if err := writeDotenvFile(writeOpts.file, values); err != nil {
				return err
			}
			return printEnvResult(cmd, writeOpts, values, writeOpts.file)
		},
	}
	addEnvValueFlags(writeCmd, writeOpts)
	writeCmd.Flags().StringVar(&writeOpts.file, "file", "", "Dotenv file to create or update")

	envCmd.AddCommand(exportCmd, writeCmd)
	return envCmd
}

func addEnvValueFlags(cmd *cobra.Command, opts *envCommandOptions) {
	cmd.Flags().StringVar(&opts.apiBase, "api-base", "", "Public API base URL (defaults to MOSOO_API_BASE or the resolved target public API host)")
	cmd.Flags().StringVar(&opts.agentID, "agent-id", "", "Published Mosoo Agent ID (defaults to MOSOO_AGENT_ID)")
	cmd.Flags().StringVar(&opts.apiToken, "api-token", "", "Mosoo API token (defaults to MOSOO_API_TOKEN or the logged-in Public API host token)")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Print machine-readable JSON")
}

func resolveEnvValues(cmd *cobra.Command, opts *envCommandOptions) (envValues, error) {
	values := envValues{
		APIBase:  firstNonEmpty(opts.apiBase, os.Getenv(apiBaseEnv)),
		AgentID:  firstNonEmpty(opts.agentID, os.Getenv(agentIDEnv)),
		APIToken: firstNonEmpty(opts.apiToken, os.Getenv(apiTokenEnv)),
	}
	if values.APIBase == "" {
		resolved, err := target.ResolveFromCommand(cmd)
		if err != nil {
			return envValues{}, err
		}
		values.APIBase = resolved.Hosts[target.SurfacePublicThreadAPI]
	}
	if values.APIToken == "" {
		token, err := apiTokenFromAuthStore(values.APIBase)
		if err != nil {
			return envValues{}, err
		}
		values.APIToken = token
	}
	if err := validateEnvValues(values); err != nil {
		return envValues{}, err
	}
	return values, nil
}

func apiTokenFromAuthStore(apiBase string) (string, error) {
	hosts, err := latheconfig.LoadHosts()
	if err != nil {
		return "", fmt.Errorf("load Mosoo auth hosts: %w", err)
	}
	entry, ok := hosts.Get(apiBase)
	if !ok {
		return "", nil
	}
	switch strings.TrimSpace(entry.AuthType) {
	case "", "bearer":
		return strings.TrimSpace(entry.OAuthToken), nil
	default:
		return "", nil
	}
}

func validateEnvValues(values envValues) error {
	missing := make([]string, 0, 2)
	if values.AgentID == "" {
		missing = append(missing, "--agent-id or "+agentIDEnv)
	}
	if values.APIToken == "" {
		missing = append(missing, "--api-token or "+apiTokenEnv)
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing %s", strings.Join(missing, " and "))
	}
	for name, value := range map[string]string{
		apiBaseEnv:  values.APIBase,
		agentIDEnv:  values.AgentID,
		apiTokenEnv: values.APIToken,
	} {
		if strings.ContainsAny(value, "\r\n") {
			return fmt.Errorf("%s must not contain newlines", name)
		}
	}
	return nil
}

func printRedactedShellExports(cmd *cobra.Command, values envValues) error {
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "export %s=%s\nexport %s=%s\nexport %s=%s\n",
		apiBaseEnv, shellQuote(values.APIBase),
		agentIDEnv, shellQuote(values.AgentID),
		apiTokenEnv, shellQuote(redactToken(values.APIToken)),
	)
	return err
}

func printEnvResult(cmd *cobra.Command, opts *envCommandOptions, values envValues, file string) error {
	result := envResult{
		File:     file,
		APIBase:  values.APIBase,
		AgentID:  values.AgentID,
		APIToken: redactToken(values.APIToken),
	}
	if wantsEnvJSON(cmd, opts) {
		return writeJSON(cmd, result)
	}
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "wrote %s\n", file)
	fmt.Fprintf(out, "%s=%s\n", apiBaseEnv, values.APIBase)
	fmt.Fprintf(out, "%s=%s\n", agentIDEnv, values.AgentID)
	fmt.Fprintf(out, "%s=%s\n", apiTokenEnv, result.APIToken)
	return nil
}

func wantsEnvJSON(cmd *cobra.Command, opts *envCommandOptions) bool {
	if opts.json {
		return true
	}
	if flag := cmd.Root().PersistentFlags().Lookup("output"); flag != nil {
		format, _ := cmd.Root().PersistentFlags().GetString("output")
		return format == "json"
	}
	return false
}

func writeDotenvFile(path string, values envValues) error {
	return writeEnvFile(path, []string{
		formatDotenvLine(apiBaseEnv, values.APIBase),
		formatDotenvLine(agentIDEnv, values.AgentID),
		formatDotenvLine(apiTokenEnv, values.APIToken),
	})
}

func writeShellExportFile(path string, values envValues) error {
	return writeEnvFile(path, []string{
		formatExportLine(apiBaseEnv, values.APIBase),
		formatExportLine(agentIDEnv, values.AgentID),
		formatExportLine(apiTokenEnv, values.APIToken),
	})
}

func writeEnvFile(path string, replacementLines []string) error {
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", path, err)
	}
	lines := mergeEnvLines(string(existing), replacementLines)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func mergeEnvLines(existing string, replacementLines []string) []string {
	managed := map[string]struct{}{
		apiBaseEnv:  {},
		agentIDEnv:  {},
		apiTokenEnv: {},
	}
	lines := make([]string, 0)
	for _, line := range strings.Split(strings.ReplaceAll(existing, "\r\n", "\n"), "\n") {
		if line == "" {
			continue
		}
		if key, ok := envLineKey(line); ok {
			if _, isManaged := managed[key]; isManaged {
				continue
			}
		}
		lines = append(lines, line)
	}
	return append(lines, replacementLines...)
}

func envLineKey(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", false
	}
	trimmed = strings.TrimPrefix(trimmed, "export ")
	idx := strings.Index(trimmed, "=")
	if idx < 1 {
		return "", false
	}
	key := strings.TrimSpace(trimmed[:idx])
	return key, key != ""
}

func formatDotenvLine(key, value string) string {
	return key + "=" + value
}

func formatExportLine(key, value string) string {
	return "export " + key + "=" + shellQuote(value)
}

func shellQuote(value string) string {
	return strconv.Quote(value)
}

func redactToken(token string) string {
	token = strings.TrimSpace(token)
	if len(token) <= 8 {
		return "********"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
