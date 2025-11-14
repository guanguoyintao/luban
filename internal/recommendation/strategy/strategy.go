// Package strategy 推荐排序策略模式
// 用于推荐结果的不同排序策略
type package strategy

import (
	"context"
	"sort"
	
	"recommendation-system/internal/domain"
)

// RankingStrategy 排序策略接口
type RankingStrategy interface {
	Rank(ctx context.Context, recommendations []domain.Recommendation, userID string) ([]domain.Recommendation, error)
	GetName() string
	GetDescription() string
}

// ScoreBasedStrategy 基于分数的排序策略
type ScoreBasedStrategy struct{}

// NewScoreBasedStrategy 创建基于分数的排序策略
func NewScoreBasedStrategy() *ScoreBasedStrategy {
	return &ScoreBasedStrategy{}
}

func (s *ScoreBasedStrategy) Rank(ctx context.Context, recommendations []domain.Recommendation, userID string) ([]domain.Recommendation, error) {
	// 按分数降序排序
	sorted := make([]domain.Recommendation, len(recommendations))
	copy(sorted, recommendations)
	
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})
	
	return sorted, nil
}

func (s *ScoreBasedStrategy) GetName() string {
	return "score_based"
}

func (s *ScoreBasedStrategy) GetDescription() string {
	return "基于推荐分数的排序策略"
}

// DiversityStrategy 多样性排序策略
type DiversityStrategy struct {
	maxSameCategory int
}

// NewDiversityStrategy 创建多样性排序策略
func NewDiversityStrategy() *DiversityStrategy {
	return &DiversityStrategy{
		maxSameCategory: 2, // 同一类别最多2个
	}
}

func (s *DiversityStrategy) Rank(ctx context.Context, recommendations []domain.Recommendation, userID string) ([]domain.Recommendation, error) {
	if len(recommendations) <= 1 {
		return recommendations, nil
	}
	
	// 先按分数排序
	sorted := make([]domain.Recommendation, len(recommendations))
	copy(sorted, recommendations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})
	
	// 应用多样性策略
	result := make([]domain.Recommendation, 0)
	categoryCount := make(map[string]int)
	
	for _, rec := range sorted {
		count := categoryCount[rec.Category]
		if count < s.maxSameCategory {
			result = append(result, rec)
			categoryCount[rec.Category] = count + 1
		}
	}
	
	return result, nil
}

func (s *DiversityStrategy) GetName() string {
	return "diversity"
}

func (s *DiversityStrategy) GetDescription() string {
	return "基于多样性的排序策略，避免同一类别过多推荐"
}

// NoveltyStrategy 新颖性排序策略
type NoveltyStrategy struct {
	userHistory map[string]map[string]bool // userID -> itemID -> seen
}

// NewNoveltyStrategy 创建新颖性排序策略
func NewNoveltyStrategy() *NoveltyStrategy {
	return &NoveltyStrategy{
		userHistory: make(map[string]map[string]bool),
	}
}

func (s *NoveltyStrategy) Rank(ctx context.Context, recommendations []domain.Recommendation, userID string) ([]domain.Recommendation, error) {
	if len(recommendations) <= 1 {
		return recommendations, nil
	}
	
	// 初始化用户历史记录
	if _, exists := s.userHistory[userID]; !exists {
		s.userHistory[userID] = make(map[string]bool)
	}
	
	// 过滤掉已看过的物品
	novelRecommendations := make([]domain.Recommendation, 0)
	for _, rec := range recommendations {
		if !s.userHistory[userID][rec.ItemID] {
			novelRecommendations = append(novelRecommendations, rec)
		}
	}
	
	// 如果没有新颖的推荐，返回原始推荐
	if len(novelRecommendations) == 0 {
		return recommendations, nil
	}
	
	// 按分数排序
	sort.Slice(novelRecommendations, func(i, j int) bool {
		return novelRecommendations[i].Score > novelRecommendations[j].Score
	})
	
	return novelRecommendations, nil
}

func (s *NoveltyStrategy) GetName() string {
	return "novelty"
}

func (s *NoveltyStrategy) GetDescription() string {
	return "基于新颖性的排序策略，优先推荐用户未看过的物品"
}

// PersonalizationStrategy 个性化排序策略
type PersonalizationStrategy struct {
	userProfiles map[string]UserProfile
}

type UserProfile struct {
	PreferredCategories []string
	PreferredAlgorithms []string
}

// NewPersonalizationStrategy 创建个性化排序策略
func NewPersonalizationStrategy() *PersonalizationStrategy {
	return &PersonalizationStrategy{
		userProfiles: make(map[string]UserProfile),
	}
}

func (s *PersonalizationStrategy) Rank(ctx context.Context, recommendations []domain.Recommendation, userID string) ([]domain.Recommendation, error) {
	if len(recommendations) <= 1 {
		return recommendations, nil
	}
	
	// 获取用户画像
	profile, exists := s.userProfiles[userID]
	if !exists {
		// 如果没有用户画像，使用默认策略
		return s.defaultRank(recommendations), nil
	}
	
	// 计算每个推荐的个性化分数
	scoredRecommendations := make([]ScoredRecommendation, 0)
	for _, rec := range recommendations {
		score := rec.Score
		
		// 类别偏好加分
		for _, prefCat := range profile.PreferredCategories {
			if rec.Category == prefCat {
				score += 0.1
				break
			}
		}
		
		// 算法偏好加分
		for _, prefAlgo := range profile.PreferredAlgorithms {
			if rec.Algorithm == prefAlgo {
				score += 0.05
				break
			}
		}
		
		scoredRecommendations = append(scoredRecommendations, ScoredRecommendation{
			Recommendation: rec,
			Score:         score,
		})
	}
	
	// 按个性化分数排序
	sort.Slice(scoredRecommendations, func(i, j int) bool {
		return scoredRecommendations[i].Score > scoredRecommendations[j].Score
	})
	
	// 提取排序后的推荐
	result := make([]domain.Recommendation, len(scoredRecommendations))
	for i, scored := range scoredRecommendations {
		result[i] = scored.Recommendation
	}
	
	return result, nil
}

func (s *PersonalizationStrategy) GetName() string {
	return "personalization"
}

func (s *PersonalizationStrategy) GetDescription() string {
	return "基于用户画像的个性化排序策略"
}

func (s *PersonalizationStrategy) defaultRank(recommendations []domain.Recommendation) []domain.Recommendation {
	sorted := make([]domain.Recommendation, len(recommendations))
	copy(sorted, recommendations)
	
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})
	
	return sorted
}

// ScoredRecommendation 带分数的推荐
type ScoredRecommendation struct {
	Recommendation domain.Recommendation
	Score         float64
}

// SetUserProfile 设置用户画像
func (s *PersonalizationStrategy) SetUserProfile(userID string, profile UserProfile) {
	s.userProfiles[userID] = profile
}

// RankingContext 排序上下文
type RankingContext struct {
	Strategy  RankingStrategy
	UserID    string
	Context   context.Context
}

// NewRankingContext 创建排序上下文
func NewRankingContext(strategy RankingStrategy, userID string, ctx context.Context) *RankingContext {
	return &RankingContext{
		Strategy: strategy,
		UserID:   userID,
		Context:  ctx,
	}
}

// Execute 执行排序
func (rc *RankingContext) Execute(recommendations []domain.Recommendation) ([]domain.Recommendation, error) {
	return rc.Strategy.Rank(rc.Context, recommendations, rc.UserID)
}