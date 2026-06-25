package target

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	latheconfig "github.com/lathe-cli/lathe/pkg/config"
	"github.com/spf13/cobra"
)

const (
	LocalTarget  = "local"
	CloudTarget  = "cloud"
	CustomTarget = "custom"

	DefaultLocalBaseURL = "http://127.0.0.1:8787"
	DefaultCloudBaseURL = "https://api.mosoo.ai"

	TargetEnv  = "MOSOO_TARGET"
	BaseURLEnv = "MOSOO_BASE_URL"
)

const (
	SourceHostnameFlag  = "hostname-flag"
	SourceHostEnv       = "host-env"
	SourceTargetFlag    = "target-flag"
	SourceTargetEnv     = "target-env"
	SourceProjectConfig = "project-config"
	SourceGlobalConfig  = "global-config"
	SourceCwdMosooRepo  = "cwd-mosoo-repo"
	SourceDefaultLocal  = "default-local"
)

const (
	SurfaceConsole         = "console"
	SurfaceConsoleREST     = "console-rest"
	SurfacePublicThreadAPI = "public-thread-api"
)

// Resolution describes the Mosoo service target selected for this invocation.
type Resolution struct {
	Target           string            `json:"target"`
	Source           string            `json:"source"`
	BaseURL          string            `json:"baseUrl"`
	Hosts            map[string]string `json:"hosts"`
	ConfigPath       string            `json:"configPath,omitempty"`
	ProjectRoot      string            `json:"projectRoot,omitempty"`
	ExplicitHostname string            `json:"explicitHostname,omitempty"`
}

type fileConfig struct {
	Target       string `json:"target"`
	BaseURL      string `json:"baseUrl"`
	BaseURLSnake string `json:"base_url"`
}

// Install adds target flags and wires runtime target resolution into generated API commands.
func Install(root *cobra.Command) {
	flags := root.PersistentFlags()
	if flags.Lookup("target") == nil {
		flags.String("target", "", "Mosoo target for generated API commands: local|cloud|custom")
	}
	if flags.Lookup("base-url") == nil {
		flags.String("base-url", "", "Mosoo service base URL used with --target or target config")
	}

	previousPreRun := root.PersistentPreRun
	previousPreRunE := root.PersistentPreRunE
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if previousPreRun != nil {
			previousPreRun(cmd, args)
		}
		if previousPreRunE != nil {
			if err := previousPreRunE(cmd, args); err != nil {
				return err
			}
		}

		surface, ok := SurfaceForCommand(cmd)
		if !ok || HasExplicitHostname(cmd) {
			return nil
		}
		resolved, err := ResolveFromCommand(cmd)
		if err != nil {
			return err
		}
		hostname := resolved.Hosts[surface]
		if hostname == "" {
			return fmt.Errorf("no hostname resolved for surface %q", surface)
		}
		return cmd.Root().PersistentFlags().Set("hostname", hostname)
	}
}

// ResolveFromCommand resolves the current target using flags, environment, config, cwd, then defaults.
func ResolveFromCommand(cmd *cobra.Command) (Resolution, error) {
	flags := cmd.Root().PersistentFlags()
	if flags.Lookup("hostname") != nil && flags.Changed("hostname") {
		hostname, _ := flags.GetString("hostname")
		return ResolveExplicitHostname(hostname, SourceHostnameFlag)
	}
	if envHost := strings.TrimSpace(os.Getenv(latheconfig.Active().CLI.HostEnv)); envHost != "" {
		return ResolveExplicitHostname(envHost, SourceHostEnv)
	}

	flagTarget, flagBaseURL := "", ""
	if flags.Lookup("target") != nil {
		flagTarget, _ = flags.GetString("target")
	}
	if flags.Lookup("base-url") != nil {
		flagBaseURL, _ = flags.GetString("base-url")
	}
	if strings.TrimSpace(flagTarget) != "" || strings.TrimSpace(flagBaseURL) != "" {
		return resolutionFromTargetBase(flagTarget, flagBaseURL, SourceTargetFlag, "", "")
	}

	if envTarget := strings.TrimSpace(os.Getenv(TargetEnv)); envTarget != "" || strings.TrimSpace(os.Getenv(BaseURLEnv)) != "" {
		return resolutionFromTargetBase(envTarget, os.Getenv(BaseURLEnv), SourceTargetEnv, "", "")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return Resolution{}, err
	}
	return Resolve(cwd)
}

// Resolve resolves a target from config and cwd context. It intentionally does not inspect
// --hostname, MOSOO_HOST, --target, or target environment variables.
func Resolve(cwd string) (Resolution, error) {
	if path, ok := findProjectConfig(cwd); ok {
		return readConfig(path, SourceProjectConfig)
	}
	if path, ok := globalConfigPath(); ok {
		return readConfig(path, SourceGlobalConfig)
	}
	if root, ok := findMosooSourceRoot(cwd); ok {
		resolved, err := resolutionFromTargetBase(LocalTarget, "", SourceCwdMosooRepo, "", root)
		if err != nil {
			return Resolution{}, err
		}
		return resolved, nil
	}
	return resolutionFromTargetBase(LocalTarget, "", SourceDefaultLocal, "", "")
}

// ResolveExplicitHostname describes an explicit Lathe hostname override. Generated commands
// use the exact hostname directly; this method exists for doctor and diagnostics.
func ResolveExplicitHostname(hostname, source string) (Resolution, error) {
	h := strings.TrimRight(strings.TrimSpace(hostname), "/")
	if h == "" {
		return Resolution{}, fmt.Errorf("hostname must not be empty")
	}
	withScheme := h
	if !strings.Contains(withScheme, "://") {
		withScheme = "https://" + withScheme
	}
	baseURL := stripAPISuffix(withScheme)
	target := targetForBaseURL(baseURL)
	return Resolution{
		Target:           target,
		Source:           source,
		BaseURL:          baseURL,
		Hosts:            HostsForBaseURL(baseURL),
		ExplicitHostname: h,
	}, nil
}

