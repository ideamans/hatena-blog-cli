# Publishing the hatena-blog-cli plugin

## Before every release

1. `go generate ./...` — regenerate `internal/llmdocs/90-commands.md`, commit
   any diff.
2. `go test ./...` — includes `TestPluginSkills`, which enforces that
   `plugin.json.version` equals `PluginVersion` in `internal/cmd/root.go` and that
   the SKILL.md frontmatter stays within the Agent Skills standard.
3. `claude plugin validate plugins/hatena-blog-cli`.
4. Bump `PluginVersion` and `plugin.json.version` together, in the same commit
   as the release tag. The release workflow refuses a mismatched tag.

## Registering in the marketplace (first release only)

Add to `.claude-plugin/marketplace.json` in `ideamans/claude-public-plugins`:

```json
{
  "name": "hatena-blog-cli",
  "source": {
    "source": "git-subdir",
    "url": "https://github.com/ideamans/hatena-blog-cli.git",
    "path": "plugins/hatena-blog-cli"
  }
}
```

## Verifying the published result

```
/plugin marketplace add ideamans/claude-public-plugins
/plugin install hatena-blog-cli@ideamans-plugins
/hatena-blog-usage
```

```bash
gh skill install ideamans/hatena-blog-cli/plugins/hatena-blog-cli/skills/hatena-blog-usage --agent copilot
```
