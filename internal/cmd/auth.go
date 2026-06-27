package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ideamans/hatena-blog-cli/internal/config"
	"github.com/ideamans/hatena-blog-cli/internal/hatena"
	"github.com/ideamans/hatena-blog-cli/internal/output"
	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "認証情報の管理",
		Long:  "はてなブログの認証情報（はてなID・ブログID・APIキー）の設定・確認・削除を行います。",
	}
	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthStatusCmd())
	cmd.AddCommand(newAuthLogoutCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var hatenaID, blogID, apiKey string
	var noVerify bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "認証情報を設定ファイルに保存します",
		Long: `はてなID・ブログID・APIキーを ~/.config/hatena-blog に保存します。

フラグで指定されなかった項目は対話的に入力を求めます。
APIキーははてなブログの「詳細設定」ページで確認できます。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(os.Stdin)

			// 既存の設定を初期値として読み込む
			existing, _ := config.Load()

			if hatenaID == "" {
				hatenaID = prompt(reader, "はてなID", existing.HatenaID)
			}
			if blogID == "" {
				blogID = prompt(reader, "ブログID (例: example.hatenablog.jp)", existing.BlogID)
			}
			if apiKey == "" {
				apiKey = prompt(reader, "APIキー", existing.APIKey)
			}

			cfg := &config.Config{
				HatenaID: strings.TrimSpace(hatenaID),
				BlogID:   strings.TrimSpace(blogID),
				APIKey:   strings.TrimSpace(apiKey),
			}
			if err := cfg.Validate(); err != nil {
				return err
			}

			if !noVerify {
				client := hatena.NewClient(cfg.HatenaID, cfg.BlogID, cfg.APIKey)
				if err := client.Verify(); err != nil {
					return fmt.Errorf("認証の確認に失敗しました（--no-verify で確認をスキップできます）: %w", err)
				}
			}

			if err := cfg.Save(); err != nil {
				return err
			}

			result := map[string]interface{}{
				"status":      "saved",
				"config_path": config.Path(),
				"hatena_id":   cfg.HatenaID,
				"blog_id":     cfg.BlogID,
				"verified":    !noVerify,
			}
			headers := []string{"項目", "値"}
			rows := [][]string{
				{"保存先", config.Path()},
				{"はてなID", cfg.HatenaID},
				{"ブログID", cfg.BlogID},
				{"認証確認", boolLabel(!noVerify, "成功", "スキップ")},
			}
			return output.Print(outputFormat, result, headers, rows)
		},
	}

	cmd.Flags().StringVar(&hatenaID, "hatena-id", "", "はてなID（ユーザー名）")
	cmd.Flags().StringVar(&blogID, "blog-id", "", "ブログID（例: example.hatenablog.jp）")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "APIキー")
	cmd.Flags().BoolVar(&noVerify, "no-verify", false, "保存前のAPI疎通確認をスキップする")

	return cmd
}

func newAuthStatusCmd() *cobra.Command {
	var verify bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "現在の認証情報の状態を表示します",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			verifyResult := "未確認"
			if verify {
				if err := cfg.Validate(); err != nil {
					return err
				}
				client := hatena.NewClient(cfg.HatenaID, cfg.BlogID, cfg.APIKey)
				if err := client.Verify(); err != nil {
					verifyResult = "失敗: " + err.Error()
				} else {
					verifyResult = "成功"
				}
			}

			configured := cfg.Validate() == nil
			result := map[string]interface{}{
				"config_path": config.Path(),
				"hatena_id":   cfg.HatenaID,
				"blog_id":     cfg.BlogID,
				"api_key_set": cfg.APIKey != "",
				"configured":  configured,
			}
			if verify {
				result["verify"] = verifyResult
			}

			headers := []string{"項目", "値"}
			rows := [][]string{
				{"設定ファイル", config.Path()},
				{"はてなID", orNone(cfg.HatenaID)},
				{"ブログID", orNone(cfg.BlogID)},
				{"APIキー", boolLabel(cfg.APIKey != "", "設定済み", "未設定")},
				{"設定状態", boolLabel(configured, "完了", "不足あり")},
			}
			if verify {
				rows = append(rows, []string{"疎通確認", verifyResult})
			}
			return output.Print(outputFormat, result, headers, rows)
		},
	}
	cmd.Flags().BoolVar(&verify, "verify", false, "APIへの疎通確認も行う")
	return cmd
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "保存された認証情報（設定ファイル）を削除します",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := config.Path()
			if err := os.Remove(path); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("設定ファイルが存在しません: %s", path)
				}
				return fmt.Errorf("設定ファイルの削除に失敗しました: %w", err)
			}
			result := map[string]interface{}{"status": "removed", "config_path": path}
			return output.Print(outputFormat, result, []string{"項目", "値"}, [][]string{
				{"状態", "削除しました"},
				{"対象", path},
			})
		},
	}
}

// prompt は標準入力から1行読み取ります。defaultVal があればEnterでそのままにできるようUIに表示します。
func prompt(reader *bufio.Reader, label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Fprintf(os.Stderr, "%s [%s]: ", label, defaultVal)
	} else {
		fmt.Fprintf(os.Stderr, "%s: ", label)
	}
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal
	}
	return line
}

func boolLabel(b bool, t, f string) string {
	if b {
		return t
	}
	return f
}

func orNone(s string) string {
	if s == "" {
		return "(未設定)"
	}
	return s
}
