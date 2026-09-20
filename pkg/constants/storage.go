package constants

// 对象存储（MinIO）相关常量。
const (
	// StorageBucket 是项目统一使用的桶名，启动时若不存在会自动创建。
	StorageBucket = "gobili"

	// 对象 key 前缀，用于区分资源类型。
	StoragePrefixAvatar = "avatar"
	StoragePrefixVideo  = "video"
	StoragePrefixCover  = "cover"
)
