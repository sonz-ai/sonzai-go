# DEPLOY — sonzai-go

## The rule

**Never release manually. Always use `just patch`.**

```bash
just patch              # bump patch, test, build, commit, push, tag, gh release
just deploy 1.4.3       # same, for an explicit version
```

That runs the complete pipeline in order:

1. Preflight (version format, clean tree, on `main`, tag free)
2. `go test ./...`
3. Bump `SDKVersion` const in `http.go`
4. `go build ./...`
5. Commit `release: vX.Y.Z`
6. `git push origin main`
7. Annotated tag `vX.Y.Z` + push (Go consumers resolve by tag via
   `go get github.com/sonz-ai/sonzai-go@vX.Y.Z` — no registry to publish to)
8. `gh release create vX.Y.Z --generate-notes`

Skip the tag and `go get` can't resolve the version. Skip the gh release
and the Releases page is empty. Both must happen.

## Don't

- Don't manually edit the `SDKVersion` const in `http.go` and commit.
  `_bump` needs to run so the version is consistent with what `_tag` signs.
- Don't `git tag` manually — let `_tag` do it so the tag message matches
  what `go get` sees.
- Don't skip `gh release create` — Go doesn't have a registry, so the
  GitHub release IS the user-facing release.
- Don't bump minor/major without explicit user approval (patch is the
  default discipline on this tree).

## Recovering a half-manual release

If someone already bumped + committed + pushed + tagged but skipped the
gh release (this happened on v1.4.1 and v1.4.2), run the missing step:

```bash
just _release 1.4.2
```

Or — cleaner — skip ahead with `just patch` to `1.4.3` and let the full
pipeline run.

## See also

[`../DEPLOY.md`](../DEPLOY.md) — canonical guide covering all four repos.
