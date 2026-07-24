# 内部XML（理解のための参考。あなたが書く必要はありません）

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
