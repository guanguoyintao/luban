package recommendation

import (
	"context"
)

// 推荐算法类型
type AlgorithmType string

const (
	AlgorithmCollaborativeFiltering AlgorithmType = "collaborative_filtering" // 协同过滤
	AlgorithmContentBasedFiltering  AlgorithmType = "content_based_filtering"  // 基于内容过滤
	AlgorithmHybridFiltering        AlgorithmType = "hybrid_filtering"        // 混合过滤
	AlgorithmDeepLearning           AlgorithmType = "deep_learning"           // 深度学习
	AlgorithmRuleBased              AlgorithmType = "rule_based"              // 基于规则
)

// 推荐场景
type RecommendationScenario string

const (
	ScenarioHomePage       RecommendationScenario = "home_page"       // 首页推荐
	ScenarioProductDetail  RecommendationScenario = "product_detail"  // 商品详情页推荐
	ScenarioShoppingCart   RecommendationScenario = "shopping_cart"   // 购物车推荐
	ScenarioSearchResult   RecommendationScenario = "search_result"   // 搜索结果推荐
	ScenarioEmailMarketing RecommendationScenario = "email_marketing" // 邮件营销推荐
)

// 推荐请求
type RecommendationRequest struct {
	UserID     string                 // 用户ID
	Scenario   RecommendationScenario // 推荐场景
	Context    map[string]interface{} // 上下文信息
	Filters    map[string]interface{} // 过滤条件
	Limit      int                    // 推荐数量限制
	Algorithm  AlgorithmType          // 指定算法类型
	Parameters map[string]interface{} // 算法参数
}

// 推荐结果
type RecommendationResult struct {
	ItemID          string                 // 物品ID
	Score           float64                // 推荐得分
	Reason          string                 // 推荐理由
	Algorithm       AlgorithmType          // 使用的算法
	Confidence      float64                // 置信度
	Metadata        map[string]interface{} // 元数据
}

// 推荐响应
type RecommendationResponse struct {
	UserID          string                 // 用户ID
	Recommendations []RecommendationResult // 推荐结果列表
	TotalCount      int                    // 总推荐数量
	Algorithm       AlgorithmType          // 实际使用的算法
	ProcessingTime  int64                  // 处理时间（毫秒）
	Metadata        map[string]interface{} // 元数据
}

// 推荐引擎接口
type RecommendationEngine interface {
	// 生成推荐
	Recommend(ctx context.Context, request RecommendationRequest) (*RecommendationResponse, error)
	
	// 批量生成推荐
	RecommendBatch(ctx context.Context, requests []RecommendationRequest) ([]*RecommendationResponse, error)
	
	// 获取推荐解释
	ExplainRecommendation(ctx context.Context, userID string, itemID string) (string, error)
	
	// 更新推荐模型
	UpdateModel(ctx context.Context, data interface{}) error
	
	// 获取推荐算法列表
	GetAvailableAlgorithms(ctx context.Context) ([]AlgorithmType, error)
	
	// 获取算法参数
	GetAlgorithmParameters(ctx context.Context, algorithm AlgorithmType) (map[string]interface{}, error)
	
	// 设置算法参数
	SetAlgorithmParameters(ctx context.Context, algorithm AlgorithmType, parameters map[string]interface{}) error
	
	// 获取推荐统计信息
	GetRecommendationStats(ctx context.Context, userID string) (map[string]interface{}, error)
	
	// 记录用户反馈
	RecordFeedback(ctx context.Context, userID string, itemID string, feedback interface{}) error
	
	// 关闭推荐引擎
	Close() error
}