package manuscript

import (
	"strings"
	"testing"
)

func TestParseWithFrontMatter(t *testing.T) {
	src := `---
title: テスト記事
draft: true
categories:
  - テスト
  - Markdown
content_type: markdown
summary: 概要です
edit_url: https://example/atom/entry/1/
---
# 本文

これは本文です。
`
	ms, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ms.HasFront {
		t.Fatal("HasFront should be true")
	}
	if ms.Front.Title != "テスト記事" {
		t.Errorf("title: %q", ms.Front.Title)
	}
	if ms.Front.Draft == nil || !*ms.Front.Draft {
		t.Errorf("draft should be true, got %v", ms.Front.Draft)
	}
	if len(ms.Front.Categories) != 2 || ms.Front.Categories[0] != "テスト" {
		t.Errorf("categories: %v", ms.Front.Categories)
	}
	if ms.Front.ContentType != "markdown" {
		t.Errorf("content_type: %q", ms.Front.ContentType)
	}
	if ms.Front.EditURL != "https://example/atom/entry/1/" {
		t.Errorf("edit_url: %q", ms.Front.EditURL)
	}
	wantBody := "# 本文\n\nこれは本文です。\n"
	if ms.Body != wantBody {
		t.Errorf("body mismatch:\n got %q\nwant %q", ms.Body, wantBody)
	}
}

func TestParseNoFrontMatter(t *testing.T) {
	src := "# ただの本文\n\n---区切りではない\n"
	ms, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ms.HasFront {
		t.Error("HasFront should be false")
	}
	if ms.Body != src {
		t.Errorf("body should equal whole input, got %q", ms.Body)
	}
	if ms.Front.Title != "" {
		t.Errorf("title should be empty, got %q", ms.Front.Title)
	}
}

func TestParseDraftFalseDistinguishable(t *testing.T) {
	withFalse, _ := Parse([]byte("---\ndraft: false\n---\nbody\n"))
	if withFalse.Front.Draft == nil {
		t.Fatal("draft:false should be non-nil pointer")
	}
	if *withFalse.Front.Draft {
		t.Error("draft should be false")
	}

	without, _ := Parse([]byte("---\ntitle: x\n---\nbody\n"))
	if without.Front.Draft != nil {
		t.Error("unspecified draft should be nil")
	}
}

func TestParseUnterminatedFrontMatter(t *testing.T) {
	_, err := Parse([]byte("---\ntitle: x\nbody without closing\n"))
	if err == nil {
		t.Error("expected error for unterminated frontmatter")
	}
}

func TestRenderRoundTrip(t *testing.T) {
	draft := true
	fm := FrontMatter{
		Title:       "往復テスト",
		Draft:       &draft,
		Categories:  []string{"A", "B"},
		ContentType: "markdown",
		EditURL:     "https://example/atom/entry/9/",
	}
	body := "# 見出し\n\n本文。\n"

	doc, err := Render(fm, body)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	s := string(doc)
	if !strings.HasPrefix(s, "---\n") {
		t.Errorf("should start with frontmatter delimiter: %q", s)
	}

	// 再パースして往復一致を確認
	ms, err := Parse(doc)
	if err != nil {
		t.Fatalf("reparse error: %v", err)
	}
	if ms.Front.Title != fm.Title {
		t.Errorf("title roundtrip: %q", ms.Front.Title)
	}
	if ms.Front.Draft == nil || *ms.Front.Draft != true {
		t.Errorf("draft roundtrip: %v", ms.Front.Draft)
	}
	if len(ms.Front.Categories) != 2 {
		t.Errorf("categories roundtrip: %v", ms.Front.Categories)
	}
	if ms.Front.EditURL != fm.EditURL {
		t.Errorf("edit_url roundtrip: %q", ms.Front.EditURL)
	}
	if ms.Body != body {
		t.Errorf("body roundtrip:\n got %q\nwant %q", ms.Body, body)
	}
}

func TestParseCRLF(t *testing.T) {
	src := "---\r\ntitle: CRLF\r\n---\r\nbody line\r\n"
	ms, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ms.Front.Title != "CRLF" {
		t.Errorf("title: %q", ms.Front.Title)
	}
}
