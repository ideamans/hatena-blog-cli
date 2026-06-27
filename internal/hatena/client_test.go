package hatena

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"
)

func TestMarshalEntry(t *testing.T) {
	e := &Entry{
		Title:       "テスト記事 <special> & \"quote\"",
		Content:     "本文です\n2行目 <tag>",
		ContentType: ContentTypeMarkdown,
		Categories:  []string{"技術", "Go & 言語"},
		Draft:       true,
		Summary:     "概要",
	}
	body, err := marshalEntry(e, "defaultauthor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(body)

	// XML宣言とルート要素
	if !strings.HasPrefix(s, `<?xml version="1.0" encoding="utf-8"?>`) {
		t.Errorf("missing xml declaration: %q", s)
	}
	if !strings.Contains(s, `xmlns:app="http://www.w3.org/2007/app"`) {
		t.Error("missing app namespace")
	}

	// 下書きフラグ
	if !strings.Contains(s, "<app:draft>yes</app:draft>") {
		t.Error("draft should be yes")
	}

	// カテゴリ
	if !strings.Contains(s, `<category term="技術" />`) {
		t.Error("missing category 技術")
	}
	if !strings.Contains(s, `term="Go &amp; 言語"`) {
		t.Errorf("category with & not escaped properly: %q", s)
	}

	// デフォルト著者
	if !strings.Contains(s, "<name>defaultauthor</name>") {
		t.Error("missing default author")
	}

	// content-typeとエスケープ
	if !strings.Contains(s, `<content type="text/x-markdown">`) {
		t.Error("missing markdown content type")
	}

	// 生成したXMLがパース可能であること（エスケープが正しい証拠）
	var parsed xmlEntry
	if err := xml.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("generated XML is not parseable: %v\n%s", err, s)
	}
	if parsed.Title != e.Title {
		t.Errorf("title roundtrip failed: got %q want %q", parsed.Title, e.Title)
	}
	if parsed.Content.Value != e.Content {
		t.Errorf("content roundtrip failed: got %q want %q", parsed.Content.Value, e.Content)
	}
}

func TestMarshalEntryDefaults(t *testing.T) {
	e := &Entry{Title: "t", Content: "c"}
	body, _ := marshalEntry(e, "author")
	s := string(body)
	if !strings.Contains(s, "<app:draft>no</app:draft>") {
		t.Error("default draft should be no")
	}
	if !strings.Contains(s, `<content type="text/plain">`) {
		t.Error("default content type should be text/plain")
	}
	if strings.Contains(s, "<summary") {
		t.Error("empty summary should be omitted")
	}
	if strings.Contains(s, "<updated>") {
		t.Error("zero updated should be omitted")
	}
}

func TestParseFeed(t *testing.T) {
	feedXML := `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom" xmlns:app="http://www.w3.org/2007/app">
  <link rel="next" href="https://blog.hatena.ne.jp/x/y/atom/entry?page=2"/>
  <entry>
    <id>tag:blog.hatena.ne.jp,2026:entry-1</id>
    <link rel="edit" href="https://example/atom/entry/1/"/>
    <link rel="alternate" type="text/html" href="https://example.hatenablog.jp/entry/1"/>
    <title>記事タイトル</title>
    <published>2026-06-27T10:00:00+09:00</published>
    <updated>2026-06-27T11:00:00+09:00</updated>
    <author><name>someone</name></author>
    <category term="A"/>
    <category term="B"/>
    <content type="text/x-markdown">本文</content>
    <app:control><app:draft>yes</app:draft></app:control>
  </entry>
</feed>`

	var feed xmlFeed
	if err := xml.Unmarshal([]byte(feedXML), &feed); err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if got := feed.nextLink(); got != "https://blog.hatena.ne.jp/x/y/atom/entry?page=2" {
		t.Errorf("nextLink wrong: %q", got)
	}
	if len(feed.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(feed.Entries))
	}

	e := feed.Entries[0].toEntry()
	if e.Title != "記事タイトル" {
		t.Errorf("title: %q", e.Title)
	}
	if e.EditURL != "https://example/atom/entry/1/" {
		t.Errorf("edit url: %q", e.EditURL)
	}
	if e.PageURL != "https://example.hatenablog.jp/entry/1" {
		t.Errorf("page url: %q", e.PageURL)
	}
	if !e.Draft {
		t.Error("should be draft")
	}
	if len(e.Categories) != 2 || e.Categories[0] != "A" || e.Categories[1] != "B" {
		t.Errorf("categories: %v", e.Categories)
	}
	if e.Author != "someone" {
		t.Errorf("author: %q", e.Author)
	}
	want, _ := time.Parse(time.RFC3339, "2026-06-27T10:00:00+09:00")
	if !e.Published.Equal(want) {
		t.Errorf("published: %v want %v", e.Published, want)
	}
}
