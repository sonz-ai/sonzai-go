set shell := ["bash", "-cu"]

# List recipes
default:
    @just --list

# Full release: bump versions, test, build, commit, tag, push, gh release.
# Usage: just deploy 1.2.3
deploy VERSION:
    @just _preflight {{VERSION}}
    @just _test
    @just _bump {{VERSION}}
    @just _build
    @just _commit {{VERSION}}
    git push origin main
    @just _publish {{VERSION}}
    @just _tag {{VERSION}}
    @just _release {{VERSION}}
    @echo "✓ Released v{{VERSION}}"

_preflight VERSION:
    @just _validate-version {{VERSION}}
    @just _check-clean
    @just _check-main
    @just _check-tag-free {{VERSION}}

_validate-version VERSION:
    #!/usr/bin/env bash
    set -euo pipefail
    if ! [[ "{{VERSION}}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
      echo "error: VERSION must match X.Y.Z (got: {{VERSION}})" >&2
      exit 1
    fi

_check-clean:
    #!/usr/bin/env bash
    set -euo pipefail
    if [[ -n "$(git status --porcelain)" ]]; then
      echo "error: working tree is dirty; commit or stash first" >&2
      git status --short
      exit 1
    fi

_check-main:
    #!/usr/bin/env bash
    set -euo pipefail
    branch="$(git rev-parse --abbrev-ref HEAD)"
    if [[ "$branch" != "main" ]]; then
      echo "error: must be on main (current: $branch)" >&2
      exit 1
    fi

_check-tag-free VERSION:
    #!/usr/bin/env bash
    set -euo pipefail
    if git rev-parse --verify --quiet "v{{VERSION}}" >/dev/null; then
      echo "error: local tag v{{VERSION}} already exists" >&2
      exit 1
    fi
    git fetch origin --tags --quiet
    if git ls-remote --tags origin "refs/tags/v{{VERSION}}" | grep -q .; then
      echo "error: remote tag v{{VERSION}} already exists on origin" >&2
      exit 1
    fi

_test:
    go test ./... -count=1

_bump VERSION:
    #!/usr/bin/env bash
    set -euo pipefail
    perl -pi -e 's/const SDKVersion = "[^"]+"/const SDKVersion = "{{VERSION}}"/' http.go
    perl -pi -e 's{(github\.com/sonz-ai/sonzai-go)\@v[0-9]+\.[0-9]+\.[0-9]+}{$1\@v{{VERSION}}}g' README.md
    echo "bumped to {{VERSION}}"

_build:
    go build ./...
    go vet ./...

_commit VERSION:
    git add http.go README.md
    git commit -m "release: v{{VERSION}}"

# Go modules are published via git tag; no separate registry push.
_publish VERSION:
    @echo "Go: no registry publish step — tag v{{VERSION}} will be resolved by proxy.golang.org"

_tag VERSION:
    git tag -a v{{VERSION}} -m "Release v{{VERSION}}"
    git push origin v{{VERSION}}

_release VERSION:
    gh release create v{{VERSION}} --title "v{{VERSION}}" --generate-notes
