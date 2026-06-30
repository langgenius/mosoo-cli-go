#!/usr/bin/env bash
set -euo pipefail

PROGRAM_NAME="mosoo install"
RELEASE_BASE_URL="${MOSOO_RELEASE_BASE_URL:-https://github.com/langgenius/mosoo-connector/releases/latest/download}"
BIN_DIR="${MOSOO_BIN_DIR:-$HOME/.local/bin}"
CODEX_HOME_VALUE="${CODEX_HOME:-$HOME/.codex}"
SKILL_DIR="${MOSOO_SKILL_DIR:-$CODEX_HOME_VALUE/skills/mosoo}"
SOURCE_ROOT="${MOSOO_INSTALL_SOURCE_ROOT:-}"
CLI_ARCHIVE_URL="${MOSOO_CLI_ARCHIVE_URL:-}"
SKILL_ARCHIVE_URL="${MOSOO_SKILL_ARCHIVE_URL:-}"
CLI_VERSION="${MOSOO_CLI_VERSION:-}"
CLI_SHA256="${MOSOO_CLI_SHA256:-}"
SKILL_SHA256="${MOSOO_SKILL_SHA256:-}"
TARGET="${MOSOO_TARGET:-cloud}"
BASE_URL="${MOSOO_BASE_URL:-}"
DEV_EMAIL="${MOSOO_DEV_EMAIL:-}"
LOGIN_URL="${MOSOO_LOGIN_URL:-https://try.mosoo.ai}"

ASSUME_YES=false
DRY_RUN=false
INSTALL_CLI=true
INSTALL_SKILL=true
RUN_LOGIN=true
RUN_DOCTOR=true
WRITE_CONFIG=false
SETUP_CLOUDFLARE=false

tmp_dirs=()

usage() {
	cat <<'EOF'
Mosoo Installer

Usage:
  install.sh [options]

Examples:
  curl -fsSL https://install.mosoo.ai/install.sh | bash
  curl -fsSL https://install.mosoo.ai/install.sh | bash -s -- --yes
  curl -fsSL https://install.mosoo.ai/install.sh | bash -s -- --dry-run
  curl -fsSL https://install.mosoo.ai/install.sh | bash -s -- --target local
  curl -fsSL https://install.mosoo.ai/install.sh | bash -s -- --cloudflare

Options:
  -y, --yes                 Run with default approvals; never prompt.
      --dry-run             Print the plan and commands without changing files.
      --bin-dir DIR         Install mosoo CLI into DIR. Default: ~/.local/bin.
      --skill-dir DIR       Install Mosoo Skill into DIR. Default: $CODEX_HOME/skills/mosoo or ~/.codex/skills/mosoo.
      --source-root DIR     Install from a local mosoo-connector checkout for development.
      --cli-url URL         Download CLI archive from URL.
      --skill-url URL       Download Skill archive from URL.
      --target TARGET       Runtime target for login and doctor: local, cloud, or custom. Default: cloud.
      --base-url URL        Base URL for --target custom, or override local/cloud base URL.
      --write-config        Write the selected target and base URL to global Mosoo config.
      --no-cli              Skip CLI install/update.
      --no-skill            Skip Skill install/update.
      --no-login            Skip auth login.
      --no-doctor           Skip final mosoo doctor --json.
      --cloudflare          Also run optional Cloudflare/Wrangler onboarding checks.
  -h, --help                Show this help.

Environment:
  MOSOO_CLI_ARCHIVE_URL     Override CLI release archive URL.
  MOSOO_SKILL_ARCHIVE_URL   Override Skill release archive URL.
  MOSOO_CLI_VERSION         Expected mosoo --version value, for example v1.2.3.
  MOSOO_API_TOKEN           Token used for non-interactive mosoo auth login.
  MOSOO_DEV_EMAIL           @mosoo.ai email used for local development login.
  MOSOO_LOGIN_URL           Web login URL shown when cloud login needs user action.
  MOSOO_TARGET              Runtime target used by login and doctor. Default: cloud.
  MOSOO_INSTALL_SOURCE_ROOT
                            Local checkout root used by development installs.
EOF
}

