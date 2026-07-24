# CLAUDE.md — hatena-blog-cli

はてなブログを AtomPub API 経由で操作する CLI。**バイナリ名は `hatena-blog`**
（リポジトリ名とリリースアーカイブ名は `hatena-blog-cli`）。

## 変更時の必須手順

**機能を追加した、フラグを増やした、既存の挙動を変えた — このいずれかをしたら、
3か所すべてを更新してから終わること。**

| 更新先 | 対象 | やり方 |
| --- | --- | --- |
| ① ドキュメント | `README.md` | 使い方が変わったときのみ |
| ② ヘルプ | cobra の `Short` / `Long` / フラグ説明 | コード内。**カタログはここから生成される** |
| ③ **LLMナレッジ** | `internal/llmdocs/00-guide.md` | 鉄則・認証・ワークフロー・エラー処理が変わったら |
| | `internal/llmdocs/10-content-format.md` | **`--content-type` の対応形式や記法を変えたら必ず** |
| | `internal/llmdocs/20-manuscript.md` | frontmatter のキーを増減したら |
| | `internal/llmdocs/80-internal-xml.md` | 生成する AtomPub XML の構造を変えたら |
| | `internal/llmdocs/90-commands.md` | **生成物。手編集しない** → `go generate ./...` |
| | `plugins/hatena-blog-cli/skills/*/SKILL.md` | 手順や前提が変わったとき |
| | `context7.json` の `rules` | 新しい落とし穴が生まれたとき |

③ を忘れやすい。ドキュメントとヘルプは人間が読んで気づくが、**LLMナレッジが
古いことには誰も気づかない**（エージェントが黙って間違えるだけ）。

判断に迷ったときの目安:

- **本文の扱い（エスケープ・content-type・記法）を変えた** → `10-content-format.md`
  は必須。ここがこの CLI の核心で、エージェントが最も間違えるところ
- frontmatter のキーを増やした → `20-manuscript.md`
- 破壊的操作や公開の挙動を変えた → `hatena-blog-usage` の SKILL.md。
  **公開は即座に読者に見える**ので、下書き優先のフローと事前同意の手順を保つこと
- 新しいサブコマンドを足した → ②を書いてから `go generate ./...`

## リリース

`PluginVersion`（`internal/cmd/root.go`）と `plugin.json` の `version` と git タグ
の3つを揃える。テストとリリースワークフローが不一致を検出する。手順は
`plugins/hatena-blog-cli/PUBLISH.md`。

## 既知の問題

`internal/config` の `TestLoadEnvOverridesFile` は、開発者の実環境に
`HATENA_BLOG_*` 環境変数や `~/.config/hatena-blog` があるとローカルで失敗する
（テストが実設定を読んでしまう）。CI では発生しない。このリポジトリ既存の問題。

## 確認

```bash
go generate ./...     # 生成物を作り直す
git diff --exit-code  # 差分が出たらコミット漏れ
go test ./...         # SKILL.md 検証とバージョン整合を含む
go run ./cmd/hatena-blog llm | head
```

## 参照

- 標準: <https://github.com/ideamans/go-llm-cli-kit/blob/main/LLM.md>
- 生成物と原本の対応: `.claude/rules/ai-artifacts-policy.md`
- 再生成: `/regen-ai`
