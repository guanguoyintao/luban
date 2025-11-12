package algorithms

import (
	"context"
	"math"
	"sort"
	"sync"

	"github.com/sirupsen/logrus"
)

// 协同过滤推荐算法
type CollaborativeFilteringEngine struct {
	mu              sync.RWMutex
	userItemMatrix  map[string]map[string]float64 // 用户-物品评分矩阵
	itemUserMatrix  map[string]map[string]float64 // 物品-用户评分矩阵
	userSimilarity  map[string]map[string]float64 // 用户相似度矩阵
	itemSimilarity  map[string]map[string]float64 // 物品相似度矩阵
	log             *logrus.Logger
	config          *CollaborativeFilteringConfig
}

// 协同过滤配置
type CollaborativeFilteringConfig struct {
	SimilarityThreshold float64 // 相似度阈值
	MaxNeighbors        int     // 最大邻居数
	MinCommonItems      int     // 最小共同物品数
	NormalizationMethod string  // 归一化方法
}

// 创建新的协同过滤引擎
func NewCollaborativeFilteringEngine(log *logrus.Logger) *CollaborativeFilteringEngine {
	if log == nil {
		log = logrus.New()
	}
	
	config := &CollaborativeFilteringConfig{
		SimilarityThreshold: 0.1,
		MaxNeighbors:        50,
		MinCommonItems:      2,
		NormalizationMethod: "mean_centering",
	}
	
	return &CollaborativeFilteringEngine{
		userItemMatrix: make(map[string]map[string]float64),
		itemUserMatrix: make(map[string]map[string]float64),
		userSimilarity: make(map[string]map[string]float64),
		itemSimilarity: make(map[string]map[string]float64),
		log:            log,
		config:         config,
	}
}

// 添加用户评分数据
func (c *CollaborativeFilteringEngine) AddUserRating(userID string, itemID string, rating float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// 更新用户-物品矩阵
	if c.userItemMatrix[userID] == nil {
		c.userItemMatrix[userID] = make(map[string]float64)
	}
	c.userItemMatrix[userID][itemID] = rating
	
	// 更新物品-用户矩阵
	if c.itemUserMatrix[itemID] == nil {
		c.itemUserMatrix[itemID] = make(map[string]float64)
	}
	c.itemUserMatrix[itemID][userID] = rating
	
	c.log.WithFields(logrus.Fields{
		"user_id": userID,
		"item_id": itemID,
		"rating":  rating,
	}).Debug("添加用户评分数据")
}

// 计算用户相似度（基于皮尔逊相关系数）
func (c *CollaborativeFilteringEngine) CalculateUserSimilarity(userID1 string, userID2 string) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	ratings1, exists1 := c.userItemMatrix[userID1]
	ratings2, exists2 := c.userItemMatrix[userID2]
	
	if !exists1 || !exists2 {
		return 0.0
	}
	
	// 找到共同评分的物品
	commonItems := []string{}
	for itemID := range ratings1 {
		if _, exists := ratings2[itemID]; exists {
			commonItems = append(commonItems, itemID)
		}
	}
	
	if len(commonItems) < c.config.MinCommonItems {
		return 0.0
	}
	
	// 计算皮尔逊相关系数
	return c.pearsonCorrelation(ratings1, ratings2, commonItems)
}

// 计算物品相似度（基于余弦相似度）
func (c *CollaborativeFilteringEngine) CalculateItemSimilarity(itemID1 string, itemID2 string) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	ratings1, exists1 := c.itemUserMatrix[itemID1]
	ratings2, exists2 := c.itemUserMatrix[itemID2]
	
	if !exists1 || !exists2 {
		return 0.0
	}
	
	// 计算余弦相似度
	return c.cosineSimilarity(ratings1, ratings2)
}

// 基于用户的协同过滤推荐
func (c *CollaborativeFilteringEngine) UserBasedRecommend(userID string, topN int) []Recommendation {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	userRatings, exists := c.userItemMatrix[userID]
	if !exists {
		return []Recommendation{}
	}
	
	// 计算与目标用户最相似的用户
	similarUsers := c.findSimilarUsers(userID)
	
	// 生成推荐
	recommendations := make(map[string]float64)
	
	for _, similarUser := range similarUsers {
		similarUserRatings := c.userItemMatrix[similarUser.UserID]
		
		for itemID, rating := range similarUserRatings {
			// 跳过用户已经评分过的物品
			if _, exists := userRatings[itemID]; exists {
				continue
			}
			
			// 累加加权评分
			if _, exists := recommendations[itemID]; !exists {
				recommendations[itemID] = 0.0
			}
			recommendations[itemID] += similarUser.Similarity * rating
		}
	}
	
	// 转换为推荐列表并排序
	result := make([]Recommendation, 0, len(recommendations))
	for itemID, score := range recommendations {
		result = append(result, Recommendation{
			ItemID: itemID,
			Score:  score,
		})
	}
	
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})
	
	// 返回前N个推荐
	if len(result) > topN {
		result = result[:topN]
	}
	
	return result
}

