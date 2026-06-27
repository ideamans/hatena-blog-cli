package hatena

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	testHatenaID = "testuser"
	testBlogID   = "test.hatenablog.com"
)

// newTestClient はテストサーバーを指す Client を生成します。
func newTestClient(serverURL string) *Client {
	return NewClient(testHatenaID, testBlogID, "testkey",
		WithBaseURL(serverURL),
		WithPageWait(0),
	)
}

func entryXML(title, editHref string, draft bool) string {
	d := "no"
	if draft {
		d = "yes"
	}
	return `<?xml version="1.0" encoding="utf-8"?>
<entry xmlns="http://www.w3.org/2005/Atom" xmlns:app="http://www.w3.org/2007/app">
  <id>tag:blog.hatena.ne.jp,2013:entry-1</id>
  <link rel="edit" href="` + editHref + `"/>
  <link rel="alternate" type="text/html" href="https://test.hatenablog.com/entry/1"/>
  <title>` + title + `</title>
  <updated>2026-06-27T10:00:00+09:00</updated>
  <author><name>testuser</name></author>
  <category term="A"/>
  <content type="text/x-markdown">本文</content>
  <app:control><app:draft>` + d + `</app:draft></app:control>
</entry>`
}

func assertWSSE(t *testing.T, r *http.Request) {
	t.Helper()
	if tok := r.Header.Get("X-WSSE"); !strings.HasPrefix(tok, "UsernameToken ") {
		t.Errorf("X-WSSE ヘッダーが不正: %q", tok)
	}
}

func TestClientVerify(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertWSSE(t, r)
		if r.URL.Path != "/testuser/test.hatenablog.com/atom" {
			t.Errorf("予期しないパス: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `<service xmlns="http://www.w3.org/2007/app"></service>`)
	}))
	defer srv.Close()

	if err := newTestClient(srv.URL).Verify(); err != nil {
		t.Fatalf("Verify should succeed: %v", err)
	}
}

func TestClientVerifyUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `Invalid login`)
	}))
	defer srv.Close()

	err := newTestClient(srv.URL).Verify()
	if err == nil {
		t.Fatal("401 should produce an error")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error should mention 401: %v", err)
	}
}

func TestClientListPagination(t *testing.T) {
	var serverURL string
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertWSSE(t, r)
		hits++
		w.WriteHeader(http.StatusOK)
		if r.URL.Query().Get("page") == "" {
			// 1ページ目: nextリンクあり
			io.WriteString(w, `<feed xmlns="http://www.w3.org/2005/Atom">
  <link rel="next" href="`+serverURL+`/testuser/test.hatenablog.com/atom/entry?page=2"/>
  `+strings.ReplaceAll(entryXML("記事1", "https://e/1/", false), `<?xml version="1.0" encoding="utf-8"?>`, "")+`
</feed>`)
		} else {
			// 2ページ目: nextリンクなし
			io.WriteString(w, `<feed xmlns="http://www.w3.org/2005/Atom">
  `+strings.ReplaceAll(entryXML("記事2", "https://e/2/", true), `<?xml version="1.0" encoding="utf-8"?>`, "")+`
</feed>`)
		}
	}))
	defer srv.Close()
	serverURL = srv.URL

	entries, err := newTestClient(srv.URL).List(0)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if hits != 2 {
		t.Errorf("should follow pagination (2 requests), got %d", hits)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Title != "記事1" || entries[1].Title != "記事2" {
		t.Errorf("titles: %q, %q", entries[0].Title, entries[1].Title)
	}
	if !entries[1].Draft {
		t.Error("entry2 should be draft")
	}
}

func TestClientListLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// nextを常に返すが、limitで打ち切られるべき
		io.WriteString(w, `<feed xmlns="http://www.w3.org/2005/Atom">
  <link rel="next" href="`+r.Host+`?page=99"/>
  `+strings.ReplaceAll(entryXML("a", "https://e/a/", false), `<?xml version="1.0" encoding="utf-8"?>`, "")+`
  `+strings.ReplaceAll(entryXML("b", "https://e/b/", false), `<?xml version="1.0" encoding="utf-8"?>`, "")+`
</feed>`)
	}))
	defer srv.Close()

	entries, err := newTestClient(srv.URL).List(1)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("limit=1 should return 1 entry, got %d", len(entries))
	}
}

func TestClientCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertWSSE(t, r)
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/testuser/test.hatenablog.com/atom/entry" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		s := string(body)
		if !strings.Contains(s, "<title>新規記事</title>") {
			t.Errorf("request body missing title: %s", s)
		}
		if !strings.Contains(s, "<app:draft>yes</app:draft>") {
			t.Errorf("request body should mark draft: %s", s)
		}
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, entryXML("新規記事", "https://blog/edit/123/", true))
	}))
	defer srv.Close()

	got, err := newTestClient(srv.URL).Create(&Entry{Title: "新規記事", Content: "本文", ContentType: ContentTypeMarkdown, Draft: true})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if got.EditURL != "https://blog/edit/123/" {
		t.Errorf("EditURL: %q", got.EditURL)
	}
	if !got.Draft {
		t.Error("returned entry should be draft")
	}
}

func TestClientUpdate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertWSSE(t, r)
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/testuser/test.hatenablog.com/atom/entry/55" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, entryXML("更新後", "x", false))
	}))
	defer srv.Close()

	editURL := srv.URL + "/testuser/test.hatenablog.com/atom/entry/55"
	got, err := newTestClient(srv.URL).Update(&Entry{Title: "更新後", EditURL: editURL})
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if got.Title != "更新後" {
		t.Errorf("title: %q", got.Title)
	}
}

func TestClientUpdateRequiresEditURL(t *testing.T) {
	_, err := NewClient("u", "b", "k").Update(&Entry{Title: "x"})
	if err == nil {
		t.Error("Update without EditURL should error")
	}
}

func TestClientDelete(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertWSSE(t, r)
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	if err := newTestClient(srv.URL).Delete(srv.URL + "/x/y/atom/entry/9"); err != nil {
		t.Fatalf("Delete error: %v", err)
	}
	if !called {
		t.Error("delete handler not called")
	}
}

func TestClientGetNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "404 Entry Not Found")
	}))
	defer srv.Close()

	_, err := newTestClient(srv.URL).Get(srv.URL + "/x")
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 error, got %v", err)
	}
}
