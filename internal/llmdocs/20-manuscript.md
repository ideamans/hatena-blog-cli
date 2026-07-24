# 原稿フォーマット（frontmatter + 本文）

--file に渡すファイルが frontmatter（先頭の --- で囲んだYAML）で始まる場合、
タイトル等のメタデータをそこから読み取れます。本文は変換されず、はてなMarkdown
記法（[tex:]・[:contents]・:embed:cite・(( ))脚注 等）をそのまま書けます。
あなた（エージェント）にとっては、メタデータと本文を1ファイルにまとめられる
ので扱いやすい形式です。

  ---
  title: 記事タイトル
  draft: true
  categories: [テスト, Markdown]
  content_type: markdown
  summary: 概要（任意）
  edit_url: https://blog.hatena.ne.jp/.../atom/entry/123/   # 更新時のみ
  ---
  本文（はてなMarkdownそのまま）

frontmatterのキーは記事フィールドに1:1対応:
  title→タイトル, draft→下書き, categories→カテゴリ, content_type→本文形式,
  summary→概要, updated→更新日時, edit_url→編集URL

優先順位: コマンドラインフラグ > frontmatter > 既定値。
使用例:
  hatena-blog entry create --file article.md --format json        # メタはfrontmatterから
  hatena-blog entry create --file article.md --published          # フラグが優先
  hatena-blog entry update --file article.md --format json        # edit_urlはfrontmatterから解決

往復編集ワークフロー（最も堅牢。Git管理にも向く）:
  1) entry pull "<edit_url>" -o article.md   で原稿化（edit_url等が埋め込まれる）
  2) article.md の frontmatter／本文を編集
  3) entry update --file article.md          で再投稿（編集URLは自動解決）
