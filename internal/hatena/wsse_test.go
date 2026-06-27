package hatena

import (
	"crypto/sha1"
	"encoding/base64"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestWsseToken(t *testing.T) {
	created := time.Date(2026, 6, 27, 1, 2, 3, 0, time.UTC)
	token, err := wsseToken("myhatenaid", "secretkey", created)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(token, "UsernameToken ") {
		t.Errorf("token should start with 'UsernameToken ': %q", token)
	}
	if !strings.Contains(token, `Username="myhatenaid"`) {
		t.Errorf("token missing username: %q", token)
	}
	if !strings.Contains(token, `Created="2026-06-27T01:02:03Z"`) {
		t.Errorf("token has wrong created format: %q", token)
	}

	// 各フィールドを抽出
	field := func(name string) string {
		re := regexp.MustCompile(name + `="([^"]*)"`)
		m := re.FindStringSubmatch(token)
		if m == nil {
			t.Fatalf("field %s not found in %q", name, token)
		}
		return m[1]
	}

	nonceB64 := field("Nonce")
	createdStr := field("Created")
	digestB64 := field("PasswordDigest")

	nonce, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		t.Fatalf("nonce is not valid base64: %v", err)
	}
	if len(nonce) != 40 {
		t.Errorf("nonce should be 40 bytes, got %d", len(nonce))
	}

	// PasswordDigest = Base64(SHA1(nonce + created + apiKey)) を独立に検証
	h := sha1.New()
	h.Write(nonce)
	h.Write([]byte(createdStr))
	h.Write([]byte("secretkey"))
	want := base64.StdEncoding.EncodeToString(h.Sum(nil))
	if digestB64 != want {
		t.Errorf("digest mismatch:\n got %q\nwant %q", digestB64, want)
	}
}

func TestWsseTokenNonceUnique(t *testing.T) {
	created := time.Date(2026, 6, 27, 1, 2, 3, 0, time.UTC)
	a, _ := wsseToken("u", "k", created)
	b, _ := wsseToken("u", "k", created)
	if a == b {
		t.Error("two tokens with same time should differ due to random nonce")
	}
}
