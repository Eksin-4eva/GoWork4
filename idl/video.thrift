// 视频服务的 RPC 契约。
namespace go video

include "rpc_model.thrift"

struct GetFeedRequest {
    // from 为 0 表示不限制起始时间；否则为毫秒时间戳
    1: optional i64 from_millis
}

struct GetFeedResponse {
    1: required rpc_model.BaseResp base
    2: optional list<rpc_model.Video> items
}

struct PublishVideoRequest {
    1: required i64 user_id
    2: required string title
    3: required string description
    // 视频与封面的 object key 由网关上传后传入
    4: required string video_key
    // cover_key 为空表示由 worker 异步抽帧生成
    5: optional string cover_key
}

struct PublishVideoResponse {
    1: required rpc_model.BaseResp base
}

struct GetVideoListRequest {
    1: required i64 user_id
    2: optional i64 page_num
    3: optional i64 page_size
}

struct GetVideoListResponse {
    1: required rpc_model.BaseResp base
    2: optional list<rpc_model.Video> items
    3: optional i64 total
}

struct GetPopularVideosRequest {
    1: optional i64 page_num
    2: optional i64 page_size
}

struct GetPopularVideosResponse {
    1: required rpc_model.BaseResp base
    2: optional list<rpc_model.Video> items
    3: optional i64 total
}

struct SearchVideosRequest {
    1: optional string keywords
    2: optional string username
    // 毫秒时间戳，0 表示不限制
    3: optional i64 from_millis
    4: optional i64 to_millis
    5: optional i64 page_num
    6: optional i64 page_size
}

struct SearchVideosResponse {
    1: required rpc_model.BaseResp base
    2: optional list<rpc_model.Video> items
    3: optional i64 total
}

struct VisitVideoRequest {
    1: required i64 video_id
}

struct VisitVideoResponse {
    1: required rpc_model.BaseResp base
}

service VideoService {
    // 获取最新视频流
    GetFeedResponse GetFeed(1: GetFeedRequest req)
    // 投稿视频
    PublishVideoResponse PublishVideo(1: PublishVideoRequest req)
    // 获取指定用户的投稿列表
    GetVideoListResponse GetVideoList(1: GetVideoListRequest req)
    // 热门排行榜
    GetPopularVideosResponse GetPopularVideos(1: GetPopularVideosRequest req)
    // 多条件搜索
    SearchVideosResponse SearchVideos(1: SearchVideosRequest req)
    // 记录播放
    VisitVideoResponse VisitVideo(1: VisitVideoRequest req)
}
