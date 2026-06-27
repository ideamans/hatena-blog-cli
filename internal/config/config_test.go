package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvOverridesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hatena-blog")
	t.Setenv(EnvConfig, path)

	// ファイルに保存
	saved := &Config{HatenaID: "fileuser", BlogID: "file.hatenablog.jp", APIKey: "filekey"}
	if err := saved.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// パーミッション確認（APIキーを含むので0600）
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("config file perm = %o, want 600", perm)
	}

	// ファイルのみ
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if cfg.HatenaID != "fileuser" || cfg.APIKey != "filekey" {
		t.Errorf("file load wrong: %+v", cfg)
	}

	// 環境変数で上書き
	t.Setenv(EnvAPIKey, "envkey")
	t.Setenv(EnvHatenaID, "envuser")
	cfg, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HatenaID != "envuser" {
		t.Errorf("env should override hatena_id: %q", cfg.HatenaID)
	}
	if cfg.APIKey != "envkey" {
		t.Errorf("env should override api_key: %q", cfg.APIKey)
	}
	if cfg.BlogID != "file.hatenablog.jp" {
		t.Errorf("blog_id should remain from file: %q", cfg.BlogID)
	}
}

func TestLoadMissingFileNoError(t *testing.T) {
	t.Setenv(EnvConfig, filepath.Join(t.TempDir(), "does-not-exist"))
	cfg, err := Load()
	if err != nil {
		t.Fatalf("missing file should not error: %v", err)
	}
	if err := cfg.Validate(); err == nil {
		t.Error("empty config should fail validation")
	}
}

func TestValidate(t *testing.T) {
	full := &Config{HatenaID: "a", BlogID: "b", APIKey: "c"}
	if err := full.Validate(); err != nil {
		t.Errorf("full config should validate: %v", err)
	}
	if err := (&Config{HatenaID: "a"}).Validate(); err == nil {
		t.Error("partial config should fail")
	}
}
