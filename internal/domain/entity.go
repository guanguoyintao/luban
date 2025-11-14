// Package domain 定义推荐系统的核心业务领域
package domain

import (
	"context"
	"time"
)

// Recommendation 推荐结果实体
type Recommendation struct {
	ItemID     string    // 物品ID
	Score      float64   // 推荐得分
	Reason     string    // 推荐理由
	Algorithm  string    // 使用的算法
	Confidence float64   // 置信度
	CreatedAt  time.Time // 创建时间
	Category   string    // 类别
}

// RecommendationService 推荐服务接口
type RecommendationService interface {
	// 获取推荐
	GetRecommendations(ctx context.Context, userID string, count int) ([]Recommendation, error)
	
	// 按类别获取推荐
	GetRecommendationsByCategory(ctx context.Context, userID string, category string, count int) ([]Recommendation, error)
}