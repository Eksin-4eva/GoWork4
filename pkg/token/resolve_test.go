package token

import (
	"context"
	"testing"

	"gobili/pkg/errno"
)

// stubExchanger 记录被调用了几次，用来断言「访问令牌有效时不换发」。
//
// 这个假实现能写出来，本身就是接口存在的价值：不用起数据库、
// 也不用起 etcd 去连一个真的 user 服务。
type stubExchanger struct {
	calls    int
	gotUser  int64
	gotOld   string
	newToken string
	err      error
}

func (s *stubExchanger) ExchangeRefreshToken(
	_ context.Context, userID int64, oldRefresh string,
) (string, error) {
	s.calls++
	s.gotUser = userID
	s.gotOld = oldRefresh
	if s.err != nil {
		return "", s.err
	}
	return s.newToken, nil
}

// 访问令牌有效时走的是纯本地验签：不查库、不发 RPC，也不该换发刷新令牌。
//
// 这条断言守的是一个性能特征——绝大多数请求（访问令牌 2 小时有效期内）
// 都走这条路，一旦它开始调用 exchanger，网关就会为每个请求多一次 RPC。
func TestResolveUserUsesAccessTokenWithoutExchanging(t *testing.T) {
	loadTestConfig(t)

	access, _, err := CreateAllToken(4242)
	if err != nil {
		t.Fatalf("CreateAllToken: %v", err)
	}

	stub := &stubExchanger{newToken: "不该被用到"}
	userID, newRefresh, err := ResolveUser(context.Background(), stub, access, "")
	if err != nil {
		t.Fatalf("ResolveUser: %v", err)
	}
	if userID != 4242 {
		t.Fatalf("userID = %d，期望 4242", userID)
	}
	if stub.calls != 0 {
		t.Fatalf("访问令牌有效时不该换发，exchanger 被调用了 %d 次", stub.calls)
	}
	if newRefresh != "" {
		t.Fatalf("没走刷新令牌路径时不该下发新刷新令牌，实际 %q", newRefresh)
	}
}

// 只有刷新令牌时：验签通过后交给 exchanger 换发，并把新令牌返回给调用方。
func TestResolveUserExchangesRefreshToken(t *testing.T) {
	loadTestConfig(t)

	_, refresh, err := CreateAllToken(777)
	if err != nil {
		t.Fatalf("CreateAllToken: %v", err)
	}

	stub := &stubExchanger{newToken: "brand-new-refresh"}
	userID, newRefresh, err := ResolveUser(context.Background(), stub, "", refresh)
	if err != nil {
		t.Fatalf("ResolveUser: %v", err)
	}
	if userID != 777 {
		t.Fatalf("userID = %d，期望 777", userID)
	}
	if stub.calls != 1 {
		t.Fatalf("exchanger 应被调用 1 次，实际 %d 次", stub.calls)
	}
	if stub.gotUser != 777 {
		t.Fatalf("换发时传的 user_id = %d，期望 777", stub.gotUser)
	}
	if stub.gotOld != refresh {
		t.Fatal("换发时应该把旧令牌原样传下去")
	}
	if newRefresh != "brand-new-refresh" {
		t.Fatalf("newRefresh = %q，期望 exchanger 返回的那个", newRefresh)
	}
}

// 签名不对的刷新令牌在这里就该被挡住，不能白白消耗一次查询/RPC。
func TestResolveUserRejectsUnverifiedRefreshBeforeExchanging(t *testing.T) {
	loadTestConfig(t)

	stub := &stubExchanger{newToken: "brand-new-refresh"}
	_, _, err := ResolveUser(context.Background(), stub, "", "not-a-jwt")
	if err == nil {
		t.Fatal("非法刷新令牌必须被拒")
	}
	if stub.calls != 0 {
		t.Fatalf("验签失败时不该调用 exchanger，实际调用 %d 次", stub.calls)
	}
	if errno.ConvertErr(err).ErrorCode != errno.AuthErrorCode {
		t.Fatalf("应返回鉴权错误码，实际 %v", err)
	}
}

// 用刷新令牌冒充访问令牌时，验签通过但类型不对，同样要挡在换发之前。
func TestResolveUserRejectsWrongTokenType(t *testing.T) {
	loadTestConfig(t)

	access, _, err := CreateAllToken(1)
	if err != nil {
		t.Fatalf("CreateAllToken: %v", err)
	}

	stub := &stubExchanger{newToken: "brand-new-refresh"}
	if _, _, err := ResolveUser(context.Background(), stub, "", access); err == nil {
		t.Fatal("访问令牌不能用来换发")
	}
	if stub.calls != 0 {
		t.Fatalf("类型不对时不该调用 exchanger，实际调用 %d 次", stub.calls)
	}
}

// 换发失败（令牌已吊销/已过期）时，错误要原样冒出来，不能被吞掉。
func TestResolveUserPropagatesExchangeFailure(t *testing.T) {
	loadTestConfig(t)

	_, refresh, err := CreateAllToken(1)
	if err != nil {
		t.Fatalf("CreateAllToken: %v", err)
	}

	stub := &stubExchanger{err: errno.AuthInvalid}
	userID, newRefresh, err := ResolveUser(context.Background(), stub, "", refresh)
	if err == nil {
		t.Fatal("换发失败时必须返回错误")
	}
	if userID != 0 || newRefresh != "" {
		t.Fatalf("失败时不应返回用户与新令牌，实际 userID=%d newRefresh=%q", userID, newRefresh)
	}
	if errno.ConvertErr(err).ErrorCode != errno.AuthErrorCode {
		t.Fatalf("应返回鉴权错误码，实际 %v", err)
	}
}

// 两个令牌都没带：直接拒绝，不去打扰 exchanger。
func TestResolveUserRejectsEmpty(t *testing.T) {
	loadTestConfig(t)

	stub := &stubExchanger{}
	if _, _, err := ResolveUser(context.Background(), stub, "", ""); err == nil {
		t.Fatal("没带任何令牌必须被拒")
	}
	if stub.calls != 0 {
		t.Fatalf("不该调用 exchanger，实际调用 %d 次", stub.calls)
	}
}

// 访问令牌无效（过期/被篡改）而刷新令牌有效时，要回退到刷新令牌这条路。
func TestResolveUserFallsBackWhenAccessInvalid(t *testing.T) {
	loadTestConfig(t)

	_, refresh, err := CreateAllToken(999)
	if err != nil {
		t.Fatalf("CreateAllToken: %v", err)
	}

	stub := &stubExchanger{newToken: "rotated"}
	userID, newRefresh, err := ResolveUser(context.Background(), stub, "tampered.jwt.value", refresh)
	if err != nil {
		t.Fatalf("ResolveUser: %v", err)
	}
	if userID != 999 {
		t.Fatalf("userID = %d，期望 999", userID)
	}
	if newRefresh != "rotated" {
		t.Fatalf("newRefresh = %q，期望 rotated", newRefresh)
	}
}
