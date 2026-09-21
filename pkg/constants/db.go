package constants

// 表名。
//
// DAL 一律用 .Table(constants.XxxTableName) 显式指定表名，不依赖 GORM 的表名推断，
// 这样改结构体名字不会静默改掉查询目标。
const (
	UserTableName         = "users"
	RefreshTokenTableName = "refresh_tokens"
	VideoTableName        = "videos"
	CommentTableName      = "comments"
	ChatMessageTableName  = "chat_messages"
)
