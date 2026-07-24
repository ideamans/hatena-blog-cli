# 生成物 — 手編集しないこと

| 生成物 | 原本 |
| --- | --- |
| `internal/llmdocs/90-commands.md` | `internal/cmd/*.go` の cobra コマンド定義（`internal/gen-llmdocs` が描画） |

手書き（編集してよい）:

- `internal/llmdocs/00-guide.md` — 鉄則・認証・ワークフロー・エラー
- `internal/llmdocs/10-content-format.md` — 本文フォーマット（この CLI の核心）
- `internal/llmdocs/20-manuscript.md` — frontmatter 原稿フォーマット
- `internal/llmdocs/80-internal-xml.md` — 生成される AtomPub XML の参考
- `plugins/hatena-blog-cli/skills/*/SKILL.md`
- `context7.json`

生成物を直接編集しても次の `go generate ./...` で消え、それまでは CI が
古い差分で落ちる。カタログを良くしたいときはコマンドの `Short` / `Long` /
フラグ説明を直すこと。

再生成は `/regen-ai`、または `go generate ./... && go test ./...`。
