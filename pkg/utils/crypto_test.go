package utils

import (
	"strings"
	"testing"
	"time"
)

func TestPasswordHashRoundTrip(t *testing.T) {
	hash, err := HashPassword("s3cret-password")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "s3cret-password" {
		t.Fatal("哈希不能等于明文")
	}

	if !CheckPassword(hash, "s3cret-password") {
		t.Fatal("正确密码应校验通过")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Fatal("错误密码不该通过")
	}
}

// bcrypt 每次加盐，同一个密码两次哈希结果必须不同。
func TestPasswordHashIsSalted(t *testing.T) {
	first, err := HashPassword("same-password")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	second, err := HashPassword("same-password")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if first == second {
		t.Fatal("两次哈希结果相同，说明没有加盐")
	}
}

func TestVerifyMfaCode(t *testing.T) {
	secret, err := GenerateMfaSecret()
	if err != nil {
		t.Fatalf("GenerateMfaSecret: %v", err)
	}

	// 密钥是 20 字节，base32 编码后是 32 个字符，且不带 "=" 填充
	if len(secret) != 32 {
		t.Fatalf("密钥长度 = %d, 期望 32", len(secret))
	}
	if strings.Contains(secret, "=") {
		t.Fatal("密钥不应带 base32 填充")
	}

	key, err := mfaBase32().DecodeString(secret)
	if err != nil {
		t.Fatalf("解码密钥失败: %v", err)
	}

	now := time.Now().Unix() / mfaPeriodSec
	// 当前窗口与前后各一个窗口都应通过（时钟漂移容忍）
	for _, counter := range []int64{now - 1, now, now + 1} {
		code := totpCode(key, counter)
		if !VerifyMfaCode(secret, code) {
			t.Errorf("窗口 %d 的验证码 %s 应通过", counter, code)
		}
	}

	// 超出容忍范围必须拒绝
	if VerifyMfaCode(secret, totpCode(key, now-5)) {
		t.Error("过期窗口的验证码不该通过")
	}
	if VerifyMfaCode(secret, "000000") {
		t.Error("随机验证码不该通过")
	}
	if VerifyMfaCode("", "123456") || VerifyMfaCode(secret, "") {
		t.Error("空密钥或空验证码必须直接拒绝")
	}
	if VerifyMfaCode("not-base32!!", "123456") {
		t.Error("非法密钥必须直接拒绝")
	}
}

func TestMfaQRCodeURI(t *testing.T) {
	uri := MfaQRCodeURI("42", "ABCDEFGH")

	if !strings.HasPrefix(uri, "otpauth://totp/") {
		t.Fatalf("URI 前缀不对: %s", uri)
	}
	if !strings.Contains(uri, "secret=ABCDEFGH") || !strings.Contains(uri, "issuer=GoBili") {
		t.Fatalf("URI 缺少必要参数: %s", uri)
	}
}

func TestQRCodeDataURI(t *testing.T) {
	uri, err := QRCodeDataURI("otpauth://totp/GoBili:42?secret=ABCDEFGH")
	if err != nil {
		t.Fatalf("QRCodeDataURI: %v", err)
	}
	if !strings.HasPrefix(uri, "data:image/png;base64,") {
		t.Fatalf("data URI 前缀不对: %.40s", uri)
	}
	if len(uri) < 200 {
		t.Fatalf("生成的图片太小，可能没有内容: %d 字节", len(uri))
	}
}
