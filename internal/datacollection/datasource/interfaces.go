// Package datasource 数据源适配器接口定义
package datasource

import (
	"context"
	"time"
)

// DataSource 数据源接口
type DataSource interface {
	// 获取用户行为数据
	GetUserBehaviorData(ctx context.Context, userID string, startTime, endTime time.Time) ([]UserBehaviorRecord, error)

	// 获取物品数据
	GetItemData(ctx context.Context, itemIDs []string) ([]ItemRecord, error)

	// 获取用户数据
	GetUserData(ctx context.Context, userID string) (*UserRecord, error)

	// 获取热门物品
	GetPopularItems(ctx context.Context, category string, limit int) ([]ItemRecord, error)

	// 获取相似用户
	GetSimilarUsers(ctx context.Context, userID string, limit int) ([]SimilarUserRecord, error)

	// 健康检查
	HealthCheck(ctx context.Context) error

	// 获取数据源名称
	GetName() string

	// 关闭数据源
	Close() error
}

// UserBehaviorRecord 用户行为记录
type UserBehaviorRecord struct {
	UserID    string
	ItemID    string
	Behavior  string
	Value     float64
	Timestamp time.Time
	Context   map[string]interface{}
}

// ItemRecord 物品记录
type ItemRecord struct {
	ItemID      string
	Category    string
	Title       string
	Description string
	Features    map[string]interface{}
	Metadata    map[string]interface{}
	Popularity  float64
}

// UserRecord 用户记录
type UserRecord struct {
	UserID        string
	Demographics  map[string]interface{}
	Preferences   map[string]interface{}
	BehaviorStats map[string]interface{}
}

// SimilarUserRecord 相似用户记录
type SimilarUserRecord struct {
	UserID     string
	Similarity float64
}

// RecallResult 召回结果
type RecallResult struct {
	Items    []ItemRecord
	Score    float64
	Source   string
	Metadata map[string]interface{}
}
