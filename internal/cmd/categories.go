package cmd

import (
	"sort"
	"strconv"

	"github.com/ideamans/hatena-blog-cli/internal/output"
	"github.com/spf13/cobra"
)

func newCategoriesCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "categories",
		Aliases: []string{"category"},
		Short:   "ブログで使われているカテゴリの一覧を表示します",
		Long:    "全記事を走査し、使用されているカテゴリと各カテゴリの記事数を集計して表示します。",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := newClient()
			if err != nil {
				return err
			}
			entries, err := client.List(0)
			if err != nil {
				return err
			}

			counts := map[string]int{}
			for _, e := range entries {
				for _, c := range e.Categories {
					counts[c]++
				}
			}

			names := make([]string, 0, len(counts))
			for name := range counts {
				names = append(names, name)
			}
			sort.Slice(names, func(i, j int) bool {
				if counts[names[i]] != counts[names[j]] {
					return counts[names[i]] > counts[names[j]]
				}
				return names[i] < names[j]
			})

			type catCount struct {
				Category string `json:"category"`
				Count    int    `json:"count"`
			}
			result := make([]catCount, 0, len(names))
			headers := []string{"カテゴリ", "記事数"}
			rows := make([][]string, 0, len(names))
			for _, name := range names {
				result = append(result, catCount{Category: name, Count: counts[name]})
				rows = append(rows, []string{name, strconv.Itoa(counts[name])})
			}
			return output.Print(outputFormat, result, headers, rows)
		},
	}
}
