package hatena

import (
	"encoding/xml"
	"strings"
	"time"
)

// コンテンツタイプの定数。content要素のtype属性に指定します。
const (
	ContentTypePlain    = "text/plain"
	ContentTypeMarkdown = "text/x-markdown"
	ContentTypeHatena   = "text/x-hatena-syntax"
	ContentTypeHTML     = "text/html"
)

// XML名前空間
const (
	nsAtom = "http://www.w3.org/2005/Atom"
	nsApp  = "http://www.w3.org/2007/app"
)

// Entry ははてなブログの記事を表します。投稿・更新の入力と、取得結果の両方に使われます。
type Entry struct {
	// ID は記事の一意な識別子（tag: URI）です。取得時のみ設定されます。
	ID string
	// EditURL は記事を編集（更新・削除）するためのAtomPub Member URIです。
	EditURL string
	// PageURL は記事の公開ページURLです。
	PageURL string
	// Title は記事タイトルです。
	Title string
	// Content は記事本文です。
	Content string
	// ContentType は本文の形式です（ContentType* 定数を参照）。
	ContentType string
	// Categories は記事のカテゴリ（タグ）一覧です。
	Categories []string
	// Draft が true の場合、記事は下書きとして扱われます。
	Draft bool
	// Summary は概要です。
	Summary string
	// Published は公開日時です（取得時のみ）。
	Published time.Time
	// Updated は更新日時です。投稿・更新時に指定すると日時を上書きできます。
	Updated time.Time
	// Author は著者名です（取得時のみ）。
	Author string
}

// --- レスポンス（受信）用のXML構造体 ---

type xmlFeed struct {
	XMLName xml.Name   `xml:"feed"`
	Links   []xmlLink  `xml:"link"`
	Entries []xmlEntry `xml:"entry"`
}

type xmlEntry struct {
	XMLName    xml.Name      `xml:"entry"`
	ID         string        `xml:"id"`
	Links      []xmlLink     `xml:"link"`
	Title      string        `xml:"title"`
	Summary    string        `xml:"summary"`
	Content    xmlContent    `xml:"content"`
	Published  string        `xml:"published"`
	Updated    string        `xml:"updated"`
	Author     xmlAuthor     `xml:"author"`
	Categories []xmlCategory `xml:"category"`
	Control    xmlControl    `xml:"control"`
}

type xmlLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
	Type string `xml:"type,attr"`
}

type xmlContent struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type xmlAuthor struct {
	Name string `xml:"name"`
}

type xmlCategory struct {
	Term string `xml:"term,attr"`
}

type xmlControl struct {
	Draft string `xml:"draft"`
}

// toEntry はXMLエントリをドメインモデルのEntryへ変換します。
func (x xmlEntry) toEntry() *Entry {
	e := &Entry{
		ID:          x.ID,
		Title:       x.Title,
		Summary:     x.Summary,
		Content:     x.Content.Value,
		ContentType: x.Content.Type,
		Author:      x.Author.Name,
		Draft:       strings.TrimSpace(x.Control.Draft) == "yes",
	}
	for _, c := range x.Categories {
		e.Categories = append(e.Categories, c.Term)
	}
	for _, l := range x.Links {
		switch l.Rel {
		case "edit":
			e.EditURL = l.Href
		case "alternate":
			e.PageURL = l.Href
		}
	}
	if t, err := time.Parse(time.RFC3339, x.Published); err == nil {
		e.Published = t
	}
	if t, err := time.Parse(time.RFC3339, x.Updated); err == nil {
		e.Updated = t
	}
	return e
}

// nextLink はfeedの rel="next" リンク（次ページURL）を返します。なければ空文字列。
func (f xmlFeed) nextLink() string {
	for _, l := range f.Links {
		if l.Rel == "next" {
			return l.Href
		}
	}
	return ""
}
