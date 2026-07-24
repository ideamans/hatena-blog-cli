---
name: hatena-blog-usage
description: Publish and manage Hatena Blog (はてなブログ) entries from the command line with the hatena-blog CLI — create drafts, publish, fetch, update, delete, and list entries and categories over AtomPub. Use when the user asks to post an article to Hatena Blog, edit or take down an existing entry, or work with their はてなブログ content.
license: MIT
compatibility: Requires the `hatena-blog` binary on PATH — run the hatena-blog-install skill if it is missing. Needs HATENA_BLOG_HATENA_ID, HATENA_BLOG_ID and HATENA_BLOG_API_KEY in the environment, or a config written by `hatena-blog auth login`. Publishing is immediately visible on the public blog.
allowed-tools: Bash(hatena-blog:*) Bash(jq:*) Bash(command:*) Read Write
---

# hatena-blog-usage

Drive はてなブログ through the `hatena-blog` CLI. The CLI builds the AtomPub XML
for you — your job is to supply the material, never the markup.

## 1. Confirm the tool and the credentials

```bash
command -v hatena-blog && hatena-blog --version
hatena-blog auth status --verify --format json
```

Missing binary? Run the `hatena-blog-install` skill.

`"verify":"成功"` means posting will work. A 401 means the API key or Hatena ID
is wrong — the key comes from はてなブログ's 詳細設定 page. Ask the user to set
`HATENA_BLOG_HATENA_ID`, `HATENA_BLOG_ID` and `HATENA_BLOG_API_KEY`, or run
`hatena-blog auth login`. **Never echo the API key back into the conversation.**

## 2. Read the reference

```bash
hatena-blog llm | head -60      # 鉄則, auth, workflows
hatena-blog llm | sed -n '/# 本文フォーマット/,/# 原稿フォーマット/p'
```

It is embedded in the binary (~400 lines) and matches the installed version.

## 3. Rules that decide whether the post comes out right

- **Always `--format json`** when you will parse the result. `table` is for
  humans.
- **Never write the XML, and never pre-escape the body.** Pass plain text; the
  CLI generates the AtomPub XML and escapes `& < > "`. Pre-escaping produces
  doubly-escaped entities in the published article.
- **Pass long or symbol-heavy bodies via `--file` or stdin**, not `--content` —
  shell quoting will otherwise mangle newlines and backticks.
  `cat article.md | hatena-blog entry create --title "…" --file -`
- **`edit_url` is the address of an entry.** Take it from the JSON output of
  create/list/get and reuse it. Never construct one by hand.
- **`update --category` replaces the whole category set.** To add one, `get` the
  current categories first and pass the union.
- `--content-type` defaults to `markdown`, which also accepts Hatena's embed
  syntax (`[https://…:embed:cite]`, `[tex:…]`, `[:contents]`, `(( ))` footnotes).

## 4. Publish safely

**A published entry is immediately public.** Prefer the draft-first flow:

```bash
hatena-blog entry create --title "…" --file article.md --draft --format json
# check the returned page_url / content, then:
hatena-blog entry update "<edit_url>" --published --format json
```

Confirm with the user before publishing, before updating an existing public
entry, and always before `entry delete` — deletion needs `--force` in a
non-interactive session and cannot be undone.

For round-trip editing, `entry pull "<edit_url>" -o article.md` writes a
frontmatter manuscript that `entry update --file article.md` can post back; the
`edit_url` travels in the frontmatter.

## 5. Report

Give back the `page_url` (what the reader sees) and the `edit_url` (what you
need next), plus whether it went out as a draft or published.

## Failure modes

| Symptom | Cause | Fix |
| --- | --- | --- |
| `command not found` | not installed | run the `hatena-blog-install` skill |
| `HTTP 401 Invalid login` | wrong API key or Hatena ID | re-check credentials; the key is issued in はてなブログ 詳細設定 |
| doubly-escaped entities in the article | body was pre-escaped | pass raw text and let the CLI escape |
| categories disappeared after an update | `--category` replaces the set | `get` first, pass the union |
| unsupported content type | bad `--content-type` | one of markdown / hatena / html / plain |
| delete stops and waits | confirmation prompt | pass `--force` when non-interactive, after the user agrees |
