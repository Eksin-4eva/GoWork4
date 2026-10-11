#!/usr/bin/env bash
#
# 端到端冒烟测试：按真实调用顺序把 15 个接口跑一遍。
#
# 前置条件：
#   docker compose -f docker/docker-compose.yml up -d
#   go run ./cmd/api  -f config/config.yaml     （HTTP 网关，8888）
#   go run ./cmd/chat -f config/config.yaml     （WebSocket 聊天，8889）
#
# 用法：make smoke    或    BASE=http://host:port bash hack/smoke.sh
#
# 注意：故意不用 bash 4 的语法（macOS 自带的是 3.2）。
set -uo pipefail

BASE="${BASE:-http://127.0.0.1:8888}"
CHAT_BASE="${CHAT_BASE:-http://127.0.0.1:8889}"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

passed=0
failed=0

# check <用例名> <期望片段> <实际响应>
check() {
  if [[ "$3" == *"$2"* ]]; then
    printf '  \033[32m✓\033[0m %s\n' "$1"
    passed=$((passed + 1))
  else
    printf '  \033[31m✗\033[0m %s\n      期望包含: %s\n      实际返回: %s\n' "$1" "$2" "$3"
    failed=$((failed + 1))
  fi
}

# ok <用例名> <条件表达式的结果>
ok() {
  if [[ -n "$2" ]]; then
    printf '  \033[32m✓\033[0m %s\n' "$1"
    passed=$((passed + 1))
  else
    printf '  \033[31m✗\033[0m %s\n' "$1"
    failed=$((failed + 1))
  fi
}

# json <python 表达式>，从标准输入解析响应体
json() {
  python3 -c "import json,sys; d=json.load(sys.stdin); print($1)" 2>/dev/null || true
}

# header_value <响应头文本> <头名>
header_value() {
  awk -v want="$(echo "$2" | tr 'A-Z' 'a-z'):" 'tolower($1)==want {print $2}' <<<"$1" | tr -d '\r'
}

# 造一个能被嗅探出真实类型的假文件
make_fixture() {
  case "$1" in
    *.mp4) printf '\x00\x00\x00\x18ftypmp42\x00\x00\x00\x00mp42isom' > "$1" ;;
    *.jpg) printf '\xff\xd8\xff\xe0\x00\x10JFIF\x00\x01' > "$1" ;;
  esac
  head -c 4096 /dev/zero >> "$1"
}