log() {
	printf '%s\n' "$*"
}

warn() {
	printf 'warning: %s\n' "$*" >&2
}

die() {
	printf 'error: %s\n' "$*" >&2
	exit 1
}

cleanup() {
	local dir
	set +u
	for dir in "${tmp_dirs[@]}"; do
		rm -rf "$dir"
	done
	set -u
}
trap cleanup EXIT

mktemp_dir() {
	local dir
	dir="$(mktemp -d)"
	tmp_dirs+=("$dir")
	printf '%s\n' "$dir"
}

print_cmd() {
	local arg
	printf '+'
	for arg in "$@"; do
		printf ' %q' "$arg"
	done
	printf '\n'
}

run() {
	if "$DRY_RUN"; then
		print_cmd "$@"
		return 0
	fi
	"$@"
}

json_string() {
	local value="$1"
	JSON_VALUE="$value" python3 -c 'import json, os; print(json.dumps(os.environ["JSON_VALUE"]))'
}

json_field() {
	local field="$1"
	JSON_FIELD="$field" python3 -c 'import json, os, sys
value = json.load(sys.stdin)
for part in os.environ["JSON_FIELD"].split("."):
    value = value[part]
print(value)'
}

need_cmd() {
	command -v "$1" >/dev/null 2>&1 || die "$1 is required but was not found on PATH"
}

confirm() {
	local prompt="$1"
	local default="${2:-n}"
	local reply suffix

	suffix="[y/N]"
	if [ "$default" = "y" ]; then
		suffix="[Y/n]"
	fi

	if "$DRY_RUN"; then
		log "dry-run: would ask: $prompt $suffix"
		return 0
	fi
	if "$ASSUME_YES"; then
		log "auto-approve: $prompt $suffix"
		return 0
	fi
	if [ ! -r /dev/tty ]; then
		die "cannot prompt without a TTY; re-run with --yes or --dry-run"
	fi

	while true; do
		printf '%s %s ' "$prompt" "$suffix" >/dev/tty
		read -r reply </dev/tty || die "failed to read confirmation"
		case "$reply" in
			"" )
				[ "$default" = "y" ] && return 0
				return 1
				;;
			y|Y|yes|YES) return 0 ;;
			n|N|no|NO) return 1 ;;
			*) printf 'Please answer y or n.\n' >/dev/tty ;;
		esac
	done
}

detect_platform() {
	local os_name arch_name
	case "$(uname -s)" in
		Darwin) os_name="darwin" ;;
		Linux) os_name="linux" ;;
		*) die "unsupported OS: $(uname -s)" ;;
	esac

	case "$(uname -m)" in
		arm64|aarch64) arch_name="arm64" ;;
		x86_64|amd64) arch_name="amd64" ;;
		*) die "unsupported architecture: $(uname -m)" ;;
	esac

	printf '%s-%s\n' "$os_name" "$arch_name"
}

default_base_url() {
	local target="$1"
	case "$target" in
		""|local) printf '%s\n' "${BASE_URL:-http://127.0.0.1:8787}" ;;
		cloud) printf '%s\n' "${BASE_URL:-https://try.mosoo.ai}" ;;
		custom) [ -n "$BASE_URL" ] || die "--base-url is required for --target custom"; printf '%s\n' "$BASE_URL" ;;
		*) die "--target must be one of local, cloud, or custom" ;;
	esac
}

console_host() {
	printf '%s/api\n' "$(default_base_url "$TARGET" | sed 's#/*$##')"
}

public_api_host() {
	printf '%s/api/v1\n' "$(default_base_url "$TARGET" | sed 's#/*$##')"
}

base_origin() {
	default_base_url "$TARGET" | sed 's#/*$##'
}

global_config_dir() {
	if [ -n "${MOSOO_CONFIG_DIR:-}" ]; then
		printf '%s\n' "$MOSOO_CONFIG_DIR"
		return
	fi
	case "$(uname -s)" in
		Darwin) printf '%s\n' "$HOME/Library/Application Support/mosoo" ;;
		*) printf '%s\n' "${XDG_CONFIG_HOME:-$HOME/.config}/mosoo" ;;
	esac
}

