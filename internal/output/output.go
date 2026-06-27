// Package output はコマンド結果の出力フォーマット（json / table）を扱います。
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-runewidth"
)

// Format は出力フォーマットを表します。
const (
	FormatJSON  = "json"
	FormatTable = "table"
)

// Print は data を指定フォーマットで標準出力へ書き出します。
// table フォーマットの場合は headers と rows を使用し、json の場合は data をそのまま整形します。
func Print(format string, data interface{}, headers []string, rows [][]string) error {
	switch strings.ToLower(format) {
	case FormatJSON, "":
		return printJSON(data)
	case FormatTable:
		printTable(headers, rows)
		return nil
	default:
		return fmt.Errorf("未対応の出力フォーマットです: %s (json, table が使用可能)", format)
	}
}

func printJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(data)
}

func printTable(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = runewidth.StringWidth(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				if w := runewidth.StringWidth(cell); w > widths[i] {
					widths[i] = w
				}
			}
		}
	}

	printRow(headers, widths)
	seps := make([]string, len(headers))
	for i, w := range widths {
		seps[i] = strings.Repeat("─", w)
	}
	fmt.Println(strings.Join(seps, "──"))
	for _, row := range rows {
		printRow(row, widths)
	}
}

func printRow(cells []string, widths []int) {
	padded := make([]string, len(widths))
	for i := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		pad := widths[i] - runewidth.StringWidth(cell)
		if pad < 0 {
			pad = 0
		}
		padded[i] = cell + strings.Repeat(" ", pad)
	}
	fmt.Println(strings.TrimRight(strings.Join(padded, "  "), " "))
}