# totp <base32 密钥>，按 RFC 6238 算当前验证码
totp() {
  python3 - "$1" <<'PY'
import base64, hashlib, hmac, struct, sys, time
secret = sys.argv[1]
key = base64.b32decode(secret + "=" * ((8 - len(secret) % 8) % 8))
digest = hmac.new(key, struct.pack(">Q", int(time.time()) // 30), hashlib.sha1).digest()
offset = digest[-1] & 0x0F
print("%06d" % ((struct.unpack(">I", digest[offset:offset + 4])[0] & 0x7FFFFFFF) % 1000000))
PY
}

make_fixture "$WORK/video.mp4"
make_fixture "$WORK/cover.jpg"
USERNAME="smoke$(date +%s)"
PASSWORD="password123"

echo "== 用户模块 =="

check "注册" '"code":"0"' \
  "$(curl -s -X POST "$BASE/user/register" -d "username=$USERNAME&password=$PASSWORD")"
check "重复注册被拒" '已被占用' \
  "$(curl -s -X POST "$BASE/user/register" -d "username=$USERNAME&password=$PASSWORD")"
check "弱密码被拒" '密码长度' \
  "$(curl -s -X POST "$BASE/user/register" -d "username=${USERNAME}x&password=123")"

headers=$(curl -s -D - -o /dev/null -X POST "$BASE/user/login" -d "username=$USERNAME&password=$PASSWORD")
ACCESS=$(header_value "$headers" access-token)
REFRESH=$(header_value "$headers" refresh-token)
ok "登录下发 Access-Token" "$ACCESS"
ok "登录下发 Refresh-Token" "$REFRESH"

check "密码错误被拒" '用户名或密码错误' \
  "$(curl -s -X POST "$BASE/user/login" -d "username=$USERNAME&password=wrong-password")"
check "未带令牌被拒" '"code":"30001"' "$(curl -s "$BASE/user/info")"
check "令牌非法被拒" '"code":"30001"' "$(curl -s "$BASE/user/info" -H "Authorization: not-a-token")"

info=$(curl -s "$BASE/user/info" -H "Authorization: $ACCESS")
check "查自己信息" "\"username\":\"$USERNAME\"" "$info"
USER_ID=$(json "d['data']['id']" <<<"$info")

avatar=$(curl -s -X PUT "$BASE/user/avatar/upload" -H "Authorization: $ACCESS" -F "file=@$WORK/cover.jpg")
check "上传头像" '/avatar/' "$avatar"
check "头像 URL 可匿名访问" "200" \
  "$(curl -s -o /dev/null -w '%{http_code}' "$(json "d['data']['avatar_url']" <<<"$avatar")")"
check "上传非图片被拒" '不支持的文件类型' \
  "$(curl -s -X PUT "$BASE/user/avatar/upload" -H "Authorization: $ACCESS" -F "file=@$WORK/video.mp4")"

rotated=$(header_value "$(curl -s -D - -o /dev/null "$BASE/user/info" -H "Authorization: $ACCESS")" access-token)
if [[ -n "$rotated" && "$rotated" != "$ACCESS" ]]; then
  ok "每次请求轮换令牌" "rotated"
else
  ok "每次请求轮换令牌" ""
fi

echo
echo "== 视频模块 =="

check "投稿（含封面）" '"code":"0"' \
  "$(curl -s -X POST "$BASE/video/publish" -H "Authorization: $ACCESS" \
      -F "file=@$WORK/video.mp4" -F "cover=@$WORK/cover.jpg" \
      -F "title=冒烟测试视频" -F "description=E2E")"
check "投稿缺文件被拒" '缺少文件字段' \
  "$(curl -s -X POST "$BASE/video/publish" -H "Authorization: $ACCESS" -F "title=没有文件")"
check "投稿空标题被拒" '标题不能为空' \
  "$(curl -s -X POST "$BASE/video/publish" -H "Authorization: $ACCESS" -F "file=@$WORK/video.mp4")"

feed=$(curl -s "$BASE/video/feed")
check "feed 含新投稿" '冒烟测试视频' "$feed"
VIDEO_ID=$(json "d['data']['items'][0]['id']" <<<"$feed")

# 返回的 URL 必须真的能取到文件。旧版就是在这里翻车的：
# 地址拼得出来，但资源根本不存在（旧版 404，本项目此前是桶私有导致的 403）。
check "视频 URL 可匿名访问" "200" \
  "$(curl -s -o /dev/null -w '%{http_code}' "$(json "d['data']['items'][0]['video_url']" <<<"$feed")")"
check "封面 URL 可匿名访问" "200" \
  "$(curl -s -o /dev/null -w '%{http_code}' "$(json "d['data']['items'][0]['cover_url']" <<<"$feed")")"

check "投稿列表" '冒烟测试视频' \
  "$(curl -s "$BASE/video/list?user_id=$USER_ID&page_num=1&page_size=10")"
check "搜索命中" '冒烟测试视频' "$(curl -s -X POST "$BASE/video/search" -d "keywords=冒烟")"
check "搜索无命中为空" '"total":0' "$(curl -s -X POST "$BASE/video/search" -d "keywords=nonexistent-keyword")"

check "记录播放" '"code":"0"' \
  "$(curl -s -X POST "$BASE/video/visit" -H "Authorization: $ACCESS" -d "video_id=$VIDEO_ID")"
check "播放不存在的视频" '视频不存在' \
  "$(curl -s -X POST "$BASE/video/visit" -H "Authorization: $ACCESS" -d "video_id=999999999")"
check "热门榜" '冒烟测试视频' "$(curl -s "$BASE/video/popular?page_num=1&page_size=10")"

echo
echo "== 评论模块 =="

check "发评论" '"code":"0"' \
  "$(curl -s -X POST "$BASE/comment/publish" -H "Authorization: $ACCESS" -d "video_id=$VIDEO_ID&content=第一条评论")"
check "评论不存在的视频" '视频不存在' \
  "$(curl -s -X POST "$BASE/comment/publish" -H "Authorization: $ACCESS" -d "video_id=999999999&content=x")"

comments=$(curl -s "$BASE/comment/list?video_id=$VIDEO_ID&page_num=1&page_size=10")
check "评论列表" '第一条评论' "$comments"
COMMENT_ID=$(json "d['data']['items'][0]['id']" <<<"$comments")

check "回复评论" '"code":"0"' \
  "$(curl -s -X POST "$BASE/comment/publish" -H "Authorization: $ACCESS" \
      -d "video_id=$VIDEO_ID&comment_id=$COMMENT_ID&content=回复")"
check "视频评论数已累加" '"comment_count":2' "$(curl -s "$BASE/video/feed")"

check "删除评论" '"code":"0"' \
  "$(curl -s -X DELETE "$BASE/comment/delete" -H "Authorization: $ACCESS" -d "comment_id=$COMMENT_ID")"
check "重复删除已不存在" '评论不存在' \
  "$(curl -s -X DELETE "$BASE/comment/delete" -H "Authorization: $ACCESS" -d "comment_id=$COMMENT_ID")"

echo
echo "== MFA 模块 =="

mfa=$(curl -s "$BASE/auth/mfa/qrcode" -H "Authorization: $ACCESS")
SECRET=$(json "d['data']['secret']" <<<"$mfa")
check "获取 MFA 二维码" 'data:image/png;base64,' "$mfa"
check "绑定时验证码错误被拒" '验证码错误' \
  "$(curl -s -X POST "$BASE/auth/mfa/bind" -H "Authorization: $ACCESS" -d "secret=$SECRET&code=000000")"
check "绑定 MFA" '"code":"0"' \
  "$(curl -s -X POST "$BASE/auth/mfa/bind" -H "Authorization: $ACCESS" \
      -d "secret=$SECRET&code=$(totp "$SECRET")")"
check "登录缺少验证码被拒" '验证码错误' \
  "$(curl -s -X POST "$BASE/user/login" -d "username=$USERNAME&password=$PASSWORD")"
check "带验证码登录成功" '"code":"0"' \
  "$(curl -s -X POST "$BASE/user/login" -d "username=$USERNAME&password=$PASSWORD&code=$(totp "$SECRET")")"

echo
echo "== 聊天服务 =="

# WebSocket 的正常收发由 internal/chat/service 的集成测试覆盖，
# 这里只验证进程起来了、鉴权生效、历史接口可用。
check "健康检查" "ok" "$(curl -s "$CHAT_BASE/healthz")"
check "历史接口未鉴权被拒" "401" \
  "$(curl -s -o /dev/null -w '%{http_code}' "$CHAT_BASE/history?peer_id=1")"
check "历史接口鉴权后可访问" '"code":"0"' \
  "$(curl -s "$CHAT_BASE/history?peer_id=1" -H "Authorization: $ACCESS")"

echo
printf '\033[1m通过 %d 项，失败 %d 项\033[0m\n' "$passed" "$failed"
[[ "$failed" -eq 0 ]]
