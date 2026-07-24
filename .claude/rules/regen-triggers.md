---
paths:
  - "internal/cmd/*.go"
  - "internal/llmdocs/0*.md"
  - "internal/llmdocs/1*.md"
  - "internal/llmdocs/2*.md"
  - "internal/llmdocs/8*.md"
---

# 埋め込みリファレンスの原本に触れました

コマンド・フラグ・ヘルプ文字列を変えたなら、終わる前に `/regen-ai` を実行して
`internal/llmdocs/90-commands.md` を一致させること。CI が再生成して差分が出たら
落ちます。

本文の扱い（エスケープ、`--content-type`、はてな記法）を変えたなら
`10-content-format.md` を手で直すこと。エージェントはこの章を信じて本文を組み立てる
ため、ズレると「エラーにならず記事が壊れる」形で表面化します。

`90-commands.md` は直接編集しないこと。
