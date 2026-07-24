// Package cmd ははてなブログCLIのコマンド定義を提供します。
package cmd

import (
	"github.com/ideamans/go-llm-cli-kit/llmcmd"
	"github.com/spf13/cobra"

	"github.com/ideamans/hatena-blog-cli/internal/config"
	"github.com/ideamans/hatena-blog-cli/internal/hatena"
)

// グローバルなフラグ
var outputFormat string

// PluginVersion はこのCLIのリリースバージョンです。
// plugins/hatena-blog-cli/.claude-plugin/plugin.json の version と一致していること
// をテストが、git タグと一致していることをリリースワークフローが検査します。
// タグを打つコミットで一緒に上げてください。
const PluginVersion = "0.2.0"

// バージョン情報（main から SetVersion で設定される）
var versionString = PluginVersion

// SetVersion はビルド時のバージョン情報を設定します。
func SetVersion(version, commit, date string) {
	versionString = version
	if commit != "" {
		versionString += " (" + commit
		if date != "" {
			versionString += ", " + date
		}
		versionString += ")"
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "hatena-blog",
		Version: versionString,
		Short:   "はてなブログを操作するCLI",
		Long: `はてなブログを操作するコマンドラインツールです。

AtomPub APIを通じて記事の投稿・取得・更新・削除を行います。

認証情報（はてなID・ブログID・APIキー）は以下の優先順位で解決されます。
  1. 環境変数 (HATENA_BLOG_HATENA_ID, HATENA_BLOG_ID, HATENA_BLOG_API_KEY)
  2. 設定ファイル (~/.config/hatena-blog)

初期設定は 'hatena-blog auth login' で対話的に行えます。
AIエージェント向けの詳細リファレンスは 'hatena-blog llm' で表示できます。`,
		SilenceUsage:  true,
		SilenceErrors: true,
		// 引数なし（サブコマンドなし）でもヘルプを出す。
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().StringVar(&outputFormat, "format", "table", "出力フォーマット: table, json")

	cmd.AddCommand(newAuthCmd())
	cmd.AddCommand(newEntryCmd())
	cmd.AddCommand(newCategoriesCmd())
	llmcmd.AddTo(cmd, LLMConfig())

	return cmd
}

// Root は組み立て済みのコマンドツリーを実行せずに返します。
// カタログ生成器が main と同じ定義から生成するために使います。
func Root() *cobra.Command { return newRootCmd() }

// Execute はルートコマンドを実行します。
func Execute() error {
	return newRootCmd().Execute()
}

// buildClient は設定からAPIクライアント（インターフェース）を生成するファクトリです。
// パッケージ変数にすることで、テスト時にモックへ差し替えられます。
var buildClient = func(cfg *config.Config) hatena.API {
	return hatena.NewClient(cfg.HatenaID, cfg.BlogID, cfg.APIKey)
}

// newClient は設定を読み込み・検証してからAPIクライアントを生成する共通ヘルパーです。
// こちらもパッケージ変数（関数値）にしており、テストでクライアントと設定をまとめて
// 差し替えられます。
var newClient = func() (hatena.API, *config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}
	return buildClient(cfg), cfg, nil
}