func HasExplicitHostname(cmd *cobra.Command) bool {
	flags := cmd.Root().PersistentFlags()
	if flags.Lookup("hostname") != nil && flags.Changed("hostname") {
		return true
	}
	return strings.TrimSpace(os.Getenv(latheconfig.Active().CLI.HostEnv)) != ""
}

func SurfaceForCommand(cmd *cobra.Command) (string, bool) {
	parts := strings.Fields(cmd.CommandPath())
	for i := 1; i < len(parts); i++ {
		switch parts[i] {
		case SurfaceConsole, SurfaceConsoleREST, SurfacePublicThreadAPI:
			return parts[i], true
		}
	}
	return "", false
}

func HostsForBaseURL(baseURL string) map[string]string {
	base := strings.TrimRight(baseURL, "/")
	return map[string]string{
		SurfaceConsole:         base + "/api",
		SurfaceConsoleREST:     base + "/api",
		SurfacePublicThreadAPI: base + "/api/v1",
	}
}

func resolutionFromTargetBase(targetValue, baseURLValue, source, configPath, projectRoot string) (Resolution, error) {
	targetValue = strings.ToLower(strings.TrimSpace(targetValue))
	baseURLValue = strings.TrimSpace(baseURLValue)

	if targetValue == "" {
		if baseURLValue != "" {
			targetValue = CustomTarget
		} else {
			targetValue = LocalTarget
		}
	}

	switch targetValue {
	case LocalTarget:
		if baseURLValue == "" {
			baseURLValue = DefaultLocalBaseURL
		}
	case CloudTarget:
		if baseURLValue == "" {
			baseURLValue = DefaultCloudBaseURL
		}
	case CustomTarget:
		if baseURLValue == "" {
			return Resolution{}, fmt.Errorf("baseUrl is required for custom target")
		}
	default:
		return Resolution{}, fmt.Errorf("target must be one of %q, %q, or %q", LocalTarget, CloudTarget, CustomTarget)
	}

	baseURL, err := normalizeBaseURL(baseURLValue)
	if err != nil {
		return Resolution{}, err
	}
	return Resolution{
		Target:      targetValue,
		Source:      source,
		BaseURL:     baseURL,
		Hosts:       HostsForBaseURL(baseURL),
		ConfigPath:  configPath,
		ProjectRoot: projectRoot,
	}, nil
}

func readConfig(path, source string) (Resolution, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Resolution{}, err
	}
	var cfg fileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Resolution{}, fmt.Errorf("parse %s: %w", path, err)
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = cfg.BaseURLSnake
	}
	return resolutionFromTargetBase(cfg.Target, baseURL, source, path, "")
}

func normalizeBaseURL(raw string) (string, error) {
	value := strings.TrimRight(strings.TrimSpace(raw), "/")
	if value == "" {
		return "", fmt.Errorf("baseUrl must not be empty")
	}
	if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
		return "", fmt.Errorf("baseUrl must include http:// or https://")
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("parse baseUrl: %w", err)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("baseUrl must include a host")
	}
	return value, nil
}

func stripAPISuffix(hostname string) string {
	value := strings.TrimRight(hostname, "/")
	for _, suffix := range []string{"/api/v1", "/api"} {
		if strings.HasSuffix(value, suffix) {
			return strings.TrimRight(strings.TrimSuffix(value, suffix), "/")
		}
	}
	return value
}

func targetForBaseURL(baseURL string) string {
	if IsLocalBaseURL(baseURL) {
		return LocalTarget
	}
	if strings.EqualFold(strings.TrimRight(baseURL, "/"), DefaultCloudBaseURL) {
		return CloudTarget
	}
	return CustomTarget
}

// IsLocalBaseURL reports whether a base URL points at the local development API.
func IsLocalBaseURL(baseURL string) bool {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return false
	}
	switch parsed.Hostname() {
	case "127.0.0.1", "localhost", "::1":
		return true
	}
	return false
}

func findProjectConfig(start string) (string, bool) {
	dir, ok := absoluteDir(start)
	if !ok {
		return "", false
	}
	for {
		path := filepath.Join(dir, ".mosoo", "config.json")
		if isFile(path) {
			return path, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func globalConfigPath() (string, bool) {
	cli := latheconfig.Active().CLI
	if dir := strings.TrimSpace(os.Getenv(cli.ConfigDirEnv)); dir != "" {
		path := filepath.Join(dir, "config.json")
		return path, isFile(path)
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", false
	}
	path := filepath.Join(dir, cli.ConfigDir, "config.json")
	return path, isFile(path)
}

func findMosooSourceRoot(start string) (string, bool) {
	dir, ok := absoluteDir(start)
	if !ok {
		return "", false
	}
	for {
		if isMosooSourceRoot(dir) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func isMosooSourceRoot(dir string) bool {
	required := []string{
		"package.json",
		"justfile",
		filepath.Join("apps", "api", "wrangler.toml"),
	}
	for _, name := range required {
		if !isFile(filepath.Join(dir, name)) {
			return false
		}
	}
	return true
}

func absoluteDir(path string) (string, bool) {
	if path == "" {
		return "", false
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", false
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", false
	}
	if !info.IsDir() {
		abs = filepath.Dir(abs)
	}
	return abs, true
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