verify_sha256() {
	local file="$1"
	local expected="$2"
	local actual
	[ -n "$expected" ] || return 0
	if command -v shasum >/dev/null 2>&1; then
		actual="$(shasum -a 256 "$file" | awk '{print $1}')"
	elif command -v sha256sum >/dev/null 2>&1; then
		actual="$(sha256sum "$file" | awk '{print $1}')"
	else
		die "shasum or sha256sum is required to verify checksum"
	fi
	[ "$actual" = "$expected" ] || die "checksum mismatch for $file"
}

download_file() {
	local url="$1"
	local dest="$2"
	need_cmd curl
	run curl -fsSL "$url" -o "$dest"
}

extract_cli_binary() {
	local archive="$1"
	local out_dir="$2"
	local binary
	need_cmd tar
	tar -xzf "$archive" -C "$out_dir"
	binary="$(find "$out_dir" -type f -name mosoo -print | head -n 1)"
	[ -n "$binary" ] || die "CLI archive did not contain a mosoo binary"
	printf '%s\n' "$binary"
}

find_skill_source() {
	local root="$1"
	local skill_file
	skill_file="$(find "$root" -type f -path '*/mosoo/SKILL.md' -print | head -n 1)"
	[ -n "$skill_file" ] || die "Skill archive did not contain mosoo/SKILL.md"
	dirname "$skill_file"
}

install_cli() {
	local platform url tmp archive extract_dir binary source_binary
	platform="$(detect_platform)"
	url="${CLI_ARCHIVE_URL:-$RELEASE_BASE_URL/mosoo-$platform.tar.gz}"

	if "$DRY_RUN"; then
		source_binary="${SOURCE_ROOT:+$SOURCE_ROOT/bin/mosoo}"
		log "CLI source: ${source_binary:-$url}${source_binary:+" (local)"}"
		run mkdir -p "$BIN_DIR"
		run install -m 0755 "${source_binary:-mosoo}" "$BIN_DIR/mosoo"
		verify_installed_cli
		return
	fi

	run mkdir -p "$BIN_DIR"
	if [ -n "$SOURCE_ROOT" ]; then
		source_binary="$SOURCE_ROOT/bin/mosoo"
		[ -x "$source_binary" ] || die "local CLI binary not found or not executable: $source_binary"
		run install -m 0755 "$source_binary" "$BIN_DIR/mosoo"
	else
		tmp="$(mktemp_dir)"
		archive="$tmp/mosoo-$platform.tar.gz"
		extract_dir="$tmp/extract"
		run mkdir -p "$extract_dir"
		download_file "$url" "$archive"
		verify_sha256 "$archive" "$CLI_SHA256"
		binary="$(extract_cli_binary "$archive" "$extract_dir")"
		run install -m 0755 "$binary" "$BIN_DIR/mosoo"
	fi

	case ":$PATH:" in
		*":$BIN_DIR:"*) ;;
		*) warn "$BIN_DIR is not on PATH; add it before running mosoo from a new shell" ;;
	esac

	verify_installed_cli
}

verify_installed_cli() {
	local mosoo output

	if "$DRY_RUN"; then
		print_cmd "$BIN_DIR/mosoo" --version
		return
	fi

	mosoo="$(resolve_mosoo_binary)"
	output="$("$mosoo" --version)" || die "installed mosoo failed to print version"
	case "$output" in
		"mosoo "*" ("*", "*) ;;
		*) die "installed mosoo printed an unexpected version string: $output" ;;
	esac
	if [ -n "$CLI_VERSION" ]; then
		case "$output" in
			"mosoo $CLI_VERSION ("*) ;;
			*) die "installed mosoo version mismatch: expected $CLI_VERSION, got $output" ;;
		esac
	fi
	log "verified CLI: $output"
}

