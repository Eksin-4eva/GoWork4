package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"
)

// MFA（TOTP，RFC 6238）相关参数。
const (
	mfaSecretBytes = 20 // 160 bit，与 Google Authenticator 建议一致
	mfaPeriodSec   = 30 // 一个验证码的有效秒数
	mfaDigits      = 6
	mfaIssuer      = "GoBili"
)

// HashPassword 用 bcrypt 生成密码哈希。
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("utils.HashPassword: %w", err)
	}
	return string(hash), nil
}

// CheckPassword 校验明文密码与哈希是否匹配。
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// GenerateMfaSecret 生成 TOTP 密钥。
//
// 用标准库的 base32 编码且不带填充：Google Authenticator 等客户端不接受 "=" 填充，
// 旧版手写了一套 base32 编解码，属于重复造轮子且容易出错。
func GenerateMfaSecret() (string, error) {
	buf := make([]byte, mfaSecretBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("utils.GenerateMfaSecret: %w", err)
	}
	return mfaBase32().EncodeToString(buf), nil
}

// MfaQRCodeURI 拼出 otpauth:// 链接，客户端扫码即可绑定。
func MfaQRCodeURI(account, secret string) string {
	params := url.Values{
		"secret": {secret},
		"issuer": {mfaIssuer},
	}
	label := url.PathEscape(mfaIssuer + ":" + account)
	return fmt.Sprintf("otpauth://totp/%s?%s", label, params.Encode())
}

// QRCodeDataURI 把文本渲染成 PNG 的 data URI，客户端可以直接放进 img 标签。
func QRCodeDataURI(content string) (string, error) {
	png, err := qrcode.Encode(content, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("utils.QRCodeDataURI: %w", err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}

// VerifyMfaCode 校验 TOTP 验证码。
//
// 容忍前后各一个时间窗：手机与服务器时钟很难完全对齐，
// 不放宽会让用户频繁遇到「验证码错误」。放宽两个窗最多多 30 秒的有效期，
// 对 6 位码来说风险可以接受。
func VerifyMfaCode(secret, code string) bool {
	secret = strings.TrimSpace(secret)
	code = strings.TrimSpace(code)
	if secret == "" || code == "" {
		return false
	}

	key, err := mfaBase32().DecodeString(strings.ToUpper(secret))
	if err != nil {
		return false
	}

	counter := time.Now().Unix() / mfaPeriodSec
	for offset := int64(-1); offset <= 1; offset++ {
		// 定长比较，避免用 == 带来的时间侧信道
		if subtle.ConstantTimeCompare([]byte(totpCode(key, counter+offset)), []byte(code)) == 1 {
			return true
		}
	}
	return false
}

// mfaBase32 返回不带填充的 base32 编码器。
func mfaBase32() *base32.Encoding {
	return base32.StdEncoding.WithPadding(base32.NoPadding)
}

// totpCode 按 RFC 6238 计算指定时间窗的验证码。
func totpCode(key []byte, counter int64) string {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	sum := mac.Sum(nil)

	// 动态截断：取最后一个字节的低 4 位作为偏移量
	offset := sum[len(sum)-1] & 0x0F
	value := (int(sum[offset]&0x7F) << 24) |
		(int(sum[offset+1]) << 16) |
		(int(sum[offset+2]) << 8) |
		int(sum[offset+3])

	return fmt.Sprintf("%0*d", mfaDigits, value%1_000_000)
}
