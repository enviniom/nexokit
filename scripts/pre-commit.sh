#!/bin/bash
set -euo pipefail

# =============================================================================
# Pre-commit hook for nexokit
# =============================================================================

FILE_SIZE_LIMIT_MB=1

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass() { printf "${GREEN}✓ %s${NC}\n" "$1"; }
fail() { printf "${RED}✗ %s${NC}\n" "$1"; exit 1; }
warn() { printf "${YELLOW}⚠ %s${NC}\n" "$1"; }

check_binaries() {
	# Use native git binary detection: numstat outputs "-  -  path" for binary files
	local binaries
	binaries=$(git diff --cached --name-only | while IFS= read -r file; do
		if [ -f "$file" ] && git diff --cached --numstat -- "$file" | grep -q '^-'; then
			echo "$file"
		fi
	done)

	if [ -n "$binaries" ]; then
		fail "Binary files detected in staged changes — remove them before committing:
$binaries"
	fi
	pass "No binary files staged"
}

check_file_size() {
	local max_bytes=$((FILE_SIZE_LIMIT_MB * 1024 * 1024))
	local large_files=""
	local file size

	while IFS= read -r file; do
		[ -z "$file" ] && continue
		[ -f "$file" ] || continue
		size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo 0)
		if [ "$size" -gt "$max_bytes" ]; then
			large_files="${large_files}\n  - ${file} ($((size / 1024)) KB)"
		fi
	done < <(git diff --cached --name-only --diff-filter=ACM)

	if [ -n "$large_files" ]; then
		warn "Files larger than ${FILE_SIZE_LIMIT_MB}MB staged:${large_files}"
	else
		pass "No oversized files staged"
	fi
}

check_env_parity() {
	[ -f .env ] || return 0
	[ -f .env.example ] || { warn ".env.example not found — skipping parity check"; return 0; }

	extract_keys() {
		sed '/^[[:space:]]*#/d;/^[[:space:]]*$/d' "$1" | awk -F= '{print $1}' | sed 's/[[:space:]]*$//' | sort -u
	}

	local env_keys example_keys missing_from_env missing_from_example

	env_keys=$(extract_keys .env)
	example_keys=$(extract_keys .env.example)

	missing_from_env=$(comm -23 <(echo "$example_keys") <(echo "$env_keys"))
	missing_from_example=$(comm -13 <(echo "$example_keys") <(echo "$env_keys"))

	if [ -n "$missing_from_env" ]; then
		warn "Keys in .env.example not set in your .env:
$missing_from_env"
	fi

	if [ -n "$missing_from_example" ]; then
		warn "Keys in .env not documented in .env.example — add them before committing:
$missing_from_example"
	fi

	if [ -z "$missing_from_env" ] && [ -z "$missing_from_example" ]; then
		pass ".env parity OK"
	fi
}

check_go_vet() {
	if ! go vet ./... 2>&1; then
		fail "go vet ./... found issues — fix them before committing"
	fi
	pass "go vet passed"
}

check_go_fmt() {
	# Collect staged Go files into an array to avoid xargs-with-empty-input pitfall
	local go_files=()
	while IFS= read -r file; do
		[ -n "$file" ] && go_files+=("$file")
	done < <(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

	if [ ${#go_files[@]} -eq 0 ]; then
		pass "No Go files staged — skipping fmt check"
		return 0
	fi

	local unformatted
	unformatted=$(gofmt -l "${go_files[@]}")

	if [ -n "$unformatted" ]; then
		fail "Unformatted Go files detected — run 'make fmt' and re-stage:
$unformatted"
	fi
	pass "go fmt OK"
}

# =============================================================================
# Main
# =============================================================================

if [ "${1:-}" = "--check-env-only" ]; then
	check_env_parity
	exit 0
fi

check_binaries
check_file_size
check_env_parity
check_go_vet
check_go_fmt

pass "All pre-commit checks passed"