install_skill() {
	local tmp archive extract_dir skill_source url
	url="${SKILL_ARCHIVE_URL:-$RELEASE_BASE_URL/mosoo-skill.tar.gz}"

	if "$DRY_RUN"; then
		skill_source="${SOURCE_ROOT:+$SOURCE_ROOT/publish/skills/mosoo}"
		log "Skill source: ${skill_source:-$url}${skill_source:+" (local)"}"
		run rm -rf "$SKILL_DIR"
		run mkdir -p "$(dirname "$SKILL_DIR")"
		run cp -R "${skill_source:-mosoo}" "$SKILL_DIR"
		return
	fi

	if [ -n "$SOURCE_ROOT" ]; then
		skill_source="$SOURCE_ROOT/publish/skills/mosoo"
		[ -f "$skill_source/SKILL.md" ] || die "local Mosoo Skill not found: $skill_source"
	else
		tmp="$(mktemp_dir)"
		archive="$tmp/mosoo-skill.tar.gz"
		extract_dir="$tmp/extract"
		run mkdir -p "$extract_dir"
		download_file "$url" "$archive"
		verify_sha256 "$archive" "$SKILL_SHA256"
		need_cmd tar
		tar -xzf "$archive" -C "$extract_dir"
		skill_source="$(find_skill_source "$extract_dir")"
	fi

	run rm -rf "$SKILL_DIR"
	run mkdir -p "$(dirname "$SKILL_DIR")"
	run cp -R "$skill_source" "$SKILL_DIR"
}

write_target_config() {
	local config_dir config_file target base
	target="${TARGET:-local}"
	base="$(default_base_url "$target")"
	config_dir="$(global_config_dir)"
	config_file="$config_dir/config.json"

	run mkdir -p "$config_dir"
	if "$DRY_RUN"; then
		log "dry-run: would write $config_file:"
		printf '{\n  "target": "%s",\n  "baseUrl": "%s"\n}\n' "$target" "$base"
		return
	fi
	printf '{\n  "target": "%s",\n  "baseUrl": "%s"\n}\n' "$target" "$base" >"$config_file"
	log "wrote $config_file"
}

resolve_mosoo_binary() {
	local candidate="$BIN_DIR/mosoo"
	if [ -x "$candidate" ]; then
		printf '%s\n' "$candidate"
		return
	fi
	if command -v mosoo >/dev/null 2>&1; then
		command -v mosoo
		return
	fi
	die "mosoo CLI is not installed; cannot continue"
}

store_api_token() {
	local mosoo token console public
	mosoo="$1"
	token="$2"
	console="$(console_host)"
	public="$(public_api_host)"

	[ -n "$token" ] || die "Mosoo API token must not be empty"
	printf '%s\n' "$token" | "$mosoo" auth login --hostname "$console" --with-token
	printf '%s\n' "$token" | "$mosoo" auth login --hostname "$public" --skip-validate --with-token
}

read_local_development_email() {
	local email
	email="$DEV_EMAIL"
	if [ -n "$email" ]; then
		printf '%s\n' "$email"
		return
	fi
	if "$ASSUME_YES"; then
		printf 'dev@mosoo.ai\n'
		return
	fi

	[ -r /dev/tty ] || die "cannot prompt for local development email without a TTY"
	printf 'Enter local development email (@mosoo.ai): ' >/dev/tty
	read -r email </dev/tty || die "failed to read email"
	[ -n "$email" ] || die "local development email must not be empty"
	printf '%s\n' "$email"
}

