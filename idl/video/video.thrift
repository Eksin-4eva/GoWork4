namespace go api.video

include "../common.thrift"

// ==================== 视频模块请求/响应 ====================

struct PublishVideoReq {
    1: i64 user_id (api.form="user_id")
    2: string title (api.form="title")
    3: string description (api.form="description")
    4: binary video_data (api.form="video_data")
    5: binary cover_data (api.form="cover_data")
    6: string video_file_name (api.form="video_file_name")
    7: string cover_file_name (api.form="cover_file_name")
}

struct PublishVideoResp {
    1: common.BaseResp base
    2: i64 video_id
    3: common.Video video
}

struct VideoListReq {
    1: i64 user_id (api.query="user_id")
    2: i32 page_num (api.query="page_num")
    3: i32 page_size (api.query="page_size")
}

struct VideoListResp {
    1: common.BaseResp base
    2: list<common.Video> items
    3: i64 total
}

struct PopularVideosReq {
    1: i32 page_num (api.query="page_num")
    2: i32 page_size (api.query="page_size")
}

struct PopularVideosResp {
    1: common.BaseResp base
    2: list<common.Video> items
    3: i64 total
}

struct SearchVideoReq {
    1: string keyword (api.json="keyword")
    2: i32 page_num (api.json="page_num")
    3: i32 page_size (api.json="page_size")
}

struct SearchVideoResp {
    1: common.BaseResp base
    2: list<common.Video> items
    3: i64 total
}

// ==================== 视频服务定义 ====================

service VideoService {
    PublishVideoResp PublishVideo(1: PublishVideoReq req) (api.post="/video/publish")
    VideoListResp GetVideoList(1: VideoListReq req) (api.get="/video/list")
    PopularVideosResp GetPopularVideos(1: PopularVideosReq req) (api.get="/video/popular")
    SearchVideoResp SearchVideo(1: SearchVideoReq req) (api.post="/video/search")
}
