package main

import (
	"fmt"
	"os"

	"github.com/ideamans/hatena-blog-cli/internal/cmd"
)

// ビルド時に goreleaser の ldflags で上書きされるバージョン情報。
var (
	version = "dev"
	commit  = ""
	date    = ""
)

func main() {
	cmd.SetVersion(version, commit, date)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}
}
