package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ideamans/go-llm-cli-kit/llmcmd"

	"github.com/ideamans/hatena-blog-cli/internal/llmdocs"
)

func TestEmbeddedReference(t *testing.T) {
	g, err := llmdocs.Docs().Markdown()
	if err != nil {
		t.Fatalf("Markdown: %v", err)
	}
	for _, want := range []string{
		"AIエージェント向けリファレンス",
		"HATENA_BLOG_API_KEY", // 認証
		"edit_url",            // 鉄則
		"本文フォーマット",            // content-type 章
		"原稿フォーマット",            // frontmatter 章
		"内部XML",               // 参考章
		"コマンドカタログ",            // 生成物
		"hatena-blog entry create",
	} {
		if !strings.Contains(g, want) {
			t.Errorf("埋め込みリファレンスに %q がありません", want)
		}
	}
}

func TestChapterOrder(t *testing.T) {
	sections, err := llmdocs.Docs().Sections()
	if err != nil {
		t.Fatalf("Sections: %v", err)
	}
	var files []string
	for _, s := range sections {
		files = append(files, s.File)
	}
	want := "00-guide.md,10-content-format.md,20-manuscript.md,80-internal-xml.md,90-commands.md"
	if got := strings.Join(files, ","); got != want {
		t.Errorf("chapters = %s, want %s", got, want)
	}
}

func TestLLMSubcommand(t *testing.T) {
	var out bytes.Buffer
	root := Root()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"llm"})
	if err := root.Execute(); err != nil {
		t.Fatalf("hatena-blog llm: %v", err)
	}
	if !strings.Contains(out.String(), "コマンドカタログ") {
		t.Error("hatena-blog llm がリファレンスを出力していません")
	}
}

// TestLegacyLLMFlag は互換の約束を守る: --llm は従来どおりコマンドラインの
// どの位置でも効く。
func TestLegacyLLMFlag(t *testing.T) {
	for _, args := range [][]string{{"--llm"}, {"entry", "list", "--llm"}} {
		var out bytes.Buffer
		handled, err := llmcmd.HandleLegacy(args, LLMConfig(), &out)
		if err != nil {
			t.Fatalf("HandleLegacy(%v): %v", args, err)
		}
		if !handled {
			t.Errorf("HandleLegacy(%v) が --llm を処理しませんでした", args)
		}
		if !strings.Contains(out.String(), "HATENA_BLOG_API_KEY") {
			t.Errorf("HandleLegacy(%v) の出力が想定と違います", args)
		}
	}
}
