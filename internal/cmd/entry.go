package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ideamans/hatena-blog-cli/internal/hatena"
	"github.com/ideamans/hatena-blog-cli/internal/output"
	"github.com/spf13/cobra"
)

func newEntryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "entry",
		Aliases: []string{"entries"},
		Short:   "記事の投稿・取得・更新・削除",
		Long:    "はてなブログの記事（エントリー）を操作します。",
	}
	cmd.AddCommand(newEntryListCmd())
	cmd.AddCommand(newEntryGetCmd())
	cmd.AddCommand(newEntryCreateCmd())
	cmd.AddCommand(newEntryUpdateCmd())
	cmd.AddCommand(newEntryDeleteCmd())
	return cmd
}

// contentTypeFromName はエイリアス名を実際のMIMEタイプに変換します。
func contentTypeFromName(name string) (string, error) {
	switch strings.ToLower(name) {
	case "markdown", "md", hatena.ContentTypeMarkdown:
		return hatena.ContentTypeMarkdown, nil
	case "hatena", "hatena-syntax", hatena.ContentTypeHatena:
		return hatena.ContentTypeHatena, nil
	case "html", hatena.ContentTypeHTML:
		return hatena.ContentTypeHTML, nil
	case "plain", "text", hatena.ContentTypePlain:
		return hatena.ContentTypePlain, nil
	default:
		return "", fmt.Errorf("未対応のコンテンツタイプです: %s (markdown, hatena, html, plain が使用可能)", name)
	}
}

func newEntryListCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "記事一覧を取得します",
		Long:  "記事の一覧を新しい順に取得します。--limit で取得件数を制限できます（0で全件）。",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := newClient()
			if err != nil {
				return err
			}
			entries, err := client.List(limit)
			if err != nil {
				return err
			}

			headers := []string{"状態", "公開日", "タイトル", "カテゴリ", "編集URL"}
			rows := make([][]string, 0, len(entries))
			for _, e := range entries {
				rows = append(rows, []string{
					draftLabel(e.Draft),
					formatDate(e.Published),
					truncate(e.Title, 40),
					strings.Join(e.Categories, ","),
					e.EditURL,
				})
			}
			return output.Print(outputFormat, entriesToJSON(entries), headers, rows)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "取得件数の上限（0で全件）")
	return cmd
}

func newEntryGetCmd() *cobra.Command {
	var showContent bool
	cmd := &cobra.Command{
		Use:   "get <編集URL>",
		Short: "記事を1件取得します",
		Long:  "編集URL（list で表示されるURL）を指定して記事を取得します。",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := newClient()
			if err != nil {
				return err
			}
			e, err := client.Get(args[0])
			if err != nil {
				return err
			}

			if outputFormat == output.FormatJSON {
				return output.Print(outputFormat, entryToJSON(e), nil, nil)
			}

			// table表示では本文以外をメタ情報として、本文は任意で表示
			headers := []string{"項目", "値"}
			rows := [][]string{
				{"タイトル", e.Title},
				{"状態", draftLabel(e.Draft)},
				{"コンテンツタイプ", e.ContentType},
				{"カテゴリ", strings.Join(e.Categories, ", ")},
				{"公開日", formatDate(e.Published)},
				{"更新日", formatDate(e.Updated)},
				{"著者", e.Author},
				{"ページURL", e.PageURL},
				{"編集URL", e.EditURL},
			}
			if err := output.Print(outputFormat, entryToJSON(e), headers, rows); err != nil {
				return err
			}
			if showContent {
				fmt.Println("\n--- 本文 ---")
				fmt.Println(e.Content)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&showContent, "content", false, "本文も表示する")
	return cmd
}

func newEntryCreateCmd() *cobra.Command {
	var title, content, file, contentType, updated, summary string
	var categories []string
	var draft bool

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"post"},
		Short:   "新しい記事を投稿します",
		Long: `新しい記事を投稿します。

本文は --content で直接指定するか、--file でファイルから読み込みます。
--file - を指定すると標準入力から読み込みます。

例:
  hatena-blog entry create --title "テスト" --content "本文" --draft
  hatena-blog entry create --title "記事" --file article.md --category 技術 --category Go
  cat article.md | hatena-blog entry create --title "記事" --file -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := resolveContent(content, file)
			if err != nil {
				return err
			}
			if title == "" {
				return fmt.Errorf("--title は必須です")
			}

			ct, err := contentTypeFromName(contentType)
			if err != nil {
				return err
			}

			e := &hatena.Entry{
				Title:       title,
				Content:     body,
				ContentType: ct,
				Categories:  categories,
				Draft:       draft,
				Summary:     summary,
			}
			if updated != "" {
				t, err := time.Parse(time.RFC3339, updated)
				if err != nil {
					return fmt.Errorf("--updated はRFC3339形式で指定してください (例: 2026-06-27T10:00:00+09:00): %w", err)
				}
				e.Updated = t
			}

			client, _, err := newClient()
			if err != nil {
				return err
			}
			created, err := client.Create(e)
			if err != nil {
				return err
			}
			return printEntryResult(created, "投稿しました")
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "記事タイトル（必須）")
	cmd.Flags().StringVar(&content, "content", "", "本文")
	cmd.Flags().StringVar(&file, "file", "", "本文を読み込むファイル（- で標準入力）")
	cmd.Flags().StringSliceVar(&categories, "category", nil, "カテゴリ（複数指定可）")
	cmd.Flags().BoolVar(&draft, "draft", false, "下書きとして保存する")
	cmd.Flags().StringVar(&contentType, "content-type", "markdown", "本文の形式: markdown, hatena, html, plain")
	cmd.Flags().StringVar(&updated, "updated", "", "更新日時（RFC3339形式、省略可）")
	cmd.Flags().StringVar(&summary, "summary", "", "概要（記事の概要欄、省略可）")
	return cmd
}

func newEntryUpdateCmd() *cobra.Command {
	var title, content, file, contentType, updated, summary string
	var categories []string
	var draft, published bool

	cmd := &cobra.Command{
		Use:   "update <編集URL>",
		Short: "既存の記事を更新します",
		Long: `編集URLを指定して既存の記事を更新します。

指定しなかった項目は現在の値を引き継ぎます（部分更新）。
下書き状態は --draft で下書きに、--published で公開に変更できます。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := newClient()
			if err != nil {
				return err
			}

			// 現在の記事を取得し、変更のあった項目だけ上書きする
			e, err := client.Get(args[0])
			if err != nil {
				return fmt.Errorf("更新対象の記事の取得に失敗しました: %w", err)
			}

			if cmd.Flags().Changed("title") {
				e.Title = title
			}
			if cmd.Flags().Changed("content") || cmd.Flags().Changed("file") {
				body, err := resolveContent(content, file)
				if err != nil {
					return err
				}
				e.Content = body
			}
			if cmd.Flags().Changed("content-type") {
				ct, err := contentTypeFromName(contentType)
				if err != nil {
					return err
				}
				e.ContentType = ct
			}
			if cmd.Flags().Changed("category") {
				e.Categories = categories
			}
			if cmd.Flags().Changed("summary") {
				e.Summary = summary
			}
			if draft {
				e.Draft = true
			}
			if published {
				e.Draft = false
			}
			if updated != "" {
				t, err := time.Parse(time.RFC3339, updated)
				if err != nil {
					return fmt.Errorf("--updated はRFC3339形式で指定してください: %w", err)
				}
				e.Updated = t
			} else {
				// 既存のupdatedをそのまま送ると過去日時になるためクリアし、サーバー側に委ねる
				e.Updated = time.Time{}
			}

			result, err := client.Update(e)
			if err != nil {
				return err
			}
			return printEntryResult(result, "更新しました")
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "記事タイトル")
	cmd.Flags().StringVar(&content, "content", "", "本文")
	cmd.Flags().StringVar(&file, "file", "", "本文を読み込むファイル（- で標準入力）")
	cmd.Flags().StringSliceVar(&categories, "category", nil, "カテゴリ（指定すると全置換）")
	cmd.Flags().BoolVar(&draft, "draft", false, "下書きに変更する")
	cmd.Flags().BoolVar(&published, "published", false, "公開状態に変更する")
	cmd.Flags().StringVar(&contentType, "content-type", "", "本文の形式: markdown, hatena, html, plain")
	cmd.Flags().StringVar(&updated, "updated", "", "更新日時（RFC3339形式、省略可）")
	cmd.Flags().StringVar(&summary, "summary", "", "概要（記事の概要欄）")
	return cmd
}

func newEntryDeleteCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <編集URL>",
		Short: "記事を削除します",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := newClient()
			if err != nil {
				return err
			}

			if !force {
				// 確認のため対象タイトルを表示
				if e, err := client.Get(args[0]); err == nil {
					fmt.Fprintf(os.Stderr, "削除対象: %s\n", e.Title)
				}
				fmt.Fprint(os.Stderr, "本当に削除しますか? [y/N]: ")
				var ans string
				fmt.Scanln(&ans)
				if !strings.EqualFold(strings.TrimSpace(ans), "y") {
					return fmt.Errorf("削除を中止しました")
				}
			}

			if err := client.Delete(args[0]); err != nil {
				return err
			}
			result := map[string]interface{}{"status": "deleted", "edit_url": args[0]}
			return output.Print(outputFormat, result, []string{"項目", "値"}, [][]string{
				{"状態", "削除しました"},
				{"編集URL", args[0]},
			})
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "確認なしで削除する")
	return cmd
}

// --- ヘルパー ---

// resolveContent は --content と --file から本文を解決します。
func resolveContent(content, file string) (string, error) {
	if file != "" {
		if file == "-" {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return "", fmt.Errorf("標準入力の読み込みに失敗しました: %w", err)
			}
			return string(data), nil
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("ファイルの読み込みに失敗しました: %w", err)
		}
		return string(data), nil
	}
	return content, nil
}

func printEntryResult(e *hatena.Entry, action string) error {
	result := entryToJSON(e)
	headers := []string{"項目", "値"}
	rows := [][]string{
		{"状態", action + "（" + draftLabel(e.Draft) + "）"},
		{"タイトル", e.Title},
		{"ページURL", e.PageURL},
		{"編集URL", e.EditURL},
	}
	return output.Print(outputFormat, result, headers, rows)
}

func entryToJSON(e *hatena.Entry) map[string]interface{} {
	return map[string]interface{}{
		"id":           e.ID,
		"title":        e.Title,
		"draft":        e.Draft,
		"categories":   e.Categories,
		"content_type": e.ContentType,
		"content":      e.Content,
		"summary":      e.Summary,
		"author":       e.Author,
		"page_url":     e.PageURL,
		"edit_url":     e.EditURL,
		"published":    formatDate(e.Published),
		"updated":      formatDate(e.Updated),
	}
}

func entriesToJSON(entries []*hatena.Entry) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(entries))
	for _, e := range entries {
		out = append(out, entryToJSON(e))
	}
	return out
}

func draftLabel(draft bool) string {
	if draft {
		return "下書き"
	}
	return "公開"
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format("2006-01-02 15:04")
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}
