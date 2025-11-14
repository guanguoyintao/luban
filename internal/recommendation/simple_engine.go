package recommendation

import (
	"context"
	"time"
	
	"github.com/sirupsen/logrus"
	
	"recommendation-system/internal/datacollection"
	"recommendation-system/internal/dataprocessing"
	"recommendation-system/internal/domain"
)

// SimpleRecommendationEngine 简单推荐引擎实现
type SimpleRecommendationEngine struct {
	logger        *logrus.Logger
	dataCollector datacollection.DataCollector
	dataProcessor dataprocessing.DataProcessor
}

// NewRecommendationEngine 创建推荐引擎
func NewRecommendationEngine(
	logger *logrus.Logger,
	dataCollector datacollection.DataCollector,
	dataProcessor dataprocessing.DataProcessor,
	contentBased interface{},
	collaborative interface{},
	hybrid interface{},
) *SimpleRecommendationEngine {
	return &SimpleRecommendationEngine{
		logger:        logger,
		dataCollector: dataCollector,
		dataProcessor: dataProcessor,
	}
}

// GetRecommendations 获取推荐
func (e *SimpleRecommendationEngine) GetRecommendations(ctx context.Context, userID string, count int) ([]domain.Recommendation, error) {
	e.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"count":   count,
	}).Info("开始生成推荐")
	
	// 创建用户数据
	userData := datacollection.UserData{
		UserID: userID,
		Demographics: map[string]interface{}{
			"age":    25,
			"gender": "male",
		},
		Preferences: map[string]interface{}{
			"categories": []string{"technology", "sports"},
		},
	}
	
	// 处理用户数据
	processedData, err := e.dataProcessor.CleanUserData(ctx, userData)
	if err != nil {
		e.logger.WithError(err).Error("处理用户数据失败")
		// 即使处理失败，我们也可以返回模拟推荐
		e.logger.Info("使用默认推荐数据")
	}
	
	_ = processedData // 使用处理后的数据
	
	// 模拟推荐结果
	recommendations := []domain.Recommendation{
		{
			ItemID:     "item_001",
			Score:      0.95,
			Reason:     "基于您的历史偏好推荐",
			Algorithm:  "hybrid_filtering",
			Confidence: 0.9,
			CreatedAt:  time.Now(),
			Category:   "technology",
		},
		{
			ItemID:     "item_002",
			Score:      0.87,
			Reason:     "与您相似的用户也喜欢",
			Algorithm:  "collaborative_filtering",
			Confidence: 0.8,
			CreatedAt:  time.Now(),
			Category:   "sports",
		},
		{
			ItemID:     "item_003",
			Score:      0.82,
			Reason:     "内容特征匹配",
			Algorithm:  "content_based_filtering",
			Confidence: 0.75,
			CreatedAt:  time.Now(),
			Category:   "technology",
		},
	}
	
	// 限制推荐数量
	if count > 0 && count < len(recommendations) {
		recommendations = recommendations[:count]
	}
	
	e.logger.WithFields(logrus.Fields{
		"user_id":         userID,
		"recommendations": len(recommendations),
	}).Info("推荐生成成功")
	
	return recommendations, nil
}

// GetRecommendationsByCategory 按类别获取推荐
func (e *SimpleRecommendationEngine) GetRecommendationsByCategory(ctx context.Context, userID string, category string, count int) ([]domain.Recommendation, error) {
	e.logger.WithFields(logrus.Fields{
		"user_id":  userID,
		"category": category,
		"count":    count,
	}).Info("开始按类别生成推荐")
	
	// 获取所有推荐
	recommendations, err := e.GetRecommendations(ctx, userID, count*2) // 获取更多以便筛选
	if err != nil {
		return nil, err
	}
	
	// 按类别筛选
	var filteredRecommendations []domain.Recommendation
	for _, rec := range recommendations {
		if rec.Category == category {
			filteredRecommendations = append(filteredRecommendations, rec)
		}
	}
	
	// 限制推荐数量
	if count > 0 && count < len(filteredRecommendations) {
		filteredRecommendations = filteredRecommendations[:count]
	}
	
	e.logger.WithFields(logrus.Fields{
		"user_id":         userID,
		"category":        category,
		"recommendations": len(filteredRecommendations),
	}).Info("按类别推荐生成成功")
	
	return filteredRecommendations, nil
}