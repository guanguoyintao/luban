package datacollection

import (
	"context"
	"time"
)

// 用户行为类型
type UserBehaviorType string

const (
	BehaviorClick      UserBehaviorType = "click"      // 点击行为
	BehaviorView       UserBehaviorType = "view"       // 浏览行为
	BehaviorPurchase   UserBehaviorType = "purchase"   // 购买行为
	BehaviorRating     UserBehaviorType = "rating"     // 评分行为
	BehaviorFavorite   UserBehaviorType = "favorite"   // 收藏行为
	BehaviorShare      UserBehaviorType = "share"      // 分享行为
)

// 用户行为数据
type UserBehavior struct {
	UserID     string           // 用户ID
	ItemID     string           // 物品ID
	Behavior   UserBehaviorType // 行为类型
	Value      float64          // 行为数值（如评分值）
	Timestamp  time.Time        // 行为发生时间
	Context    map[string]interface{} // 上下文信息
}

// 物品数据
type ItemData struct {
	ItemID      string                 // 物品ID
	Category    string                 // 物品类别
	Title       string                 // 物品标题
	Description string                 // 物品描述
	Features    map[string]interface{} // 物品特征
	Metadata    map[string]interface{} // 元数据
}

// 用户数据
type UserData struct {
	UserID      string                 // 用户ID
	Demographics map[string]interface{} // 人口统计学信息
	Preferences map[string]interface{} // 用户偏好
	Metadata    map[string]interface{} // 元数据
}

// 数据采集器接口
type DataCollector interface {
	// 收集用户行为数据
	CollectUserBehavior(ctx context.Context, behavior UserBehavior) error
	
	// 批量收集用户行为数据
	CollectUserBehaviors(ctx context.Context, behaviors []UserBehavior) error
	
	// 收集物品数据
	CollectItemData(ctx context.Context, item ItemData) error
	
	// 批量收集物品数据
	CollectItemsData(ctx context.Context, items []ItemData) error
	
	// 收集用户数据
	CollectUserData(ctx context.Context, user UserData) error
	
	// 批量收集用户数据
	CollectUsersData(ctx context.Context, users []UserData) error
	
	// 获取用户行为历史
	GetUserBehaviorHistory(ctx context.Context, userID string, limit int) ([]UserBehavior, error)
	
	// 获取物品数据
	GetItemData(ctx context.Context, itemID string) (*ItemData, error)
	
	// 获取用户数据
	GetUserData(ctx context.Context, userID string) (*UserData, error)
	
	// 关闭采集器
	Close() error
}