run_local_development_login() {
	local mosoo email origin console login_url token_url tmp cookie_jar login_body token_body token_response token
	console="$(console_host)"
	origin="$(base_origin)"
	login_url="$console/auth/development-backdoor/mosoo-ai-login"
	token_url="$console/access-tokens"

	if "$DRY_RUN"; then
		print_cmd curl -fsSL -c cookies.txt -H "content-type: application/json" -H "origin: $origin" --data '{"email":"dev@mosoo.ai"}' "$login_url"
			print_cmd curl -fsSL -b cookies.txt -H "content-type: application/json" -H "origin: $origin" --data '{"label":"Mosoo CLI local install"}' "$token_url"
		print_cmd "$BIN_DIR/mosoo" auth login --hostname "$console" --with-token
		print_cmd "$BIN_DIR/mosoo" auth login --hostname "$(public_api_host)" --skip-validate --with-token
		return
	fi

	need_cmd curl
	need_cmd python3
	mosoo="$(resolve_mosoo_binary)"
	email="$(read_local_development_email)"
	case "$email" in
		*@mosoo.ai) ;;
		*) die "local development login email must use @mosoo.ai" ;;
	esac

	tmp="$(mktemp_dir)"
	cookie_jar="$tmp/cookies.txt"
	login_body="$tmp/login.json"
	token_body="$tmp/token.json"
	token_response="$tmp/token-response.json"

	printf '{"email":%s}\n' "$(json_string "$email")" >"$login_body"
	printf '{"label":"Mosoo CLI local install"}\n' >"$token_body"

	curl -fsSL -c "$cookie_jar" \
		-H "content-type: application/json" \
		-H "origin: $origin" \
		--data @"$login_body" \
		"$login_url" >/dev/null ||
		die "local development login failed; ensure the local Mosoo API is running and the development backdoor is enabled"

	curl -fsSL -b "$cookie_jar" \
		-H "content-type: application/json" \
		-H "origin: $origin" \
		--data @"$token_body" \
		"$token_url" >"$token_response" ||
		die "local API token creation failed after development login"

	token="$(json_field value <"$token_response")"
	store_api_token "$mosoo" "$token"
}

run_login() {
	local mosoo token
	if [ "$TARGET" = "local" ]; then
		log "Using local development login."
		run_local_development_login
		return
	fi

	if "$DRY_RUN"; then
		mosoo="$BIN_DIR/mosoo"
		log "dry-run: would show cloud login instructions: $LOGIN_URL"
		print_cmd "$mosoo" auth login --hostname "$(console_host)" --with-token
		print_cmd "$mosoo" auth login --hostname "$(public_api_host)" --skip-validate --with-token
		return
	fi

	mosoo="$(resolve_mosoo_binary)"

	if [ -n "${MOSOO_API_TOKEN:-}" ]; then
		token="$MOSOO_API_TOKEN"
	elif "$ASSUME_YES"; then
		warn "MOSOO_API_TOKEN is not set; skipping non-interactive cloud login"
		log "Sign in to Mosoo Cloud first, then rerun this installer:"
		log "  $LOGIN_URL"
		return
	else
		[ -r /dev/tty ] || die "cannot prompt for token without a TTY"
		cat >/dev/tty <<EOF
Cloud login needs a Mosoo API token from a logged-in Mosoo web session.

1. Open Mosoo Cloud:
   $LOGIN_URL
2. Sign in or create an account with email verification.
3. Copy the install command from the web app, or create and copy an API token.
4. Paste the API token here, or rerun this installer with MOSOO_API_TOKEN set.

EOF
		printf 'Enter Mosoo API token: ' >/dev/tty
		read -rs token </dev/tty || die "failed to read token"
		printf '\n' >/dev/tty
	fi

	store_api_token "$mosoo" "$token"
}

run_doctor() {
	local mosoo args
	mosoo="$(resolve_mosoo_binary)"
	args=(doctor --json)
	if [ -n "$TARGET" ]; then
		args+=(--target "$TARGET")
	fi
	if [ -n "$BASE_URL" ]; then
		args+=(--base-url "$BASE_URL")
	fi
	run "$mosoo" "${args[@]}"
}

setup_cloudflare() {
	if "$DRY_RUN"; then
		run command -v wrangler
		run wrangler login
		run wrangler whoami
		return
	fi
	if ! command -v wrangler >/dev/null 2>&1; then
		warn "wrangler is not installed"
		log "Install Wrangler only for Cloudflare deployment tasks:"
		log "  npm install -g wrangler"
		return
	fi
	wrangler --version
	if confirm "Run wrangler login now?" "n"; then
		run wrangler login
	fi
	run wrangler whoami || warn "wrangler is installed but not authenticated"
}

