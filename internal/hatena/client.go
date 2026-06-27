package hatena

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// acceptHeader ははてなブログAtomPub APIが返すコンテンツを受け取るためのAcceptヘッダーです。
const acceptHeader = "application/x.atom+xml, application/atom+xml, application/atomsvc+xml, application/xml, text/xml, */*"

// cinnamonWait はページネーション時のリクエスト間隔です（はてなサーバーへの配慮）。
const cinnamonWait = 250 * time.Millisecond

// Client ははてなブログAtomPub APIのクライアントです。
type Client struct {
	hatenaID string
	blogID   string
	apiKey   string
	http     *http.Client
}

// NewClient は新しいクライアントを生成します。
func NewClient(hatenaID, blogID, apiKey string) *Client {
	return &Client{
		hatenaID: hatenaID,
		blogID:   blogID,
		apiKey:   apiKey,
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

// collectionURI は記事コレクションのエンドポイントを返します。
func (c *Client) collectionURI() string {
	return fmt.Sprintf("https://blog.hatena.ne.jp/%s/%s/atom/entry", c.hatenaID, c.blogID)
}

// rootURI はサービス文書（ルート）のエンドポイントを返します。
func (c *Client) rootURI() string {
	return fmt.Sprintf("https://blog.hatena.ne.jp/%s/%s/atom", c.hatenaID, c.blogID)
}

// do はWSSE認証ヘッダーを付与してHTTPリクエストを実行し、レスポンスボディを返します。
func (c *Client) do(method, url string, body []byte) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("リクエストの生成に失敗しました: %w", err)
	}

	token, err := wsseToken(c.hatenaID, c.apiKey, time.Now())
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-WSSE", token)
	req.Header.Set("Accept", acceptHeader)
	if body != nil {
		req.Header.Set("Content-Type", "application/atom+xml; charset=utf-8")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("リクエストの送信に失敗しました: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンスの読み込みに失敗しました: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("APIエラー (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return respBody, nil
}

// Verify は認証情報が有効か、サービス文書を取得して確認します。
func (c *Client) Verify() error {
	_, err := c.do(http.MethodGet, c.rootURI(), nil)
	return err
}

// List は記事一覧を取得します。limit 件取得するまでページを辿ります。
// limit が 0 以下の場合は全件取得します。
func (c *Client) List(limit int) ([]*Entry, error) {
	var entries []*Entry
	url := c.collectionURI()

	for url != "" {
		body, err := c.do(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		var feed xmlFeed
		if err := xml.Unmarshal(body, &feed); err != nil {
			return nil, fmt.Errorf("フィードの解析に失敗しました: %w", err)
		}

		for _, xe := range feed.Entries {
			entries = append(entries, xe.toEntry())
			if limit > 0 && len(entries) >= limit {
				return entries[:limit], nil
			}
		}

		url = feed.nextLink()
		if url != "" {
			time.Sleep(cinnamonWait)
		}
	}

	return entries, nil
}

// Get は指定したMember URI（編集URL）の記事を取得します。
func (c *Client) Get(memberURL string) (*Entry, error) {
	body, err := c.do(http.MethodGet, memberURL, nil)
	if err != nil {
		return nil, err
	}
	var xe xmlEntry
	if err := xml.Unmarshal(body, &xe); err != nil {
		return nil, fmt.Errorf("記事の解析に失敗しました: %w", err)
	}
	return xe.toEntry(), nil
}

// Create は新しい記事を投稿し、作成された記事（編集URL等を含む）を返します。
func (c *Client) Create(e *Entry) (*Entry, error) {
	body, err := marshalEntry(e, c.hatenaID)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(http.MethodPost, c.collectionURI(), body)
	if err != nil {
		return nil, err
	}
	var xe xmlEntry
	if err := xml.Unmarshal(resp, &xe); err != nil {
		return nil, fmt.Errorf("レスポンスの解析に失敗しました: %w", err)
	}
	return xe.toEntry(), nil
}

// Update は既存の記事を更新します。e.EditURL に編集URLが設定されている必要があります。
func (c *Client) Update(e *Entry) (*Entry, error) {
	if e.EditURL == "" {
		return nil, fmt.Errorf("更新には編集URL (EditURL) が必要です")
	}
	body, err := marshalEntry(e, c.hatenaID)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(http.MethodPut, e.EditURL, body)
	if err != nil {
		return nil, err
	}
	var xe xmlEntry
	if err := xml.Unmarshal(resp, &xe); err != nil {
		return nil, fmt.Errorf("レスポンスの解析に失敗しました: %w", err)
	}
	return xe.toEntry(), nil
}

// Delete は指定した編集URLの記事を削除します。
func (c *Client) Delete(memberURL string) error {
	_, err := c.do(http.MethodDelete, memberURL, nil)
	return err
}

// marshalEntry はEntryをAtomPubのXMLリクエストボディへ変換します。
// encoding/xml の名前空間制御が煩雑なため、テンプレートで直接組み立てます。
func marshalEntry(e *Entry, defaultAuthor string) ([]byte, error) {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="utf-8"?>` + "\n")
	b.WriteString(`<entry xmlns="` + nsAtom + `" xmlns:app="` + nsApp + `">` + "\n")

	b.WriteString("  <title>" + escapeXML(e.Title) + "</title>\n")

	author := e.Author
	if author == "" {
		author = defaultAuthor
	}
	b.WriteString("  <author><name>" + escapeXML(author) + "</name></author>\n")

	contentType := e.ContentType
	if contentType == "" {
		contentType = ContentTypePlain
	}
	b.WriteString(`  <content type="` + escapeXMLAttr(contentType) + `">` + escapeXML(e.Content) + "</content>\n")

	if e.Summary != "" {
		b.WriteString(`  <summary type="text">` + escapeXML(e.Summary) + "</summary>\n")
	}

	if !e.Updated.IsZero() {
		b.WriteString("  <updated>" + e.Updated.Format(time.RFC3339) + "</updated>\n")
	}

	for _, cat := range e.Categories {
		b.WriteString(`  <category term="` + escapeXMLAttr(cat) + `" />` + "\n")
	}

	draft := "no"
	if e.Draft {
		draft = "yes"
	}
	b.WriteString("  <app:control><app:draft>" + draft + "</app:draft></app:control>\n")

	b.WriteString("</entry>\n")
	return []byte(b.String()), nil
}

// escapeXML はXML要素値用に文字をエスケープします。
func escapeXML(s string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

// escapeXMLAttr は属性値用のエスケープです。ダブルクォートも確実に処理します。
func escapeXMLAttr(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
	)
	return r.Replace(s)
}
