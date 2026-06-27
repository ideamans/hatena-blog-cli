// Package cmd ははてなブログCLIのコマンド定義を提供します。
package cmd

import (
	"errors"
	"fmt"

	"github.com/ideamans/hatena-blog-cli/internal/config"
	"github.com/ideamans/hatena-blog-cli/internal/hatena"
	"github.com/spf13/cobra"
)

// グローバルなフラグ
var (
	outputFormat string
	llmHelp      bool
)

// バージョン情報（main から SetVersion で設定される）
var versionString = "dev"

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
LLMエージェント向けの詳細ガイドは 'hatena-blog --llm' で表示できます。`,
		SilenceUsage:  true,
		SilenceErrors: true,
		// --llm は全コマンド共通で最優先に処理する。
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if llmHelp {
				fmt.Fprint(cmd.OutOrStdout(), llmGuide)
				// ガイドを表示したら以降の処理は行わない。
				return errLLMHelpShown
			}
			return nil
		},
		// 引数なし（サブコマンドなし）でも --llm を機能させるため RunE を持たせる。
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().StringVar(&outputFormat, "format", "table", "出力フォーマット: table, json")
	cmd.PersistentFlags().BoolVar(&llmHelp, "llm", false, "LLMエージェント向けの詳細な利用ガイドを表示する")

	cmd.AddCommand(newAuthCmd())
	cmd.AddCommand(newEntryCmd())
	cmd.AddCommand(newCategoriesCmd())

	return cmd
}

// errLLMHelpShown は --llm ガイド表示後に通常処理を打ち切るためのセンチネルです。
var errLLMHelpShown = errors.New("__llm_help_shown__")

// Execute はルートコマンドを実行します。
func Execute() error {
	err := newRootCmd().Execute()
	if errors.Is(err, errLLMHelpShown) {
		return nil
	}
	return err
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