print_plan() {
	local platform config_path cli_source skill_source login_plan
	platform="$(detect_platform)"
	config_path="$(global_config_dir)/config.json"
	cli_source="${CLI_ARCHIVE_URL:-$RELEASE_BASE_URL/mosoo-$platform.tar.gz}"
	skill_source="${SKILL_ARCHIVE_URL:-$RELEASE_BASE_URL/mosoo-skill.tar.gz}"
	login_plan="$RUN_LOGIN"
	if "$RUN_LOGIN" && [ "$TARGET" = "local" ]; then
		login_plan="local development email login"
	fi
	if [ -n "$SOURCE_ROOT" ]; then
		cli_source="$SOURCE_ROOT/bin/mosoo (local)"
		skill_source="$SOURCE_ROOT/publish/skills/mosoo (local)"
	fi
	cat <<EOF
$PROGRAM_NAME plan

Platform: $platform
Install/update CLI: $INSTALL_CLI
  Target: $BIN_DIR/mosoo
  Source: $cli_source
  Expected version: ${CLI_VERSION:-any build metadata}
Install/update Skill: $INSTALL_SKILL
  Target: $SKILL_DIR
  Source: $skill_source
Write config: $WRITE_CONFIG
  Target config: ${TARGET:-local}
  Config file: $config_path
Run login: $login_plan
Run doctor: $RUN_DOCTOR
Cloudflare setup: $SETUP_CLOUDFLARE
EOF
}

parse_args() {
	while [ "$#" -gt 0 ]; do
		case "$1" in
			-y|--yes) ASSUME_YES=true ;;
			--dry-run) DRY_RUN=true ;;
			--bin-dir) shift; [ "$#" -gt 0 ] || die "--bin-dir requires a value"; BIN_DIR="$1" ;;
			--skill-dir) shift; [ "$#" -gt 0 ] || die "--skill-dir requires a value"; SKILL_DIR="$1" ;;
			--source-root) shift; [ "$#" -gt 0 ] || die "--source-root requires a value"; SOURCE_ROOT="$1" ;;
			--cli-url) shift; [ "$#" -gt 0 ] || die "--cli-url requires a value"; CLI_ARCHIVE_URL="$1" ;;
			--skill-url) shift; [ "$#" -gt 0 ] || die "--skill-url requires a value"; SKILL_ARCHIVE_URL="$1" ;;
			--target) shift; [ "$#" -gt 0 ] || die "--target requires a value"; TARGET="$1" ;;
			--base-url) shift; [ "$#" -gt 0 ] || die "--base-url requires a value"; BASE_URL="$1" ;;
			--write-config) WRITE_CONFIG=true ;;
			--no-cli) INSTALL_CLI=false ;;
			--no-skill) INSTALL_SKILL=false ;;
			--no-login) RUN_LOGIN=false ;;
			--no-doctor) RUN_DOCTOR=false ;;
			--cloudflare) SETUP_CLOUDFLARE=true ;;
			-h|--help) usage; exit 0 ;;
			--) shift; break ;;
			*) die "unknown option: $1" ;;
		esac
		shift
	done
}

main() {
	parse_args "$@"
	need_cmd uname
	need_cmd sed
	need_cmd find
	need_cmd dirname
	print_plan

	if "$INSTALL_CLI" && confirm "Install or update Mosoo CLI at $BIN_DIR/mosoo?" "y"; then
		install_cli
	fi
	if "$INSTALL_SKILL" && confirm "Install or update Mosoo Skill at $SKILL_DIR?" "y"; then
		install_skill
	fi
	if "$WRITE_CONFIG" && confirm "Write Mosoo target config?" "y"; then
		write_target_config
	fi
	if "$RUN_LOGIN" && confirm "Run Mosoo auth login now?" "y"; then
		run_login
	fi
	if "$SETUP_CLOUDFLARE" && confirm "Run optional Cloudflare/Wrangler onboarding checks?" "n"; then
		setup_cloudflare
	fi
	if "$RUN_DOCTOR" && confirm "Run mosoo doctor --json?" "y"; then
		run_doctor
	fi

	log "Mosoo install finished."
}

main "$@"
