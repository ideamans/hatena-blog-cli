---
name: hatena-blog-install
description: Make the hatena-blog command available, installing it only if it is missing. Use when another skill reports that `hatena-blog` is not on PATH, or when the user asks to install, update or upgrade the ideamans Hatena Blog CLI. Prefers an already-installed binary, then the latest GitHub release, then a build from source with go install.
license: MIT
compatibility: Requires curl (or wget) and tar to install from a release, or a Go toolchain for the source fallback. Standalone — does not need hatena-blog to be present already. Installs from the public repository github.com/ideamans/hatena-blog-cli, so no GitHub authentication is needed.
allowed-tools: Bash(curl:*) Bash(wget:*) Bash(tar:*) Bash(unzip:*) Bash(go:*) Bash(uname:*) Bash(command:*) Bash(which:*) Bash(mkdir:*) Bash(mv:*) Bash(cp:*) Bash(rm:*) Bash(chmod:*) Bash(ls:*) Bash(test:*) Bash(echo:*) Read
---

# hatena-blog-install

Make the `hatena-blog` command usable, doing the least work that achieves it.

## Route 1 — an existing installation on PATH

```bash
command -v hatena-blog && hatena-blog --version
```

If that resolves, **use it and stop here.** Do not check for a newer release —
it costs an API call and the user did not ask for an upgrade.

Two checks before trusting the hit:

- **It is the right tool.** `hatena-blog llm | head -1` must read
  `# hatena-blog CLI — AIエージェント向けリファレンス`.
- **It is recent enough.** If `hatena-blog llm` is not a known command, the
  binary predates the embedded reference. Say so and continue to route 2.

Continue past this section only when the command is missing, is the wrong tool,
is too old, or the user explicitly asked to update.

## Route 2 — the latest GitHub release

The repository is public, so no authentication is needed.

```bash
VERSION=$(curl -fsSL https://api.github.com/repos/ideamans/hatena-blog-cli/releases/latest \
  | grep '"tag_name"' | head -1 | cut -d'"' -f4)   # e.g. v0.2.0

OS=$(uname -s | tr '[:upper:]' '[:lower:]')            # darwin | linux
ARCH=$(uname -m); [ "$ARCH" = "x86_64" ] && ARCH=amd64  # amd64 | arm64
curl -fsSL -o /tmp/hatena-blog.tar.gz \
  "https://github.com/ideamans/hatena-blog-cli/releases/download/${VERSION}/hatena-blog-cli_${VERSION#v}_${OS}_${ARCH}.tar.gz"
```

**The archive is named `hatena-blog-cli_…` but the binary inside is
`hatena-blog`** — without the `-cli` suffix. Windows ships a `.zip`.

If the download 404s, list the actual assets on the release page rather than
retrying variations.

### Install onto PATH

```bash
tar -xzf /tmp/hatena-blog.tar.gz -C /tmp
mkdir -p ~/.local/bin && mv /tmp/hatena-blog ~/.local/bin/ && chmod +x ~/.local/bin/hatena-blog
```

Prefer the first writable directory already on PATH — `~/.local/bin`, then
`/usr/local/bin`. Two things not to do on your own initiative:

- If nothing on PATH is writable, leave the binary in `/tmp`, print the exact
  `sudo mv` command and let the user run it. Do not run `sudo` yourself.
- If `~/.local/bin` is not on PATH, give the user the line to add to their shell
  profile. Do not edit the profile for them.

## Route 3 — build from source

Needs a Go toolchain and compiles rather than downloads. Note the
`/cmd/hatena-blog` suffix; installing the module root would not build anything.

```bash
go install github.com/ideamans/hatena-blog-cli/cmd/hatena-blog@latest
```

The binary lands in `$(go env GOPATH)/bin` and is named `hatena-blog`.

## Verify

```bash
hatena-blog --version
hatena-blog llm | head -5
hatena-blog auth status
```

Report which route was taken, the version and the install path.

Then say what is still needed: nothing works without credentials. The user needs
their Hatena ID, the blog ID (e.g. `example.hatenablog.jp`) and an API key from
はてなブログ's 詳細設定 page, set as `HATENA_BLOG_HATENA_ID`, `HATENA_BLOG_ID`
and `HATENA_BLOG_API_KEY`, or entered via `hatena-blog auth login`.
