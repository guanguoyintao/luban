// Package application 推荐服务实现
package application

import (
	"context"
	"fmt"

	"github.com/guanguoyintao/luban/internal/domain"
)

// RecommendationPresenter 推荐服务实现
type RecommendationPresenter struct {
	recommendationService domain.RecommendationService
}

// NewRecommendationPresenter 创建推荐服务
func NewRecommendationPresenter(service domain.RecommendationService) *RecommendationPresenter {
	return &RecommendationPresenter{
		recommendationService: service,
	}
}

// GetRecommendations 获取推荐
func (p *RecommendationPresenter) GetRecommendations(ctx context.Context, userID string, count int) ([]domain.Recommendation, error) {
	if userID == "" {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	if count <= 0 {
		count = 10 // 默认推荐数量
	}

	return p.recommendationService.GetRecommendations(ctx, userID, count)
}

// GetRecommendationsByCategory 按类别获取推荐
func (p *RecommendationPresenter) GetRecommendationsByCategory(ctx context.Context, userID string, category string, count int) ([]domain.Recommendation, error) {
	if userID == "" {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	if category == "" {
		return nil, fmt.Errorf("类别不能为空")
	}

	if count <= 0 {
		count = 10 // 默认推荐数量
	}

	// 这里可以添加按类别筛选的逻辑
	recommendations, err := p.recommendationService.GetRecommendations(ctx, userID, count)
	if err != nil {
		return nil, err
	}

	// 按类别筛选（简化实现）
	var filteredRecommendations []domain.Recommendation
	for _, rec := range recommendations {
		if rec.Category == category {
			filteredRecommendations = append(filteredRecommendations, rec)
		}
	}

	return filteredRecommendations, nil
}
