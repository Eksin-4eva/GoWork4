namespace go api

struct User {
    1: i64 id
    2: string username
    3: string avatar_url
    4: i64 follower_count
    5: i64 following_count
    6: string created_at
    7: string updated_at
}

struct Video {
    1: i64 id
    2: i64 user_id
    3: string title
    4: string description
    5: string video_url
    6: string cover_url
    7: i64 visit_count
    8: i64 like_count
    9: i64 comment_count
    10: string created_at
    11: string updated_at
    12: User author
}

struct Comment {
    1: i64 id
    2: i64 video_id
    3: i64 user_id
    4: string content
    5: i64 parent_id
    6: i64 like_count
    7: i64 child_count
    8: string created_at
    9: string updated_at
    10: User author
}

struct Relation {
    1: i64 id
    2: i64 user_id
    3: i64 to_user_id
    4: i8 status
    5: string created_at
    6: string updated_at
}

struct BaseResp {
    1: i32 code
    2: string message
}
