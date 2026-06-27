// Package config はてなブログCLIの認証情報・設定の読み書きを担当します。
//
// 認証情報は以下の優先順位で解決されます。
//  1. 環境変数 (HATENA_BLOG_API_KEY, HATENA_BLOG_HATENA_ID, HATENA_BLOG_ID)
//  2. 設定ファイル (~/.config/hatena-blog、JSON形式)
//
// 環境変数で指定された項目はファイルの値を上書きします。
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// 環境変数名
const (
	EnvAPIKey   = "HATENA_BLOG_API_KEY"   // APIキー
	EnvHatenaID = "HATENA_BLOG_HATENA_ID" // はてなID（ユーザー名）
	EnvBlogID   = "HATENA_BLOG_ID"        // ブログID（例: example.hatenablog.jp）
	EnvConfig   = "HATENA_BLOG_CONFIG"    // 設定ファイルパスの上書き
)

// Config ははてなブログAtomPub APIへ接続するための認証情報です。
type Config struct {
	// HatenaID ははてなID（ブログ所有者のユーザー名）です。WSSE認証のUsernameにも使われます。
	HatenaID string `json:"hatena_id"`
	// BlogID はブログの識別子です（例: example.hatenablog.jp）。
	BlogID string `json:"blog_id"`
	// APIKey ははてなブログの詳細設定ページで発行されるAPIキーです。
	APIKey string `json:"api_key"`
}

// Path は設定ファイルのパスを返します。
// HATENA_BLOG_CONFIG が設定されていればそれを、なければ ~/.config/hatena-blog を返します。
func Path() string {
	if p := os.Getenv(EnvConfig); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".config", "hatena-blog")
	}
	return filepath.Join(home, ".config", "hatena-blog")
}

// Load は設定ファイルを読み込み、環境変数で上書きしたConfigを返します。
// ファイルが存在しなくてもエラーにはせず、環境変数のみで構成を試みます。
func Load() (*Config, error) {
	cfg := &Config{}

	// まずファイルから読み込む（存在しなければ無視）
	if data, err := os.ReadFile(Path()); err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("設定ファイルの解析に失敗しました (%s): %w", Path(), err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗しました (%s): %w", Path(), err)
	}

	// 環境変数で上書き
	if v := os.Getenv(EnvHatenaID); v != "" {
		cfg.HatenaID = v
	}
	if v := os.Getenv(EnvBlogID); v != "" {
		cfg.BlogID = v
	}
	if v := os.Getenv(EnvAPIKey); v != "" {
		cfg.APIKey = v
	}

	return cfg, nil
}

// Validate は認証に必要な項目がすべて揃っているか確認します。
func (c *Config) Validate() error {
	var missing []string
	if c.HatenaID == "" {
		missing = append(missing, "はてなID ("+EnvHatenaID+")")
	}
	if c.BlogID == "" {
		missing = append(missing, "ブログID ("+EnvBlogID+")")
	}
	if c.APIKey == "" {
		missing = append(missing, "APIキー ("+EnvAPIKey+")")
	}
	if len(missing) > 0 {
		return fmt.Errorf("認証情報が不足しています: %v\n'hatena-blog auth login' で設定するか、環境変数を指定してください", missing)
	}
	return nil
}

// Save は設定をファイル (~/.config/hatena-blog) に保存します。
// APIキーを含むため、パーミッションは 0600 に設定します。
func (c *Config) Save() error {
	path := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("設定ディレクトリの作成に失敗しました: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("設定のシリアライズに失敗しました: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("設定ファイルの書き込みに失敗しました (%s): %w", path, err)
	}
	return nil
}
