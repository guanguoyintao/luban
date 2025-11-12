package algorithms

import (
	"context"
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// 基于内容过滤推荐算法
type ContentBasedFilteringEngine struct {
	mu              sync.RWMutex
	userProfiles    map[string]UserProfile      // 用户画像
	itemFeatures    map[string]ItemFeatures     // 物品特征
	userItemHistory map[string]map[string]float64 // 用户-物品历史交互
	log             *logrus.Logger
	config          *ContentBasedFilteringConfig
}

// 内容过滤配置
type ContentBasedFilteringConfig struct {
	FeatureWeightThreshold float64 // 特征权重阈值
	MaxFeatures            int     // 最大特征数
	SimilarityThreshold    float64 // 相似度阈值
	LearningRate           float64 // 学习率
	DecayFactor            float64 // 衰减因子
}

// 用户画像
type UserProfile struct {
	UserID        string                 // 用户ID
	FeatureVector map[string]float64    // 特征向量
	Preferences   map[string]float64    // 偏好权重
	UpdateTime    int64                  // 更新时间
}

// 物品特征
type ItemFeatures struct {
	ItemID      string                 // 物品ID
	Category    string                 // 类别
	Keywords    []string               // 关键词
	Features    map[string]float64     // 特征向量
	Metadata    map[string]interface{} // 元数据
}

// 创建新的内容过滤引擎
func NewContentBasedFilteringEngine(log *logrus.Logger) *ContentBasedFilteringEngine {
	if log == nil {
		log = logrus.New()
	}
	
	config := &ContentBasedFilteringConfig{
		FeatureWeightThreshold: 0.1,
		MaxFeatures:            100,
		SimilarityThreshold:    0.3,
		LearningRate:           0.01,
		DecayFactor:            0.95,
	}
	
	return &ContentBasedFilteringEngine{
		userProfiles:    make(map[string]UserProfile),
		itemFeatures:    make(map[string]ItemFeatures),
		userItemHistory: make(map[string]map[string]float64),
		log:             log,
		config:          config,
	}
}

// 添加物品特征
func (c *ContentBasedFilteringEngine) AddItemFeatures(itemID string, category string, keywords []string, features map[string]float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	item := ItemFeatures{
		ItemID:      itemID,
		Category:    category,
		Keywords:    keywords,
		Features:    features,
		Metadata:    make(map[string]interface{}),
	}
	
	c.itemFeatures[itemID] = item
	
	c.log.WithFields(logrus.Fields{
		"item_id":  itemID,
		"category": category,
		"keywords": len(keywords),
		"features": len(features),
	}).Debug("添加物品特征")
}

// 添加用户行为（更新用户画像）
func (c *ContentBasedFilteringEngine) AddUserBehavior(userID string, itemID string, rating float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// 更新用户-物品历史
	if c.userItemHistory[userID] == nil {
		c.userItemHistory[userID] = make(map[string]float64)
	}
	c.userItemHistory[userID][itemID] = rating
	
	// 获取物品特征
	itemFeatures, exists := c.itemFeatures[itemID]
	if !exists {
		c.log.WithField("item_id", itemID).Warn("物品特征不存在")
		return
	}
	
	// 更新用户画像
	c.updateUserProfile(userID, itemFeatures, rating)
	
	c.log.WithFields(logrus.Fields{
		"user_id": userID,
		"item_id": itemID,
		"rating":  rating,
	}).Debug("添加用户行为")
}

// 更新用户画像
func (c *ContentBasedFilteringEngine) updateUserProfile(userID string, itemFeatures ItemFeatures, rating float64) {
	profile, exists := c.userProfiles[userID]
	if !exists {
		profile = UserProfile{
			UserID:        userID,
			FeatureVector: make(map[string]float64),
			Preferences:   make(map[string]float64),
			UpdateTime:    0,
		}
	}
	
	// 特征衰减
	c.applyFeatureDecay(&profile)
	
	// 更新特征向量
	for feature, value := range itemFeatures.Features {
		if profile.FeatureVector[feature] == 0 {
			profile.FeatureVector[feature] = 0
		}
		profile.FeatureVector[feature] += c.config.LearningRate * rating * value
	}
	
	// 更新关键词偏好
	for _, keyword := range itemFeatures.Keywords {
		if profile.Preferences[keyword] == 0 {
			profile.Preferences[keyword] = 0
		}
		profile.Preferences[keyword] += c.config.LearningRate * rating
	}
	
	// 更新类别偏好
	if profile.Preferences[itemFeatures.Category] == 0 {
		profile.Preferences[itemFeatures.Category] = 0
	}
	profile.Preferences[itemFeatures.Category] += c.config.LearningRate * rating
	
	profile.UpdateTime = getCurrentTimestamp()
	c.userProfiles[userID] = profile
}

// 应用特征衰减
func (c *ContentBasedFilteringEngine) applyFeatureDecay(profile *UserProfile) {
	currentTime := getCurrentTimestamp()
	timeDiff := currentTime - profile.UpdateTime
	
	if timeDiff > 0 {
		decayFactor := math.Pow(c.config.DecayFactor, float64(timeDiff)/86400) // 按天衰减
		
		// 衰减特征向量
		for feature, value := range profile.FeatureVector {
			profile.FeatureVector[feature] = value * decayFactor
		}
		
		// 衰减偏好
		for preference, weight := range profile.Preferences {
			profile.Preferences[preference] = weight * decayFactor
		}
	}
}

// 生成推荐
func (c *ContentBasedFilteringEngine) GenerateRecommendations(userID string, topN int) []Recommendation {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	profile, exists := c.userProfiles[userID]
	if !exists {
		return []Recommendation{}
	}
	
	recommendations := make([]Recommendation, 0)
	userHistory := c.userItemHistory[userID]
	
	// 对所有物品计算相似度
	for itemID, itemFeatures := range c.itemFeatures {
		// 跳过用户已经交互过的物品
		if _, exists := userHistory[itemID]; exists {
			continue
		}
		
		// 计算用户画像与物品特征的相似度
		similarity := c.calculateSimilarity(profile, itemFeatures)
		
		if similarity >= c.config.SimilarityThreshold {
			recommendations = append(recommendations, Recommendation{
				ItemID: itemID,
				Score:  similarity,
			})
		}
	}
	
	// 按相似度排序
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})
	
	// 返回前N个推荐
	if len(recommendations) > topN {
		recommendations = recommendations[:topN]
	}
	
	return recommendations
}

