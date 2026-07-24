package main

import (
	"fmt"
	"os"

	"github.com/ideamans/go-llm-cli-kit/llmcmd"

	"github.com/ideamans/hatena-blog-cli/internal/cmd"
)

// ビルド時に goreleaser の ldflags で上書きされるバージョン情報。
var (
	version = cmd.PluginVersion
	commit  = ""
	date    = ""
)

func main() {
	// --llm はコマンドラインのどの位置にあってもリファレンスを表示して終了する。
	// `hatena-blog llm` へ移行したが、既存の呼び出しを壊さないため互換を維持する。
	if handled, err := llmcmd.HandleLegacy(os.Args[1:], cmd.LLMConfig(), os.Stdout); handled {
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
			os.Exit(1)
		}
		return
	}

	cmd.SetVersion(version, commit, date)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}
}
