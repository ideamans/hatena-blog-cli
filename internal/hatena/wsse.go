package hatena

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"time"
)

// wsseToken は X-WSSE ヘッダーに設定するUsernameTokenを生成します。
//
// PasswordDigest = Base64( SHA1( nonce + created + apiKey ) )
// の形式で、はてなブログのAtomPub APIが要求するWSSE認証に対応します。
//
// username にははてなID、password にはAPIキーを渡します。
func wsseToken(username, password string, created time.Time) (string, error) {
	nonce := make([]byte, 40)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("nonceの生成に失敗しました: %w", err)
	}

	createdStr := created.UTC().Format("2006-01-02T15:04:05Z")

	h := sha1.New()
	h.Write(nonce)
	h.Write([]byte(createdStr))
	h.Write([]byte(password))
	digest := base64.StdEncoding.EncodeToString(h.Sum(nil))

	nonceB64 := base64.StdEncoding.EncodeToString(nonce)

	return fmt.Sprintf(
		`UsernameToken Username="%s", PasswordDigest="%s", Nonce="%s", Created="%s"`,
		username, digest, nonceB64, createdStr,
	), nil
}