// 计算用户画像与物品特征的相似度
func (c *ContentBasedFilteringEngine) calculateSimilarity(profile UserProfile, item ItemFeatures) float64 {
	// 特征向量相似度
	featureSimilarity := c.calculateFeatureSimilarity(profile.FeatureVector, item.Features)
	
	// 关键词相似度
	keywordSimilarity := c.calculateKeywordSimilarity(profile.Preferences, item.Keywords)
	
	// 类别相似度
	categorySimilarity := c.calculateCategorySimilarity(profile.Preferences, item.Category)
	
	// 综合相似度（加权平均）
	totalSimilarity := 0.5*featureSimilarity + 0.3*keywordSimilarity + 0.2*categorySimilarity
	
	return totalSimilarity
}

// 计算特征向量相似度（余弦相似度）
func (c *ContentBasedFilteringEngine) calculateFeatureSimilarity(userFeatures map[string]float64, itemFeatures map[string]float64) float64 {
	// 找到共同特征
	var dotProduct, normUser, normItem float64
	
	for feature, userValue := range userFeatures {
		if itemValue, exists := itemFeatures[feature]; exists {
			dotProduct += userValue * itemValue
		}
		normUser += userValue * userValue
	}
	
	for _, itemValue := range itemFeatures {
		normItem += itemValue * itemValue
	}
	
	// 避免除零
	if normUser == 0 || normItem == 0 {
		return 0.0
	}
	
	return dotProduct / (math.Sqrt(normUser) * math.Sqrt(normItem))
}

// 计算关键词相似度
func (c *ContentBasedFilteringEngine) calculateKeywordSimilarity(userPreferences map[string]float64, itemKeywords []string) float64 {
	if len(itemKeywords) == 0 {
		return 0.0
	}
	
	var totalScore float64
	matchedCount := 0
	
	for _, keyword := range itemKeywords {
		if preference, exists := userPreferences[keyword]; exists {
			totalScore += preference
			matchedCount++
		}
	}
	
	if matchedCount == 0 {
		return 0.0
	}
	
	// 平均匹配得分
	return totalScore / float64(len(itemKeywords))
}

// 计算类别相似度
func (c *ContentBasedFilteringEngine) calculateCategorySimilarity(userPreferences map[string]float64, category string) float64 {
	if preference, exists := userPreferences[category]; exists {
		return preference
	}
	return 0.0
}

// 获取用户画像
func (c *ContentBasedFilteringEngine) GetUserProfile(userID string) (*UserProfile, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	profile, exists := c.userProfiles[userID]
	return &profile, exists
}

// 更新用户偏好
func (c *ContentBasedFilteringEngine) UpdateUserPreference(userID string, preference string, weight float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	profile, exists := c.userProfiles[userID]
	if !exists {
		profile = UserProfile{
			UserID:        userID,
			FeatureVector: make(map[string]float64),
			Preferences:   make(map[string]float64),
			UpdateTime:    getCurrentTimestamp(),
		}
	}
	
	profile.Preferences[preference] = weight
	profile.UpdateTime = getCurrentTimestamp()
	c.userProfiles[userID] = profile
	
	c.log.WithFields(logrus.Fields{
		"user_id":    userID,
		"preference": preference,
		"weight":     weight,
	}).Debug("更新用户偏好")
}

// 获取热门关键词
func (c *ContentBasedFilteringEngine) GetPopularKeywords(limit int) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	keywordCount := make(map[string]int)
	
	// 统计关键词出现频率
	for _, item := range c.itemFeatures {
		for _, keyword := range item.Keywords {
			keywordCount[keyword]++
		}
	}
	
	// 转换为排序列表
	type KeywordCount struct {
		Keyword string
		Count   int
	}
	
	keywordList := make([]KeywordCount, 0, len(keywordCount))
	for keyword, count := range keywordCount {
		keywordList = append(keywordList, KeywordCount{keyword, count})
	}
	
	// 按出现频率排序
	sort.Slice(keywordList, func(i, j int) bool {
		return keywordList[i].Count > keywordList[j].Count
	})
	
	// 返回前N个热门关键词
	result := make([]string, 0, limit)
	for i := 0; i < limit && i < len(keywordList); i++ {
		result = append(result, keywordList[i].Keyword)
	}
	
	return result
}

// 文本预处理
func (c *ContentBasedFilteringEngine) preprocessText(text string) []string {
	// 转换为小写
	text = strings.ToLower(text)
	
	// 简单的分词（可以替换为更复杂的NLP处理）
	words := strings.Fields(text)
	
	// 去除停用词（简单的示例）
	stopWords := map[string]bool{
		"the": true, "is": true, "at": true, "which": true, "on": true,
	}
	
	result := []string{}
	for _, word := range words {
		if !stopWords[word] && len(word) > 2 {
			result = append(result, word)
		}
	}
	
	return result
}

// 获取当前时间戳
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}