package cmd

import (
	"github.com/ideamans/go-llm-cli-kit/llmcmd"

	"github.com/ideamans/hatena-blog-cli/internal/llmdocs"
)

// LLMConfig は `hatena-blog llm` サブコマンドの設定です。
// main からも参照し、非推奨の --llm フラグをコマンドラインのどの位置でも
// 受け付けるために使います。
func LLMConfig() llmcmd.Config {
	return llmcmd.Config{
		Docs:  llmdocs.Docs(),
		Short: "AIエージェント向けの詳細リファレンスを表示する",
	}
}
