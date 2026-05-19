# =============================================================================
# AngelaMos | 2026
# Justfile
# =============================================================================
# test-audit — deterministic certification practice-test auditor
# =============================================================================

set export
set shell := ["bash", "-uc"]

project := file_name(justfile_directory())
version := `git describe --tags --always 2>/dev/null || echo "dev"`
ldflags := "-s -w -X github.com/CarterPerez-dev/test-audit/internal/app.version=" + version

# =============================================================================
# Default
# =============================================================================

default:
    @just --list --unsorted

# =============================================================================
# Linting and Formatting
# =============================================================================

[group('lint')]
lint *ARGS:
    golangci-lint run --timeout=5m {{ARGS}}

[group('lint')]
fmt:
    gofumpt -w .
    golines -w --max-len=100 .

[group('lint')]
tidy:
    go mod tidy

[group('lint')]
vet:
    go vet ./...

# =============================================================================
# Testing
# =============================================================================

[group('test')]
test *ARGS:
    go test -race ./... {{ARGS}}

[group('test')]
test-v *ARGS:
    go test -race -v ./... {{ARGS}}

[group('test')]
cover:
    go test -race -cover ./...

[group('test')]
cover-html:
    go test -race -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Wrote coverage.html"

# =============================================================================
# CI / Quality
# =============================================================================

[group('ci')]
ci: lint test
    @echo "All checks passed."

[group('ci')]
check: vet test
    @echo "Quick check passed."

# =============================================================================
# Development
# =============================================================================

[group('dev')]
run *ARGS:
    go run . {{ARGS}}

[group('dev')]
dev FILE:
    go run . {{FILE}} --stdout

# =============================================================================
# Build
# =============================================================================

[group('build')]
build:
    go build -ldflags="{{ldflags}}" -o bin/test-audit .
    @echo "Built: bin/test-audit ($(du -h bin/test-audit | cut -f1))"

[group('build')]
install:
    go install -ldflags="{{ldflags}}" .

[group('build')]
build-all:
    @mkdir -p bin
    GOOS=linux  GOARCH=amd64 go build -ldflags="{{ldflags}}" -o bin/test-audit_linux_amd64  .
    GOOS=linux  GOARCH=arm64 go build -ldflags="{{ldflags}}" -o bin/test-audit_linux_arm64  .
    GOOS=darwin GOARCH=amd64 go build -ldflags="{{ldflags}}" -o bin/test-audit_darwin_amd64 .
    GOOS=darwin GOARCH=arm64 go build -ldflags="{{ldflags}}" -o bin/test-audit_darwin_arm64 .
    @ls -lh bin/

# =============================================================================
# Release
# =============================================================================
# `just release v0.2.0`   cut a release at the version you give
# `just release-patch`    auto-bump patch  (0.1.2 -> 0.1.3)
# `just release-minor`    auto-bump minor  (0.1.2 -> 0.2.0)
# `just release-major`    auto-bump major  (0.1.2 -> 1.0.0)
# Each pushes the tag, triggering the GitHub Actions release workflow.
# =============================================================================

[group('release')]
release VERSION:
    #!/usr/bin/env bash
    set -euo pipefail
    VERSION="{{VERSION}}"
    if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z0-9.-]+)?$ ]]; then
        echo "error: version must look like v1.2.3 (got: $VERSION)" >&2
        exit 1
    fi
    if [[ -n "$(git status --porcelain)" ]]; then
        echo "error: working tree is dirty — commit or stash first" >&2
        git status --short >&2
        exit 1
    fi
    if git rev-parse "$VERSION" >/dev/null 2>&1; then
        echo "error: tag $VERSION already exists" >&2
        exit 1
    fi
    BRANCH=$(git rev-parse --abbrev-ref HEAD)
    echo ""
    echo "  releasing $VERSION from $BRANCH"
    echo ""
    git push origin "$BRANCH"
    git tag -a "$VERSION" -m "$VERSION"
    git push origin "$VERSION"
    echo ""
    echo "  tag pushed. workflow now building binaries..."
    if command -v gh &>/dev/null; then
        sleep 3
        RUN_ID=$(gh run list --workflow=release.yml --limit=1 --json databaseId -q '.[0].databaseId')
        gh run watch "$RUN_ID" --exit-status
        echo ""
        gh release view "$VERSION" 2>/dev/null | head -10 || true
    else
        echo "  install 'gh' to auto-watch, or check the Actions tab."
    fi

[group('release')]
release-patch:
    @just release "$(just _next-version patch)"

[group('release')]
release-minor:
    @just release "$(just _next-version minor)"

[group('release')]
release-major:
    @just release "$(just _next-version major)"

[group('release')]
release-dry:
    @echo "would release: $(just _next-version patch)"
    @git log --oneline "$(git describe --tags --abbrev=0 2>/dev/null || git rev-list --max-parents=0 HEAD)..HEAD"

_next-version BUMP:
    #!/usr/bin/env bash
    set -euo pipefail
    LAST=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    SEMVER=${LAST#v}
    IFS=. read -r MAJ MIN PAT <<< "$SEMVER"
    case "{{BUMP}}" in
        major) echo "v$((MAJ+1)).0.0" ;;
        minor) echo "v${MAJ}.$((MIN+1)).0" ;;
        patch) echo "v${MAJ}.${MIN}.$((PAT+1))" ;;
        *) echo "error: unknown bump {{BUMP}}" >&2; exit 1 ;;
    esac

# =============================================================================
# Utilities
# =============================================================================

[group('util')]
info:
    @echo "Project:  {{project}}"
    @echo "Version:  {{version}}"
    @echo "Go:       $(go version | cut -d' ' -f3)"

[group('util')]
clean:
    -rm -rf bin/ coverage.out coverage.html
    @echo "Cleaned build artifacts."
