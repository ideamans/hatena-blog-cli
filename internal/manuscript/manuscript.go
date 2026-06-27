// Package manuscript は frontmatter 付きの原稿フォーマットを扱います。
//
// 原稿は「YAML frontmatter（メタデータ）+ 本文」で構成されます。本文は変換せず、
// はてなMarkdown等のコンテンツをそのまま保持します（frontmatterはメタデータ表現専用）。
//
//	---
//	title: 記事タイトル
//	draft: true
//	categories: [テスト, Markdown]
//	content_type: markdown
//	---
//	本文（はてなMarkdownそのまま）
//
// この層は hatena パッケージに依存しません。Entryとの相互変換は呼び出し側で行います。
package manuscript

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// FrontMatter は原稿のメタデータです。Entryの各フィールドに対応します。
// Draft はポインタにして「未指定」と false を区別します。
type FrontMatter struct {
	Title       string   `yaml:"title,omitempty"`
	Draft       *bool    `yaml:"draft,omitempty"`
	Categories  []string `yaml:"categories,omitempty"`
	ContentType string   `yaml:"content_type,omitempty"`
	Summary     string   `yaml:"summary,omitempty"`
	Updated     string   `yaml:"updated,omitempty"`
	EditURL     string   `yaml:"edit_url,omitempty"`

	// 以下は取得時に書き出される読み取り専用情報（更新時は使用しない）。
	ID        string `yaml:"id,omitempty"`
	PageURL   string `yaml:"page_url,omitempty"`
	Published string `yaml:"published,omitempty"`
}

// Manuscript は解析済みの原稿です。
type Manuscript struct {
	Front    FrontMatter
	Body     string
	HasFront bool // frontmatterが存在したか
}

// Parse は原稿データを frontmatter と本文に分解します。
// 先頭が "---" 行で始まらない場合は frontmatter なしとみなし、全体を本文とします。
func Parse(data []byte) (*Manuscript, error) {
	s := strings.TrimPrefix(string(data), "\ufeff") // UTF-8 BOM除去

	first, rest, hasNL := splitLine(s)
	if !hasNL || strings.TrimRight(first, "\r") != "---" {
		// frontmatterなし
		return &Manuscript{Body: s}, nil
	}

	// 終端の "---"（または "..."）行を探す
	var fmLines []string
	body := ""
	found := false
	remaining := rest
	for {
		line, next, ok := splitLine(remaining)
		trimmed := strings.TrimRight(line, "\r")
		if trimmed == "---" || trimmed == "..." {
			body = next
			found = true
			break
		}
		fmLines = append(fmLines, line)
		if !ok {
			break
		}
		remaining = next
	}
	if !found {
		return nil, fmt.Errorf("frontmatterの終端 '---' が見つかりません")
	}

	var fm FrontMatter
	if err := yaml.Unmarshal([]byte(strings.Join(fmLines, "\n")), &fm); err != nil {
		return nil, fmt.Errorf("frontmatterの解析に失敗しました: %w", err)
	}
	// 本文先頭の余分な空行を1つだけ取り除く
	body = strings.TrimPrefix(body, "\n")
	return &Manuscript{Front: fm, Body: body, HasFront: true}, nil
}

// Render は frontmatter と本文から原稿テキストを生成します。
func Render(fm FrontMatter, body string) ([]byte, error) {
	y, err := yaml.Marshal(&fm)
	if err != nil {
		return nil, fmt.Errorf("frontmatterのシリアライズに失敗しました: %w", err)
	}
	var b bytes.Buffer
	b.WriteString("---\n")
	b.Write(y)
	b.WriteString("---\n\n")
	b.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		b.WriteString("\n")
	}
	return b.Bytes(), nil
}

// splitLine は s を最初の改行で2分割します。改行が含まれていれば hasNL=true。
func splitLine(s string) (line, rest string, hasNL bool) {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i], s[i+1:], true
	}
	return s, "", false
}
