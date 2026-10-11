package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"gobili/config"
	"gobili/pkg/constants"
	"gobili/pkg/errno"
)

const testConfigPath = "../../config/config.example.yaml"

func loadTestConfig(t *testing.T) {
	t.Helper()
	if err := config.Init(testConfigPath); err != nil {
		t.Fatalf("加载测试配置失败: %v", err)
	}
}

func TestCreateAndCheckToken(t *testing.T) {
	loadTestConfig(t)

	access, refresh, err := CreateAllToken(12345)
	if err != nil {
		t.Fatalf("CreateAllToken: %v", err)
	}
	if access == refresh {
		t.Fatal("访问令牌与刷新令牌不应相同")
	}

	got, err := CheckToken(access, constants.TypeAccessToken)
	if err != nil {
		t.Fatalf("校验访问令牌失败: %v", err)
	}
	if got != 12345 {
		t.Fatalf("user_id = %d, 期望 12345", got)
	}

	if _, err := CheckToken(refresh, constants.TypeRefreshToken); err != nil {
		t.Fatalf("校验刷新令牌失败: %v", err)
	}
}

// 同一秒内签发的两个令牌必须不同。
//
// 没有 jti 时载荷完全相同（JWT 时间戳只到秒），签名也相同，
// 刷新令牌会直接撞上 refresh_tokens.uniq_token 唯一键。
// 这个问题是跑端到端冒烟测试时暴露出来的。
func TestCreateTokenIsUniqueWithinSameSecond(t *testing.T) {
	loadTestConfig(t)

	first, _, err := CreateAllToken(1)
	if err != nil {
		t.Fatalf("CreateAllToken: %v", err)
	}
	second, _, err := CreateAllToken(1)
	if err != nil {
		t.Fatalf("CreateAllToken: %v", err)
	}
	if first == second {
		t.Fatal("同一秒内签发的两个令牌相同，会导致刷新令牌撞唯一键")
	}
}

// 拿刷新令牌去访问业务接口必须被拒——这是双令牌方案的核心约束。
func TestCheckTokenRejectsWrongType(t *testing.T) {
	loadTestConfig(t)

	_, refresh, err := CreateAllToken(1)
	if err != nil {
		t.Fatalf("CreateAllToken: %v", err)
	}

	if _, err := CheckToken(refresh, constants.TypeAccessToken); err == nil {
		t.Fatal("刷新令牌不该通过访问令牌的校验")
	}
}

func TestCheckTokenRejectsTampered(t *testing.T) {
	loadTestConfig(t)

	access, _, err := CreateAllToken(1)
	if err != nil {
		t.Fatalf("CreateAllToken: %v", err)
	}

	tampered := access[:len(access)-1] + "x"
	if _, err := CheckToken(tampered, constants.TypeAccessToken); err == nil {
		t.Fatal("被篡改的令牌必须校验失败")
	}
	if _, err := CheckToken("not-a-jwt", constants.TypeAccessToken); err == nil {
		t.Fatal("非法格式必须校验失败")
	}
}

// 令牌必须带签发者，才能挡掉「同一密钥的其他系统签的令牌」。
func TestCheckTokenRejectsWrongIssuer(t *testing.T) {
	loadTestConfig(t)

	cfg := config.Get().JWT
	claims := Claims{
		UserID: 1,
		Type:   constants.TypeAccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "some-other-service",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.AccessSecret))
	if err != nil {
		t.Fatalf("构造令牌失败: %v", err)
	}

	if _, err := CheckToken(raw, constants.TypeAccessToken); err == nil {
		t.Fatal("签发者不匹配的令牌必须被拒")
	}
}

func TestCheckTokenRejectsExpired(t *testing.T) {
	loadTestConfig(t)

	cfg := config.Get().JWT
	claims := Claims{
		UserID: 1,
		Type:   constants.TypeAccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    constants.JWTIssuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.AccessSecret))
	if err != nil {
		t.Fatalf("构造令牌失败: %v", err)
	}

	if _, err := CheckToken(raw, constants.TypeAccessToken); err == nil {
		t.Fatal("过期令牌必须被拒")
	}
}

// 校验失败的原因不对外区分，统一落到鉴权错误码。
func TestCheckTokenHidesReason(t *testing.T) {
	loadTestConfig(t)

	_, err := CheckToken("garbage", constants.TypeAccessToken)
	if errno.ConvertErr(err).ErrorCode != errno.AuthErrorCode {
		t.Fatalf("应返回鉴权错误码，实际 %v", err)
	}
}