// 基于物品的协同过滤推荐
func (c *CollaborativeFilteringEngine) ItemBasedRecommend(userID string, topN int) []Recommendation {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	userRatings, exists := c.userItemMatrix[userID]
	if !exists {
		return []Recommendation{}
	}
	
	// 计算物品相似度矩阵（如果还没有计算）
	if len(c.itemSimilarity) == 0 {
		c.buildItemSimilarityMatrix()
	}
	
	recommendations := make(map[string]float64)
	
	// 对用户评分过的每个物品
	for userItemID, userRating := range userRatings {
		// 找到相似的物品
		similarItems := c.findSimilarItems(userItemID)
		
		for _, similarItem := range similarItems {
			// 跳过用户已经评分过的物品
			if _, exists := userRatings[similarItem.ItemID]; exists {
				continue
			}
			
			// 累加加权评分
			if _, exists := recommendations[similarItem.ItemID]; !exists {
				recommendations[similarItem.ItemID] = 0.0
			}
			recommendations[similarItem.ItemID] += similarItem.Similarity * userRating
		}
	}
	
	// 转换为推荐列表并排序
	result := make([]Recommendation, 0, len(recommendations))
	for itemID, score := range recommendations {
		result = append(result, Recommendation{
			ItemID: itemID,
			Score:  score,
		})
	}
	
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})
	
	// 返回前N个推荐
	if len(result) > topN {
		result = result[:topN]
	}
	
	return result
}

// 皮尔逊相关系数计算
func (c *CollaborativeFilteringEngine) pearsonCorrelation(ratings1 map[string]float64, ratings2 map[string]float64, commonItems []string) float64 {
	if len(commonItems) == 0 {
		return 0.0
	}
	
	// 计算平均值
	var sum1, sum2 float64
	for _, itemID := range commonItems {
		sum1 += ratings1[itemID]
		sum2 += ratings2[itemID]
	}
	mean1 := sum1 / float64(len(commonItems))
	mean2 := sum2 / float64(len(commonItems))
	
	// 计算分子和分母
	var numerator, denominator1, denominator2 float64
	for _, itemID := range commonItems {
		diff1 := ratings1[itemID] - mean1
		diff2 := ratings2[itemID] - mean2
		numerator += diff1 * diff2
		denominator1 += diff1 * diff1
		denominator2 += diff2 * diff2
	}
	
	// 避免除零
	if denominator1 == 0 || denominator2 == 0 {
		return 0.0
	}
	
	correlation := numerator / math.Sqrt(denominator1*denominator2)
	
	// 考虑共同物品数量的影响
	weight := float64(len(commonItems)) / float64(c.config.MaxNeighbors)
	return correlation * weight
}

// 余弦相似度计算
func (c *CollaborativeFilteringEngine) cosineSimilarity(ratings1 map[string]float64, ratings2 map[string]float64) float64 {
	// 找到共同用户
	commonUsers := []string{}
	for userID := range ratings1 {
		if _, exists := ratings2[userID]; exists {
			commonUsers = append(commonUsers, userID)
		}
	}
	
	if len(commonUsers) == 0 {
		return 0.0
	}
	
	// 计算点积和模
	var dotProduct, norm1, norm2 float64
	for _, userID := range commonUsers {
		rating1 := ratings1[userID]
		rating2 := ratings2[userID]
		dotProduct += rating1 * rating2
		norm1 += rating1 * rating1
		norm2 += rating2 * rating2
	}
	
	// 避免除零
	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}
	
	return dotProduct / math.Sqrt(norm1*norm2)
}

// 找到相似用户
func (c *CollaborativeFilteringEngine) findSimilarUsers(userID string) []SimilarUser {
	similarUsers := []SimilarUser{}
	
	for otherUserID := range c.userItemMatrix {
		if otherUserID == userID {
			continue
		}
		
		similarity := c.CalculateUserSimilarity(userID, otherUserID)
		if similarity >= c.config.SimilarityThreshold {
			similarUsers = append(similarUsers, SimilarUser{
				UserID:     otherUserID,
				Similarity: similarity,
			})
		}
	}
	
	// 按相似度排序
	sort.Slice(similarUsers, func(i, j int) bool {
		return similarUsers[i].Similarity > similarUsers[j].Similarity
	})
	
	// 限制邻居数量
	if len(similarUsers) > c.config.MaxNeighbors {
		similarUsers = similarUsers[:c.config.MaxNeighbors]
	}
	
	return similarUsers
}

// 找到相似物品
func (c *CollaborativeFilteringEngine) findSimilarItems(itemID string) []SimilarItem {
	similarItems := []SimilarItem{}
	
	for otherItemID := range c.itemUserMatrix {
		if otherItemID == itemID {
			continue
		}
		
		similarity := c.CalculateItemSimilarity(itemID, otherItemID)
		if similarity >= c.config.SimilarityThreshold {
			similarItems = append(similarItems, SimilarItem{
				ItemID:     otherItemID,
				Similarity: similarity,
			})
		}
	}
	
	// 按相似度排序
	sort.Slice(similarItems, func(i, j int) bool {
		return similarItems[i].Similarity > similarItems[j].Similarity
	})
	
	// 限制邻居数量
	if len(similarItems) > c.config.MaxNeighbors {
		similarItems = similarItems[:c.config.MaxNeighbors]
	}
	
	return similarItems
}

// 构建物品相似度矩阵
func (c *CollaborativeFilteringEngine) buildItemSimilarityMatrix() {
	c.itemSimilarity = make(map[string]map[string]float64)
	
	for itemID1 := range c.itemUserMatrix {
		c.itemSimilarity[itemID1] = make(map[string]float64)
		
		for itemID2 := range c.itemUserMatrix {
			if itemID1 == itemID2 {
				continue
			}
			
			similarity := c.CalculateItemSimilarity(itemID1, itemID2)
			c.itemSimilarity[itemID1][itemID2] = similarity
		}
	}
}

// 推荐结果
type Recommendation struct {
	ItemID string
	Score  float64
}

// 相似用户
type SimilarUser struct {
	UserID     string
	Similarity float64
}

// 相似物品
type SimilarItem struct {
	ItemID     string
	Similarity float64
}