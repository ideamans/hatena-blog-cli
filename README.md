# hatena-blog-cli

はてなブログを操作するコマンドラインツール（Go製）。

はてなブログの [AtomPub API](https://developer.hatena.ne.jp/ja/documents/blog/apis/atom) を通じて、記事の投稿・取得・更新・削除を行います。

## インストール

```bash
go install github.com/ideamans/hatena-blog-cli/cmd/hatena-blog@latest
```

または、リポジトリをクローンしてビルド:

```bash
go build -o hatena-blog ./cmd/hatena-blog
```

## 認証情報

以下の3点が必要です。

| 項目       | 説明                                         |
| ---------- | -------------------------------------------- |
| はてなID   | ブログ所有者のユーザー名                      |
| ブログID   | ブログの識別子（例: `example.hatenablog.jp`） |
| APIキー    | はてなブログの「詳細設定」ページで発行        |

認証情報は以下の優先順位で解決されます。

1. 環境変数
   - `HATENA_BLOG_HATENA_ID`
   - `HATENA_BLOG_ID`
   - `HATENA_BLOG_API_KEY`
2. 設定ファイル `~/.config/hatena-blog`（JSON、パーミッション600で保存）

設定ファイルのパスは `HATENA_BLOG_CONFIG` で変更できます。

### 初期設定

対話的に設定して `~/.config/hatena-blog` に保存します（保存前にAPI疎通確認を行います）。

```bash
hatena-blog auth login
```

フラグでの指定も可能です。

```bash
hatena-blog auth login \
  --hatena-id myname \
  --blog-id example.hatenablog.jp \
  --api-key xxxxxxxxxxxx
```

状態の確認 / 削除:

```bash
hatena-blog auth status            # 設定状態を表示
hatena-blog auth status --verify   # API疎通も確認
hatena-blog auth logout            # 設定ファイルを削除
```

## 使い方

### 記事一覧

```bash
hatena-blog entry list                # 最新20件
hatena-blog entry list --limit 0      # 全件
hatena-blog entry list --format json  # JSON出力
```

### 記事の取得

`list` に表示される編集URLを指定します。

```bash
hatena-blog entry get "https://blog.hatena.ne.jp/<id>/<blog>/atom/entry/<entry-id>/" --content
```

### 記事の投稿

```bash
# 本文を直接指定して下書き投稿
hatena-blog entry create --title "テスト記事" --content "本文です" --draft

# ファイル（Markdown）から本文を読み込み、カテゴリを付与して公開
hatena-blog entry create \
  --title "Go入門" \
  --file article.md \
  --content-type markdown \
  --category 技術 --category Go

# 標準入力から本文を読み込む
cat article.md | hatena-blog entry create --title "記事" --file -
```

主なオプション:

| フラグ           | 説明                                            |
| ---------------- | ----------------------------------------------- |
| `--title`        | 記事タイトル（必須）                            |
| `--content`      | 本文                                            |
| `--file`         | 本文を読み込むファイル（`-` で標準入力）        |
| `--category`     | カテゴリ（複数指定可）                          |
| `--draft`        | 下書きとして保存                                |
| `--content-type` | `markdown`（既定） / `hatena` / `html` / `plain` |
| `--updated`      | 更新日時（RFC3339、例 `2026-06-27T10:00:00+09:00`） |
| `--summary`      | 概要                                            |

### 原稿フォーマット（frontmatter + 本文）

`--file` に渡すファイルが frontmatter（先頭の `---` で囲んだYAML）で始まる場合、
タイトル・カテゴリ・下書き等のメタデータをそこから読み取ります。**本文は変換せず
そのまま投稿される**ので、はてなMarkdownの記法（`[tex:]`・`[:contents]`・`:embed:cite`
など）もそのまま書けます。frontmatterはメタデータ表現専用です。

```markdown
---
title: 記事タイトル
draft: true
categories: [テスト, Markdown]
content_type: markdown
summary: 概要（任意）
updated: 2026-06-27T10:00:00+09:00   # 任意
edit_url: https://blog.hatena.ne.jp/.../atom/entry/123/   # 更新時のみ。pullで自動付与
---
本文（はてなMarkdownそのまま）
```

```bash
# メタデータはfrontmatterから。フラグ指定は不要
hatena-blog entry create --file article.md

# フラグはfrontmatterより優先（この場合は公開で投稿）
hatena-blog entry create --file article.md --published
```

frontmatterと既存フィールド／XMLの対応:

| frontmatter | 対応する記事フィールド | XML |
| ----------- | ---------------------- | --- |
| `title` | タイトル | `<title>` |
| `draft` | 下書き | `<app:draft>` |
| `categories` | カテゴリ | `<category term>` |
| `content_type` | 本文形式 | `<content type>` |
| `summary` | 概要 | `<summary>` |
| `updated` | 更新日時 | `<updated>` |
| `edit_url` | 編集URL | link rel=edit |
| `id` / `page_url` / `published` | （読み取り専用、pullで付与） | — |

優先順位は **コマンドラインフラグ > frontmatter > 既定値** です。
frontmatterなしの素のファイル／`--content` 指定も従来どおり動作します。

### 記事の更新

編集URLを指定します。指定しなかった項目は既存の値を引き継ぎます（部分更新）。

```bash
# タイトルだけ変更
hatena-blog entry update "<編集URL>" --title "新しいタイトル"

# 下書きを公開に
hatena-blog entry update "<編集URL>" --published

# 本文を差し替え
hatena-blog entry update "<編集URL>" --file updated.md

# frontmatter付き原稿で更新（編集URLは frontmatter の edit_url から解決）
hatena-blog entry update --file article.md
```

### 記事の書き出し（往復編集）

既存記事を frontmatter 付き原稿ファイルに書き出します。ローカルで編集して
`entry update --file` で再投稿する、Git管理しやすいワークフローに使えます。

```bash
hatena-blog entry pull "<編集URL>" -o article.md   # ファイルへ
hatena-blog entry pull "<編集URL>"                  # 標準出力へ
# → article.md を編集 → hatena-blog entry update --file article.md
```

### 記事の削除

```bash
hatena-blog entry delete "<編集URL>"        # 確認プロンプトあり
hatena-blog entry delete "<編集URL>" --force # 確認なし
```

### カテゴリ集計

全記事を走査し、使用カテゴリと記事数を集計します。

```bash
hatena-blog categories
```

## LLMエージェント向けガイド

LLMエージェントがこのCLI単体ではてなブログ投稿を完結できるよう、本文フォーマット
（Markdown / はてな記法 / HTML）・コマンド・推奨ワークフロー・内部XMLまでを網羅した
詳細ガイドを内蔵しています。

```bash
hatena-blog --llm
```

`--llm` はどのサブコマンドに付けても全文ガイドを表示して終了します
（例: `hatena-blog entry create --llm`）。

## 出力フォーマット

すべてのコマンドで `--format` を指定できます。

- `table`（既定）— 人間向けの表形式
- `json` — プログラム連携向け（本文を含む全フィールド）

## 開発

```bash
go build ./...
go test ./...
go vet ./...
```

### プロジェクト構成

```
cmd/hatena-blog/       エントリポイント
internal/
  cmd/                 Cobraコマンド定義（auth, entry, categories）
  config/              認証情報の読み書き（環境変数 / ~/.config/hatena-blog）
  hatena/              AtomPub APIクライアント（WSSE認証・XML処理）
  manuscript/          frontmatter付き原稿の解析・生成
  output/              出力フォーマット（table / json）
```

## ライセンス

MIT
