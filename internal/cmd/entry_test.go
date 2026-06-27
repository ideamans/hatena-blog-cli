package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ideamans/hatena-blog-cli/internal/config"
	"github.com/ideamans/hatena-blog-cli/internal/hatena"
)

// mockAPI は hatena.API のテスト用実装です。呼び出しを記録します。
type mockAPI struct {
	verifyErr error
	listFn    func(int) ([]*hatena.Entry, error)
	getFn     func(string) (*hatena.Entry, error)

	lastListLimit int
	getURLs       []string
	created       []*hatena.Entry
	updated       []*hatena.Entry
	deleted       []string
}

func (m *mockAPI) Verify() error { return m.verifyErr }

func (m *mockAPI) List(limit int) ([]*hatena.Entry, error) {
	m.lastListLimit = limit
	if m.listFn != nil {
		return m.listFn(limit)
	}
	return nil, nil
}

func (m *mockAPI) Get(url string) (*hatena.Entry, error) {
	m.getURLs = append(m.getURLs, url)
	if m.getFn != nil {
		return m.getFn(url)
	}
	return &hatena.Entry{Title: "既存", EditURL: url, ContentType: hatena.ContentTypeMarkdown}, nil
}

func (m *mockAPI) Create(e *hatena.Entry) (*hatena.Entry, error) {
	m.created = append(m.created, e)
	out := *e
	if out.EditURL == "" {
		out.EditURL = "https://blog/edit/new/"
	}
	return &out, nil
}

func (m *mockAPI) Update(e *hatena.Entry) (*hatena.Entry, error) {
	m.updated = append(m.updated, e)
	out := *e
	return &out, nil
}

func (m *mockAPI) Delete(url string) error {
	m.deleted = append(m.deleted, url)
	return nil
}

// runCmd は newClient をモックに差し替えてルートコマンドを実行し、標準出力を捕捉します。
func runCmd(t *testing.T, m *mockAPI, args ...string) (string, error) {
	t.Helper()

	origNew := newClient
	newClient = func() (hatena.API, *config.Config, error) {
		return m, &config.Config{HatenaID: "u", BlogID: "b", APIKey: "k"}, nil
	}
	defer func() { newClient = origNew }()

	// 標準出力を捕捉
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = origStdout }()

	root := newRootCmd()
	root.SetArgs(args)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	err := root.Execute()

	w.Close()
	out, _ := io.ReadAll(r)
	return string(out), err
}

func writeFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "article.md")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestCreateFromFrontMatter(t *testing.T) {
	path := writeFile(t, `---
title: frontmatterタイトル
draft: true
categories: [テスト, CLI]
content_type: markdown
---
# 本文

これは本文です。
`)
	m := &mockAPI{}
	if _, err := runCmd(t, m, "entry", "create", "--file", path, "--format", "json"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.created) != 1 {
		t.Fatalf("expected 1 create, got %d", len(m.created))
	}
	e := m.created[0]
	if e.Title != "frontmatterタイトル" {
		t.Errorf("title from frontmatter: %q", e.Title)
	}
	if !e.Draft {
		t.Error("draft should be true from frontmatter")
	}
	if len(e.Categories) != 2 || e.Categories[0] != "テスト" {
		t.Errorf("categories: %v", e.Categories)
	}
	if e.ContentType != hatena.ContentTypeMarkdown {
		t.Errorf("content type: %q", e.ContentType)
	}
	if !strings.Contains(e.Content, "これは本文です。") {
		t.Errorf("body: %q", e.Content)
	}
}

