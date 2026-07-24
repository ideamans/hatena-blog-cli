// Package llmdocs は `hatena-blog llm` が出力するリファレンスを埋め込みます。
//
// 00- から 80- は手書き、90-commands.md は cobra のコマンドツリーから
// `go generate ./...` で生成してコミットします（go:embed がビルド時に実体を
// 要求するため）。CI が再生成して差分が出たら失敗します。
package llmdocs

import (
	"embed"

	kit "github.com/ideamans/go-llm-cli-kit/llmdocs"
)

//go:generate go run ../gen-llmdocs

//go:embed *.md
var files embed.FS

// Docs は埋め込みリファレンスです。
func Docs() *kit.Docs { return kit.New(files, ".") }
