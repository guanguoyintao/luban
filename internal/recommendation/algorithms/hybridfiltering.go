package algorithms

import (
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// 混合过滤推荐算法
type HybridFilteringEngine struct {
	mu              sync.RWMutex
	collaborative   *CollaborativeFilteringEngine   // 协同过滤引擎
	contentBased    *ContentBasedFilteringEngine    // 基于内容过滤引擎
	weights         map[string]float64              // 算法权重
	log             *logrus.Logger
	config          *HybridFilteringConfig
}

// 混合过滤配置
type HybridFilteringConfig struct {
	CollaborativeWeight float64 // 协同过滤权重
	ContentBasedWeight  float64 // 内容过滤权重
	DiversityWeight     float64 // 多样性权重
	PopularityWeight      float64 // 流行度权重
	RecencyWeight        float64 // 时效性权重
	EnableDiversity      bool    // 是否启用多样性
	EnablePopularity     bool    // 是否启用流行度
	EnableRecency        bool    // 是否启用时效性
}

// 混合推荐结果
type HybridRecommendation struct {
	ItemID          string
	Score           float64
	CollaborativeScore float64
	ContentBasedScore  float64
	DiversityScore     float64
	PopularityScore    float64
	RecencyScore       float64
	Confidence         float64
	Reason             string
}

// 创建新的混合过滤引擎
func NewHybridFilteringEngine(collaborative *CollaborativeFilteringEngine, contentBased *ContentBasedFilteringEngine, log *logrus.Logger) *HybridFilteringEngine {
	if log == nil {
		log = logrus.New()
	}
	
	config := &HybridFilteringConfig{
		CollaborativeWeight: 0.4,
		ContentBasedWeight:  0.4,
		DiversityWeight:     0.1,
		PopularityWeight:      0.05,
		RecencyWeight:        0.05,
		EnableDiversity:      true,
		EnablePopularity:     true,
		EnableRecency:        true,
	}
	
	return &HybridFilteringEngine{
		collaborative: collaborative,
		contentBased:  contentBased,
		weights:       make(map[string]float64),
		log:           log,
		config:        config,
	}
}

// 生成混合推荐
func (h *HybridFilteringEngine) GenerateRecommendations(userID string, topN int) []HybridRecommendation {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	// 获取协同过滤推荐
	collaborativeRecs := h.collaborative.UserBasedRecommend(userID, topN*2)
	
	// 获取内容过滤推荐
	contentBasedRecs := h.contentBased.GenerateRecommendations(userID, topN*2)
	
	// 合并推荐结果
	allRecommendations := h.mergeRecommendations(collaborativeRecs, contentBasedRecs)
	
	// 计算混合得分
	hybridRecs := h.calculateHybridScores(userID, allRecommendations)
	
	// 应用多样性优化
	if h.config.EnableDiversity {
		hybridRecs = h.applyDiversityOptimization(hybridRecs)
	}
	
	// 按最终得分排序
	sort.Slice(hybridRecs, func(i, j int) bool {
		return hybridRecs[i].Score > hybridRecs[j].Score
	})
	
	// 返回前N个推荐
	if len(hybridRecs) > topN {
		hybridRecs = hybridRecs[:topN]
	}
	
	return hybridRecs
}

// 合并推荐结果
func (h *HybridFilteringEngine) mergeRecommendations(collaborativeRecs []Recommendation, contentBasedRecs []Recommendation) map[string]HybridRecommendation {
	merged := make(map[string]HybridRecommendation)
	
	// 处理协同过滤推荐
	for _, rec := range collaborativeRecs {
		if _, exists := merged[rec.ItemID]; !exists {
			merged[rec.ItemID] = HybridRecommendation{
				ItemID:             rec.ItemID,
				CollaborativeScore: rec.Score,
				ContentBasedScore:  0.0,
			}
		} else {
			hybridRec := merged[rec.ItemID]
			hybridRec.CollaborativeScore = rec.Score
			merged[rec.ItemID] = hybridRec
		}
	}
	
	// 处理内容过滤推荐
	for _, rec := range contentBasedRecs {
		if _, exists := merged[rec.ItemID]; !exists {
			merged[rec.ItemID] = HybridRecommendation{
				ItemID:             rec.ItemID,
				CollaborativeScore: 0.0,
				ContentBasedScore:  rec.Score,
			}
		} else {
			hybridRec := merged[rec.ItemID]
			hybridRec.ContentBasedScore = rec.Score
			merged[rec.ItemID] = hybridRec
		}
	}
	
	return merged
}

// 计算混合得分
func (h *HybridFilteringEngine) calculateHybridScores(userID string, recommendations map[string]HybridRecommendation) []HybridRecommendation {
	results := make([]HybridRecommendation, 0, len(recommendations))
	
	for _, rec := range recommendations {
		// 基础混合得分
		baseScore := h.config.CollaborativeWeight*rec.CollaborativeScore + 
					h.config.ContentBasedWeight*rec.ContentBasedScore
		
		// 计算多样性得分
		diversityScore := 0.0
		if h.config.EnableDiversity {
			diversityScore = h.calculateDiversityScore(userID, rec.ItemID)
		}
		
		// 计算流行度得分
		popularityScore := 0.0
		if h.config.EnablePopularity {
			popularityScore = h.calculatePopularityScore(rec.ItemID)
		}
		
		// 计算时效性得分
		recencyScore := 0.0
		if h.config.EnableRecency {
			recencyScore = h.calculateRecencyScore(rec.ItemID)
		}
		
		// 最终得分
		finalScore := baseScore + 
					h.config.DiversityWeight*diversityScore + 
					h.config.PopularityWeight*popularityScore + 
					h.config.RecencyWeight*recencyScore
		
		rec.Score = finalScore
		rec.DiversityScore = diversityScore
		rec.PopularityScore = popularityScore
		rec.RecencyScore = recencyScore
		rec.Confidence = h.calculateConfidence(rec)
		rec.Reason = h.generateRecommendationReason(rec)
		
		results = append(results, rec)
	}
	
	return results
}

// 计算多样性得分
func (h *HybridFilteringEngine) calculateDiversityScore(userID string, itemID string) float64 {
	// 简单的多样性计算：基于物品类别的新颖性
	// 这里可以扩展为更复杂的多样性算法
	
	// 获取用户历史中的物品类别
	userHistory := h.collaborative.userItemMatrix[userID]
	if len(userHistory) == 0 {
		return 1.0 // 新用户，给予最高多样性得分
	}
	
	// 获取目标物品的类别
	itemFeatures, exists := h.contentBased.itemFeatures[itemID]
	if !exists {
		return 0.5
	}
	
	// 计算类别重复度
	categoryCount := make(map[string]int)
	for histItemID := range userHistory {
		if histItemFeatures, exists := h.contentBased.itemFeatures[histItemID]; exists {
			categoryCount[histItemFeatures.Category]++
		}
	}
	
	// 如果用户已经有很多同类别的物品，降低多样性得分
	histCategoryCount := categoryCount[itemFeatures.Category]
	maxCount := 0
	for _, count := range categoryCount {
		if count > maxCount {
			maxCount = count
		}
	}
	
	if maxCount == 0 {
		return 1.0
	}
	
	// 多样性得分 = 1 - (历史同类物品数 / 最大历史类别数)
	diversityScore := 1.0 - float64(histCategoryCount)/float64(maxCount)
	return math.Max(0.0, diversityScore)
}

// 计算流行度得分
func (h *HybridFilteringEngine) calculatePopularityScore(itemID string) float64 {
	// 基于物品被评分的次数计算流行度
	ratingCount := 0
	for _, userRatings := range h.collaborative.userItemMatrix {
		if _, exists := userRatings[itemID]; exists {
			ratingCount++
		}
	}
	
	// 归一化流行度得分
	maxRatingCount := 0
	for _, userRatings := range h.collaborative.userItemMatrix {
		if len(userRatings) > maxRatingCount {
			maxRatingCount = len(userRatings)
		}
	}
	
	if maxRatingCount == 0 {
		return 0.5
	}
	
	return float64(ratingCount) / float64(maxRatingCount)
}

// 计算时效性得分
func (h *HybridFilteringEngine) calculateRecencyScore(itemID string) float64 {
	// 基于物品被评分的最近时间计算时效性
	// 这里简化处理，实际应用中需要记录评分时间
	
	// 获取物品被评分的用户数作为时效性的简单指标
	ratingCount := 0
	for _, userRatings := range h.collaborative.userItemMatrix {
		if _, exists := userRatings[itemID]; exists {
			ratingCount++
		}
	}
	
	// 简单的时效性计算：评分用户数越多，时效性越高
	return math.Min(1.0, float64(ratingCount)/100.0)
}

// 计算置信度
func (h *HybridFilteringEngine) calculateConfidence(rec HybridRecommendation) float64 {
	// 基于协同过滤和内容过滤的得分一致性计算置信度
	collaborativeScore := rec.CollaborativeScore
	contentBasedScore := rec.ContentBasedScore
	
	// 如果两个算法都给出了较高的得分，置信度较高
	if collaborativeScore > 0.5 && contentBasedScore > 0.5 {
		return 0.9
	}
	
	// 如果两个算法得分差异很大，置信度较低
	scoreDiff := math.Abs(collaborativeScore - contentBasedScore)
	if scoreDiff > 0.8 {
		return 0.3
	}
	
	// 默认置信度
	return 0.7
}

// 生成推荐理由
func (h *HybridFilteringEngine) generateRecommendationReason(rec HybridRecommendation) string {
	reasons := []string{}
	
	if rec.CollaborativeScore > 0.5 {
		reasons = append(reasons, "基于您的历史偏好")
	}
	
	if rec.ContentBasedScore > 0.5 {
		reasons = append(reasons, "与您喜欢的内容相似")
	}
	
	if rec.DiversityScore > 0.7 {
		reasons = append(reasons, "为您推荐新类型")
	}
	
	if rec.PopularityScore > 0.7 {
		reasons = append(reasons, "热门推荐")
	}
	
	if len(reasons) == 0 {
		return "为您推荐"
	}
	
	return strings.Join(reasons, "，")
}

// 应用多样性优化
func (h *HybridFilteringEngine) applyDiversityOptimization(recommendations []HybridRecommendation) []HybridRecommendation {
	if len(recommendations) <= 1 {
		return recommendations
	}
	
	// 简单的多样性优化：确保推荐列表中不同类别的物品
	optimized := []HybridRecommendation{recommendations[0]}
	selectedCategories := make(map[string]bool)
	
	// 记录第一个物品的类别
	if itemFeatures, exists := h.contentBased.itemFeatures[recommendations[0].ItemID]; exists {
		selectedCategories[itemFeatures.Category] = true
	}
	
	// 从剩余的推荐中选择多样性较高的物品
	for i := 1; i < len(recommendations); i++ {
		bestIdx := -1
		bestDiversityScore := -1.0
		
		for j := i; j < len(recommendations); j++ {
			itemFeatures, exists := h.contentBased.itemFeatures[recommendations[j].ItemID]
			if !exists {
				continue
			}
			
			// 如果类别已经存在，降低多样性得分
			diversityScore := recommendations[j].DiversityScore
			if selectedCategories[itemFeatures.Category] {
				diversityScore *= 0.5
			}
			
			if diversityScore > bestDiversityScore {
				bestDiversityScore = diversityScore
				bestIdx = j
			}
		}
		
		if bestIdx != -1 {
			// 交换位置
			recommendations[i], recommendations[bestIdx] = recommendations[bestIdx], recommendations[i]
			
			// 记录选择的类别
			if itemFeatures, exists := h.contentBased.itemFeatures[recommendations[i].ItemID]; exists {
				selectedCategories[itemFeatures.Category] = true
			}
			
			optimized = append(optimized, recommendations[i])
		}
	}
	
	return optimized
}

// 更新权重
func (h *HybridFilteringEngine) UpdateWeights(weights map[string]float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.weights = weights
	
	h.log.WithFields(logrus.Fields{
		"collaborative_weight": weights["collaborative"],
		"content_based_weight": weights["content_based"],
		"diversity_weight":     weights["diversity"],
		"popularity_weight":    weights["popularity"],
		"recency_weight":       weights["recency"],
	}).Info("更新混合过滤权重")
}

// 获取权重
func (h *HybridFilteringEngine) GetWeights() map[string]float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	weights := make(map[string]float64)
	weights["collaborative"] = h.config.CollaborativeWeight
	weights["content_based"] = h.config.ContentBasedWeight
	weights["diversity"] = h.config.DiversityWeight
	weights["popularity"] = h.config.PopularityWeight
	weights["recency"] = h.config.RecencyWeight
	
	return weights
}

// 设置配置
func (h *HybridFilteringEngine) SetConfig(config *HybridFilteringConfig) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.config = config
	h.log.Info("更新混合过滤配置")
}

// 获取配置
func (h *HybridFilteringEngine) GetConfig() *HybridFilteringConfig {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	return h.config
}

// 获取算法性能统计
func (h *HybridFilteringEngine) GetPerformanceStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	stats := make(map[string]interface{})
	
	// 协同过滤统计
	collaborativeStats := map[string]interface{}{
		"user_count": len(h.collaborative.userItemMatrix),
		"item_count": len(h.collaborative.itemUserMatrix),
	}
	stats["collaborative"] = collaborativeStats
	
	// 内容过滤统计
	contentBasedStats := map[string]interface{}{
		"user_profile_count": len(h.contentBased.userProfiles),
		"item_feature_count": len(h.contentBased.itemFeatures),
	}
	stats["content_based"] = contentBasedStats
	
	// 权重配置
	stats["weights"] = h.GetWeights()
	
	return stats
}