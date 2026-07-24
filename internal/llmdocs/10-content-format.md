# 本文フォーマット（--content-type）— ここが核心

本文の「形式」を --content-type で指定します。あなたが用意する本文は、
指定した形式のプレーンテキストです（XMLでもエスケープ済みでもありません）。

  値（別名）          実際のMIMEタイプ        説明
  ------------------  ----------------------  --------------------------------
  markdown (md)       text/x-markdown         既定。推奨。下記参照。
  hatena              text/x-hatena-syntax    はてな記法。
  html                text/html               HTMLをそのまま。
  plain (text)        text/plain              整形なしの素のテキスト。

CLIの既定は markdown です。迷ったら markdown を使ってください。

## Markdown（--content-type markdown）推奨

一般的なMarkdownに加え、はてな独自の埋め込み記法も本文中に書けます。
本文の例（これをそのまま --file に渡すファイルの中身にします）:

  # 見出し1（記事タイトルとは別。本文中の大見出し）
  ## 見出し2

  段落です。**太字**、*斜体*、`インラインコード`、[リンク](https://example.com)。

  - 箇条書き1
  - 箇条書き2

  1. 番号付き1
  2. 番号付き2

  > 引用文

  ```go
  // コードブロック（言語指定でシンタックスハイライト）
  fmt.Println("hello")
  ```

  ![代替テキスト](https://example.com/image.png)

はてな独自の埋め込み記法（Markdown本文中にそのまま書ける）:

  [https://example.com/:embed:cite]     URLのブログカード埋め込み
  [https://youtu.be/xxxx:embed]         YouTube等の埋め込み
  [tex:x^2 + y^2]                       数式（TeX記法）
  [:contents]                           目次の自動生成
  (()) 内に脚注を書く: 本文((これは脚注です))。

## はてな記法（--content-type hatena）

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

## HTML（--content-type html）/ plain

html: <p>…</p> などのHTMLをそのまま本文に書きます。CLIがXMLエスケープを
行うため、あなたはHTMLタグを生のまま渡してOKです（二重エスケープしない）。

plain: 整形なし。改行はそのまま反映されます。
