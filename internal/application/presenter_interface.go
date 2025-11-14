// Package application 推荐服务接口
package application

import (
	"context"
	
	"recommendation-system/internal/domain"
)

// RecommendationUseCase 推荐服务用例接口
type RecommendationUseCase interface {
	// 获取推荐
	GetRecommendations(ctx context.Context, userID string, count int) ([]domain.Recommendation, error)
	
	// 按类别获取推荐
	GetRecommendationsByCategory(ctx context.Context, userID string, category string, count int) ([]domain.Recommendation, error)
}