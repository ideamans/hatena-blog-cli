// Command gen-llmdocs regenerates the generated chapters of the embedded LLM
// reference. It is run by go generate, and re-run in CI to prove the committed
// artifact still matches the command tree.
package main

import (
	"fmt"
	"os"

	"github.com/ideamans/go-llm-cli-kit/catalog"

	"github.com/ideamans/hatena-blog-cli/internal/cmd"
)

const outPath = "90-commands.md"

func main() {
	md := catalog.Markdown(cmd.Root(), catalog.Options{
		Title: "コマンドカタログ",
		Intro: "cobra のコマンド定義から `go generate ./...` で生成しています。\n" +
			"手編集しないこと。直すのはコマンド定義側です。",
		// llm は 00-guide.md で説明済み。
		Skip: []string{"llm"},
	})

	if err := os.WriteFile(outPath, []byte(md), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "gen-llmdocs:", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d bytes)\n", outPath, len(md))
}
