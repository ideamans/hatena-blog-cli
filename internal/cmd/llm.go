package cmd

// llmGuide は `hatena-blog --llm` で表示される、LLMエージェント向けの詳細な利用ガイドです。
// このCLI単体ではてなブログへの投稿・管理を完結できるよう、本文フォーマット・コマンド・
// ワークフロー・内部XMLまでを網羅します。
//
// Goのraw文字列リテラル内ではバックチック(`)を表現できないため、コードフェンス等が
// 必要な箇所は通常の文字列リテラルとの連結で組み立てています。
const fence = "```"

const llmGuide = `# hatena-blog CLI — LLMエージェント向け利用ガイド

あなた（LLMエージェント）は、この単一のCLI ` + "`hatena-blog`" + ` だけではてなブログの
記事を自在に投稿・取得・更新・削除できます。このガイドは、そのために必要な知識を
すべて含みます。

================================================================
0. 最重要の鉄則（必ず守ること）
================================================================

1. 出力を機械処理するときは必ず ` + "`--format json`" + ` を付ける。
   table形式は人間用で、解析には向きません。

2. XMLは自分で書かない。
   あなたはタイトル・本文・カテゴリなどの「素材」をフラグで渡すだけです。
   AtomPub XMLの生成・エスケープ（&, <, >, " など）はCLIが自動で行います。
   （内部で生成されるXMLの実体は「8. 内部XML」を参照。仕組みの理解用であって、
   あなたが組み立てる必要はありません。）

3. 長い本文・記号やコードを含む本文は、--content ではなく --file かパイプで渡す。
   シェルのクォート崩れや改行の問題を避けられます。
     - ファイルから:        --file article.md
     - 標準入力から:        cat article.md | hatena-blog entry create --title "…" --file -

4. 記事の「住所」は編集URL（edit_url）。
   取得・更新・削除はすべて編集URLで対象を指定します。投稿/一覧/取得の
   JSON出力に含まれる "edit_url" を保持して使い回してください。

5. 破壊的操作（delete）は --force を付けない限り確認プロンプトで止まります。
   非対話環境では明示的に --force を付けてください。

================================================================
1. 認証（事前準備）
================================================================

以下の3点が必要です。環境変数が最優先で参照されます。

  HATENA_BLOG_HATENA_ID   はてなID（ユーザー名）
  HATENA_BLOG_ID          ブログID（例: example.hatenablog.jp）
  HATENA_BLOG_API_KEY     APIキー（はてなブログ「詳細設定」で発行）

未設定だと各コマンドは認証エラーで終了します。疎通確認:

  hatena-blog auth status --verify --format json

  → "verify":"成功" なら投稿可能な状態です。

================================================================
2. 本文フォーマット（--content-type）— ここが核心
================================================================

本文の「形式」を --content-type で指定します。あなたが用意する本文は、
指定した形式のプレーンテキストです（XMLでもエスケープ済みでもありません）。

  値（別名）          実際のMIMEタイプ        説明
  ------------------  ----------------------  --------------------------------
  markdown (md)       text/x-markdown         既定。推奨。下記参照。
  hatena              text/x-hatena-syntax    はてな記法。
  html                text/html               HTMLをそのまま。
  plain (text)        text/plain              整形なしの素のテキスト。

CLIの既定は markdown です。迷ったら markdown を使ってください。

----------------------------------------------------------------
2-1. Markdown（--content-type markdown）推奨
----------------------------------------------------------------

一般的なMarkdownに加え、はてな独自の埋め込み記法も本文中に書けます。
本文の例（これをそのまま --file に渡すファイルの中身にします）:

  # 見出し1（記事タイトルとは別。本文中の大見出し）
  ## 見出し2

  段落です。**太字**、*斜体*、` + "`インラインコード`" + `、[リンク](https://example.com)。

  - 箇条書き1
  - 箇条書き2

  1. 番号付き1
  2. 番号付き2

  > 引用文

  ` + fence + `go
  // コードブロック（言語指定でシンタックスハイライト）
  fmt.Println("hello")
  ` + fence + `

  ![代替テキスト](https://example.com/image.png)

はてな独自の埋め込み記法（Markdown本文中にそのまま書ける）:

  [https://example.com/:embed:cite]     URLのブログカード埋め込み
  [https://youtu.be/xxxx:embed]         YouTube等の埋め込み
  [tex:x^2 + y^2]                       数式（TeX記法）
  [:contents]                           目次の自動生成
  (()) 内に脚注を書く: 本文((これは脚注です))。

----------------------------------------------------------------
2-2. はてな記法（--content-type hatena）
----------------------------------------------------------------

はてなブログ独自のWiki風記法です。本文の例:

  * 大見出し
  ** 中見出し
  *** 小見出し

  - 箇条書き1
  - 箇条書き2
  -- ネストした箇条書き

  + 番号付き1
  + 番号付き2

  >> 引用ブロック <<

  >|go|
  fmt.Println("コードブロック（言語指定）")
  ||<

  [https://example.com/:title]          リンク（タイトル自動取得）
  [https://example.com/:embed]          埋め込み
  [tex:x^2+y^2]                         数式

----------------------------------------------------------------
2-3. HTML（--content-type html）/ plain
----------------------------------------------------------------

html: <p>…</p> などのHTMLをそのまま本文に書きます。CLIがXMLエスケープを
行うため、あなたはHTMLタグを生のまま渡してOKです（二重エスケープしない）。

plain: 整形なし。改行はそのまま反映されます。

================================================================
3. コマンドリファレンス
================================================================

すべて --format json で機械可読な出力が得られます。

--- 投稿: entry create（別名 post） ---
  必須: --title、本文（--content か --file）
  任意: --category（複数可）, --draft, --content-type, --updated, --summary
  例（下書きをMarkdownで作成）:
    hatena-blog entry create \
      --title "記事タイトル" \
      --file body.md \
      --content-type markdown \
      --category 技術 --category Go \
      --draft \
      --format json
  出力(JSON)の主なキー:
    id, title, draft, categories, content_type, content,
    page_url, edit_url, published, updated
  → edit_url を必ず保持すること。

--- 一覧: entry list ---
  hatena-blog entry list --limit 20 --format json    # 最新20件
  hatena-blog entry list --limit 0  --format json     # 全件（自動ページ送り）
  出力はエントリJSONの配列。各要素に edit_url を含む。

--- 取得: entry get ---
  hatena-blog entry get "<edit_url>" --format json
  本文(content)を含む全フィールドを返す。

--- 更新: entry update（部分更新） ---
  指定したフラグの項目だけ変更し、未指定項目は現状維持。
    本文だけ差し替え:   hatena-blog entry update "<edit_url>" --file new.md --format json
    タイトルだけ変更:   hatena-blog entry update "<edit_url>" --title "新題" --format json
    下書き→公開:        hatena-blog entry update "<edit_url>" --published --format json
    公開→下書き:        hatena-blog entry update "<edit_url>" --draft --format json
    カテゴリ全置換:     hatena-blog entry update "<edit_url>" --category A --category B
  注意: --category を指定すると既存カテゴリは「全置換」されます（追記ではない）。

--- 削除: entry delete ---
  hatena-blog entry delete "<edit_url>" --force --format json
  （--force なしだと確認プロンプトで停止）

--- カテゴリ集計: categories ---
  hatena-blog categories --format json
  全記事を走査し [{"category":名前,"count":件数}, …] を返す。

================================================================
4. 典型ワークフロー（エージェント向け）
================================================================

A. 新規記事を下書き投稿してから公開する安全な手順:
   1) entry create … --draft --format json  → 返却JSONの edit_url を記録
   2) page_url や内容を確認（必要なら entry get で本文確認）
   3) 問題なければ entry update "<edit_url>" --published --format json

B. 既存記事を探して更新する手順:
   1) entry list --format json で一覧取得、title等から目的の edit_url を特定
   2) entry get "<edit_url>" --format json で現在の本文を取得
   3) 本文を編集してファイルに保存し、
      entry update "<edit_url>" --file edited.md --format json

================================================================
5. updated（日時）の扱い
================================================================

--updated はRFC3339形式（例: 2026-06-27T10:00:00+09:00）。
通常は指定不要です。create時に未指定なら投稿時刻、update時に未指定なら
サーバー側が日時を更新します。日付を意図的に固定したい場合のみ指定します。

================================================================
6. エラーとリトライ
================================================================

エラーは stderr に「エラー: …」形式、終了コード非0で出ます。
  - HTTP 401 Invalid login → 認証情報（特にAPIキー/はてなID）の誤り。
  - 未対応のコンテンツタイプ → --content-type の値を確認。
ネットワーク起因の一時失敗は数百ミリ秒おいて再試行して構いません。
一覧の全件取得時はサーバー配慮のため250msのウェイトが自動で入ります。

================================================================
7. やってはいけないこと
================================================================

- 本文をXMLやHTMLエンティティで事前エスケープしない（二重エスケープになる）。
- edit_url を推測で組み立てない。必ずAPI出力の edit_url を使う。
- カテゴリ追記のつもりで update --category を使わない（全置換になる）。
  追記したい場合は get で現在のカテゴリを取得し、合算して渡す。

================================================================
8. 内部XML（理解のための参考。あなたが書く必要はありません）
================================================================

create/update時、CLIはあなたが渡した素材から次のようなAtomPub XMLを生成し、
WSSE認証付きで https://blog.hatena.ne.jp/<id>/<blog>/atom/entry へ送信します。

  <?xml version="1.0" encoding="utf-8"?>
  <entry xmlns="http://www.w3.org/2005/Atom" xmlns:app="http://www.w3.org/2007/app">
    <title>記事タイトル</title>
    <author><name>あなたのはてなID</name></author>
    <content type="text/x-markdown"># 本文…</content>
    <summary type="text">概要（任意）</summary>
    <category term="技術" />
    <category term="Go" />
    <app:control><app:draft>yes</app:draft></app:control>
  </entry>

ポイント:
  - <content> の type 属性が --content-type の値に対応。
  - <app:draft> は yes=下書き / no=公開。--draft / --published に対応。
  - <category term="…"> が --category に対応（複数可）。
  - タイトルや本文中の特殊文字（& < > "）はCLIが自動でXMLエスケープ。
  - 取得時は上記に加え id, link rel="edit"(=edit_url), link rel="alternate"
    (=page_url), published, updated などが返ります。

以上。これらを用いれば、このCLI単体ではてなブログ運用を完結できます。
`