func TestCreateFlagOverridesFrontMatter(t *testing.T) {
	path := writeFile(t, `---
title: 元タイトル
draft: false
---
本文
`)
	m := &mockAPI{}
	// --title と --draft でfrontmatterを上書き
	if _, err := runCmd(t, m, "entry", "create", "--file", path, "--title", "上書き", "--draft"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e := m.created[0]
	if e.Title != "上書き" {
		t.Errorf("flag should override title: %q", e.Title)
	}
	if !e.Draft {
		t.Error("flag --draft should override frontmatter draft:false")
	}
}

func TestCreateMissingTitle(t *testing.T) {
	path := writeFile(t, "本文のみ。frontmatterもtitleもない。\n")
	m := &mockAPI{}
	_, err := runCmd(t, m, "entry", "create", "--file", path)
	if err == nil {
		t.Fatal("missing title should error")
	}
	if len(m.created) != 0 {
		t.Error("should not call Create when title is missing")
	}
}

func TestUpdateResolvesEditURLFromFrontMatter(t *testing.T) {
	editURL := "https://blog/edit/77/"
	path := writeFile(t, `---
title: 更新タイトル
edit_url: `+editURL+`
draft: true
---
更新後の本文
`)
	m := &mockAPI{
		getFn: func(url string) (*hatena.Entry, error) {
			return &hatena.Entry{Title: "旧", EditURL: url, ContentType: hatena.ContentTypeMarkdown}, nil
		},
	}
	// 位置引数なし。edit_url は frontmatter から解決されるべき
	if _, err := runCmd(t, m, "entry", "update", "--file", path, "--format", "json"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.getURLs) != 1 || m.getURLs[0] != editURL {
		t.Errorf("Get should be called with frontmatter edit_url, got %v", m.getURLs)
	}
	if len(m.updated) != 1 {
		t.Fatalf("expected 1 update, got %d", len(m.updated))
	}
	u := m.updated[0]
	if u.Title != "更新タイトル" {
		t.Errorf("title from frontmatter: %q", u.Title)
	}
	if !strings.Contains(u.Content, "更新後の本文") {
		t.Errorf("body should be updated: %q", u.Content)
	}
	if !u.Draft {
		t.Error("draft from frontmatter should apply")
	}
}

func TestUpdateMissingEditURL(t *testing.T) {
	path := writeFile(t, "---\ntitle: x\n---\n本文\n")
	m := &mockAPI{}
	_, err := runCmd(t, m, "entry", "update", "--file", path)
	if err == nil {
		t.Fatal("update without edit URL should error")
	}
}

func TestUpdateFlagPrecedence(t *testing.T) {
	m := &mockAPI{}
	url := "https://blog/edit/5/"
	if _, err := runCmd(t, m, "entry", "update", url, "--title", "フラグ題", "--published"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.updated) != 1 {
		t.Fatalf("expected 1 update")
	}
	if m.updated[0].Title != "フラグ題" {
		t.Errorf("title: %q", m.updated[0].Title)
	}
	if m.updated[0].Draft {
		t.Error("--published should set draft=false")
	}
}

func TestListLimitPassthrough(t *testing.T) {
	m := &mockAPI{
		listFn: func(limit int) ([]*hatena.Entry, error) {
			return []*hatena.Entry{{Title: "a"}}, nil
		},
	}
	if _, err := runCmd(t, m, "entry", "list", "--limit", "7"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.lastListLimit != 7 {
		t.Errorf("limit should pass through as 7, got %d", m.lastListLimit)
	}
}

func TestDeleteForce(t *testing.T) {
	m := &mockAPI{}
	url := "https://blog/edit/9/"
	if _, err := runCmd(t, m, "entry", "delete", url, "--force"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.deleted) != 1 || m.deleted[0] != url {
		t.Errorf("delete should record url, got %v", m.deleted)
	}
}

func TestPullWritesManuscript(t *testing.T) {
	m := &mockAPI{
		getFn: func(url string) (*hatena.Entry, error) {
			return &hatena.Entry{
				Title:       "取得記事",
				EditURL:     url,
				Content:     "# 見出し\n本文\n",
				ContentType: hatena.ContentTypeMarkdown,
				Categories:  []string{"X"},
				Draft:       true,
			}, nil
		},
	}
	out := filepath.Join(t.TempDir(), "pulled.md")
	if _, err := runCmd(t, m, "entry", "pull", "https://blog/edit/1/", "-o", out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("output file not written: %v", err)
	}
	s := string(data)
	if !strings.HasPrefix(s, "---\n") {
		t.Errorf("manuscript should start with frontmatter: %q", s)
	}
	if !strings.Contains(s, "title: 取得記事") {
		t.Errorf("frontmatter should contain title: %q", s)
	}
	if !strings.Contains(s, "edit_url: https://blog/edit/1/") {
		t.Errorf("frontmatter should contain edit_url: %q", s)
	}
	if !strings.Contains(s, "content_type: markdown") {
		t.Errorf("content_type should be friendly name: %q", s)
	}
	if !strings.Contains(s, "本文") {
		t.Errorf("body should be present: %q", s)
	}
}

func TestContentTypeNameRoundTrip(t *testing.T) {
	cases := map[string]string{
		"markdown": hatena.ContentTypeMarkdown,
		"hatena":   hatena.ContentTypeHatena,
		"html":     hatena.ContentTypeHTML,
		"plain":    hatena.ContentTypePlain,
	}
	for name, mime := range cases {
		got, err := contentTypeFromName(name)
		if err != nil || got != mime {
			t.Errorf("contentTypeFromName(%q) = %q, %v", name, got, err)
		}
		if back := contentTypeToName(mime); back != name {
			t.Errorf("contentTypeToName(%q) = %q, want %q", mime, back, name)
		}
	}
	if _, err := contentTypeFromName("bogus"); err == nil {
		t.Error("unknown content type should error")
	}
}
