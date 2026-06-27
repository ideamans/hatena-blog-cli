package main

import (
	"fmt"
	"os"

	"github.com/ideamans/hatena-blog-cli/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}
}
