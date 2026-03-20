namespace go api.relation

include "../common.thrift"

// ==================== 社交模块请求/响应 ====================

// 关注操作
struct RelationActionReq {
    1: i64 user_id (api.json="user_id")
    2: i64 to_user_id (api.json="to_user_id")
    3: i8 action_type (api.json="action_type") // 1=关注, 2=取消
}

struct RelationActionResp {
    1: common.BaseResp base
}

// 关注列表
struct FollowingListReq {
    1: i64 user_id (api.query="user_id")
    2: i32 page_num (api.query="page_num")
    3: i32 page_size (api.query="page_size")
}

struct FollowingListResp {
    1: common.BaseResp base
    2: list<common.User> users
    3: i64 total
}

// 粉丝列表
struct FollowerListReq {
    1: i64 user_id (api.query="user_id")
    2: i32 page_num (api.query="page_num")
    3: i32 page_size (api.query="page_size")
}

struct FollowerListResp {
    1: common.BaseResp base
    2: list<common.User> users
    3: i64 total
}

// 好友列表（互相关注）
struct FriendsListReq {
    1: i64 user_id (api.query="user_id")
    2: i32 page_num (api.query="page_num")
    3: i32 page_size (api.query="page_size")
}

struct FriendsListResp {
    1: common.BaseResp base
    2: list<common.User> users
    3: i64 total
}

// ==================== 社交服务定义 ====================

service RelationService {
    RelationActionResp RelationAction(1: RelationActionReq req) (api.post="/relation/action")
    FollowingListResp GetFollowingList(1: FollowingListReq req) (api.get="/following/list")
    FollowerListResp GetFollowerList(1: FollowerListReq req) (api.get="/follower/list")
    FriendsListResp GetFriendsList(1: FriendsListReq req) (api.get="/friends/list")
}
