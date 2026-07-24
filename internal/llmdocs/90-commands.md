# コマンドカタログ

cobra のコマンド定義から `go generate ./...` で生成しています。
手編集しないこと。直すのはコマンド定義側です。

## Global flags

| flag | type | default | description |
| --- | --- | --- | --- |
| `--format` | string | `table` | 出力フォーマット: table, json |

## `hatena-blog auth`

認証情報の管理

はてなブログの認証情報（はてなID・ブログID・APIキー）の設定・確認・削除を行います。

### `hatena-blog auth login`

認証情報を設定ファイルに保存します

はてなID・ブログID・APIキーを ~/.config/hatena-blog に保存します。

フラグで指定されなかった項目は対話的に入力を求めます。
APIキーははてなブログの「詳細設定」ページで確認できます。

| flag | type | default | description |
| --- | --- | --- | --- |
| `--api-key` | string | — | APIキー |
| `--blog-id` | string | — | ブログID（例: example.hatenablog.jp） |
| `--hatena-id` | string | — | はてなID（ユーザー名） |
| `--no-verify` | bool | `false` | 保存前のAPI疎通確認をスキップする |

### `hatena-blog auth logout`

保存された認証情報（設定ファイル）を削除します

### `hatena-blog auth status`

現在の認証情報の状態を表示します

| flag | type | default | description |
| --- | --- | --- | --- |
| `--verify` | bool | `false` | APIへの疎通確認も行う |

## `hatena-blog categories`

ブログで使われているカテゴリの一覧を表示します

全記事を走査し、使用されているカテゴリと各カテゴリの記事数を集計して表示します。

Aliases: category

## `hatena-blog entry`

記事の投稿・取得・更新・削除

はてなブログの記事（エントリー）を操作します。

Aliases: entries

### `hatena-blog entry create`

新しい記事を投稿します

新しい記事を投稿します。

本文は --content で直接指定するか、--file でファイルから読み込みます。
--file - を指定すると標準入力から読み込みます。

原稿ファイルが frontmatter（先頭の --- で囲んだYAML）で始まる場合、
タイトル・カテゴリ・下書き等のメタデータをそこから読み取ります。
コマンドラインのフラグは frontmatter より優先されます（フラグ > frontmatter > 既定）。

frontmatterの例:
  ---
  title: 記事タイトル
  draft: true
  categories: [テスト, Markdown]
  content_type: markdown
  ---
  本文（はてなMarkdownそのまま）

例:
  hatena-blog entry create --title "テスト" --content "本文" --draft
  hatena-blog entry create --file article.md          # メタデータはfrontmatterから
  hatena-blog entry create --file article.md --draft  # 下書きフラグで上書き
  cat article.md | hatena-blog entry create --file -

Aliases: post

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | stringSlice | `[]` | カテゴリ（複数指定可） |
| `--content` | string | — | 本文 |
| `--content-type` | string | `markdown` | 本文の形式: markdown, hatena, html, plain |
| `--draft` | bool | `false` | 下書きとして保存する |
| `--file` | string | — | 本文を読み込むファイル（- で標準入力） |
| `--summary` | string | — | 概要（記事の概要欄、省略可） |
| `--title` | string | — | 記事タイトル（必須） |
| `--updated` | string | — | 更新日時（RFC3339形式、省略可） |

### `hatena-blog entry delete`

記事を削除します

```
hatena-blog entry delete <編集URL>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `-f`, `--force` | bool | `false` | 確認なしで削除する |

### `hatena-blog entry get`

記事を1件取得します

編集URL（list で表示されるURL）を指定して記事を取得します。

```
hatena-blog entry get <編集URL>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--content` | bool | `false` | 本文も表示する |

### `hatena-blog entry list`

記事一覧を取得します

記事の一覧を新しい順に取得します。--limit で取得件数を制限できます（0で全件）。

| flag | type | default | description |
| --- | --- | --- | --- |
| `--limit` | int | `20` | 取得件数の上限（0で全件） |

### `hatena-blog entry pull`

記事をfrontmatter付き原稿ファイルに書き出します

編集URLの記事を取得し、frontmatter（メタデータ）+ 本文 の原稿形式で書き出します。

書き出した原稿はローカルで編集し、'entry update --file <原稿>' で再投稿できます
（編集URLはfrontmatterの edit_url から自動解決されます）。

例:
  hatena-blog entry pull "<編集URL>" -o article.md
  hatena-blog entry pull "<編集URL>"          # 標準出力へ

```
hatena-blog entry pull <編集URL>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `-o`, `--output` | string | — | 出力先ファイル（省略時は標準出力） |

### `hatena-blog entry update`

既存の記事を更新します

編集URLを指定して既存の記事を更新します。

指定しなかった項目は現在の値を引き継ぎます（部分更新）。
下書き状態は --draft で下書きに、--published で公開に変更できます。

frontmatter付きの原稿を --file で渡す場合、編集URLは引数を省略して
frontmatterの edit_url から解決できます（'entry pull' で書き出した原稿が使えます）。
優先順位は フラグ > frontmatter > 現在の値 です。

```
hatena-blog entry update [編集URL]
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--category` | stringSlice | `[]` | カテゴリ（指定すると全置換） |
| `--content` | string | — | 本文 |
| `--content-type` | string | — | 本文の形式: markdown, hatena, html, plain |
| `--draft` | bool | `false` | 下書きに変更する |
| `--file` | string | — | 本文を読み込むファイル（- で標準入力） |
| `--published` | bool | `false` | 公開状態に変更する |
| `--summary` | string | — | 概要（記事の概要欄） |
| `--title` | string | — | 記事タイトル |
| `--updated` | string | — | 更新日時（RFC3339形式、省略可） |
