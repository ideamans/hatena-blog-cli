# hatena-blog CLI — AIエージェント向けリファレンス

あなた（AIエージェント）は、この単一のCLI `hatena-blog` だけではてなブログの
記事を投稿・取得・更新・削除できます。このリファレンスは、そのために必要な知識を
すべて含みます。

このリファレンスはバイナリに埋め込まれています。`hatena-blog llm` は常に実行中の
バージョンそのものを説明します。

## 最重要の鉄則（必ず守ること）

1. 出力を機械処理するときは必ず `--format json` を付ける。
   table形式は人間用で、解析には向きません。

2. XMLは自分で書かない。
   あなたはタイトル・本文・カテゴリなどの「素材」をフラグで渡すだけです。
   AtomPub XMLの生成・エスケープ（&, <, >, " など）はCLIが自動で行います。
   （内部で生成されるXMLの実体は「内部XML」の章を参照。仕組みの理解用であって、
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

## 認証（事前準備）

以下の3点が必要です。環境変数が最優先で参照されます。

  HATENA_BLOG_HATENA_ID   はてなID（ユーザー名）
  HATENA_BLOG_ID          ブログID（例: example.hatenablog.jp）
  HATENA_BLOG_API_KEY     APIキー（はてなブログ「詳細設定」で発行）

未設定だと各コマンドは認証エラーで終了します。疎通確認:

  hatena-blog auth status --verify --format json

  → "verify":"成功" なら投稿可能な状態です。

## 典型ワークフロー

A. 新規記事を下書き投稿してから公開する安全な手順:
   1) entry create … --draft --format json  → 返却JSONの edit_url を記録
   2) page_url や内容を確認（必要なら entry get で本文確認）
   3) 問題なければ entry update "<edit_url>" --published --format json

B. 既存記事を探して更新する手順:
   1) entry list --format json で一覧取得、title等から目的の edit_url を特定
   2) entry get "<edit_url>" --format json で現在の本文を取得
   3) 本文を編集してファイルに保存し、
      entry update "<edit_url>" --file edited.md --format json

## updated（日時）の扱い

--updated はRFC3339形式（例: 2026-06-27T10:00:00+09:00）。
通常は指定不要です。create時に未指定なら投稿時刻、update時に未指定なら
サーバー側が日時を更新します。日付を意図的に固定したい場合のみ指定します。

## エラーとリトライ

エラーは stderr に「エラー: …」形式、終了コード非0で出ます。
  - HTTP 401 Invalid login → 認証情報（特にAPIキー/はてなID）の誤り。
  - 未対応のコンテンツタイプ → --content-type の値を確認。
ネットワーク起因の一時失敗は数百ミリ秒おいて再試行して構いません。
一覧の全件取得時はサーバー配慮のため250msのウェイトが自動で入ります。

## やってはいけないこと

- 本文をXMLやHTMLエンティティで事前エスケープしない（二重エスケープになる）。
- edit_url を推測で組み立てない。必ずAPI出力の edit_url を使う。
- カテゴリ追記のつもりで update --category を使わない（全置換になる）。
  追記したい場合は get で現在のカテゴリを取得し、合算して渡す。
