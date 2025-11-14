// Package datasource 内存数据源适配器实现
package datasource

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MemoryDataSource 内存数据源
type MemoryDataSource struct {
	name           string
	log            *logrus.Logger
	mu             sync.RWMutex
	userBehaviors  map[string][]UserBehaviorRecord
	items          map[string]ItemRecord
	users          map[string]UserRecord
	popularItems   map[string][]ItemRecord // category -> items
}

// NewMemoryDataSource 创建内存数据源
func NewMemoryDataSource(config DataSourceConfig, log *logrus.Logger) *MemoryDataSource {
	if log == nil {
		log = logrus.New()
	}
	
	ds := &MemoryDataSource{
		name:          config.Name,
		log:           log,
		userBehaviors: make(map[string][]UserBehaviorRecord),
		items:         make(map[string]ItemRecord),
		users:         make(map[string]UserRecord),
		popularItems:  make(map[string][]ItemRecord),
	}
	
	// 初始化一些测试数据
	ds.initializeTestData()
	
	return ds
}

// initializeTestData 初始化测试数据
func (m *MemoryDataSource) initializeTestData() {
	// 初始化物品数据
	items := []ItemRecord{
		{
			ItemID:      "item_001",
			Category:    "technology",
			Title:       "iPhone 15 Pro",
			Description: "最新款iPhone，配备A17 Pro芯片",
			Features: map[string]interface{}{
				"brand": "Apple",
				"price": 9999.0,
				"rating": 4.8,
			},
			Metadata: map[string]interface{}{
				"release_date": "2023-09-12",
				"storage":      "256GB",
			},
			Popularity: 0.95,
		},
		{
			ItemID:      "item_002",
			Category:    "sports",
			Title:       "Nike Air Max",
			Description: "经典运动鞋，舒适透气",
			Features: map[string]interface{}{
				"brand": "Nike",
				"price": 899.0,
				"rating": 4.5,
			},
			Metadata: map[string]interface{}{
				"color": "black/white",
				"size":  "42",
			},
			Popularity: 0.87,
		},
		{
			ItemID:      "item_003",
			Category:    "technology",
			Title:       "MacBook Pro M3",
			Description: "专业级笔记本电脑",
			Features: map[string]interface{}{
				"brand": "Apple",
				"price": 15999.0,
				"rating": 4.9,
			},
			Metadata: map[string]interface{}{
				"memory": "16GB",
				"storage": "512GB SSD",
			},
			Popularity: 0.82,
		},
	}
	
	for _, item := range items {
		m.items[item.ItemID] = item
	}
	
	// 初始化用户数据
	users := []UserRecord{
		{
			UserID: "user_123",
			Demographics: map[string]interface{}{
				"age":    25,
				"gender": "male",
				"city":   "Beijing",
			},
			Preferences: map[string]interface{}{
				"categories": []string{"technology", "sports"},
				"brands":     []string{"Apple", "Nike"},
			},
			BehaviorStats: map[string]interface{}{
				"total_behaviors": 156,
				"active_days":     30,
			},
		},
		{
			UserID: "user_456",
			Demographics: map[string]interface{}{
				"age":    30,
				"gender": "female",
				"city":   "Shanghai",
			},
			Preferences: map[string]interface{}{
				"categories": []string{"fashion", "beauty"},
				"brands":     []string{"Zara", "L'Oreal"},
			},
			BehaviorStats: map[string]interface{}{
				"total_behaviors": 203,
				"active_days":     45,
			},
		},
	}
	
	for _, user := range users {
		m.users[user.UserID] = user
	}
	
	// 初始化用户行为数据
	behaviors := []UserBehaviorRecord{
		{
			UserID:    "user_123",
			ItemID:    "item_001",
			Behavior:  "view",
			Value:     1.0,
			Timestamp: time.Now().Add(-24 * time.Hour),
			Context: map[string]interface{}{
				"device": "mobile",
				"source": "search",
			},
		},
		{
			UserID:    "user_123",
			ItemID:    "item_002",
			Behavior:  "click",
			Value:     1.0,
			Timestamp: time.Now().Add(-12 * time.Hour),
			Context: map[string]interface{}{
				"device": "desktop",
				"source": "recommendation",
			},
		},
		{
			UserID:    "user_123",
			ItemID:    "item_003",
			Behavior:  "purchase",
			Value:     15999.0,
			Timestamp: time.Now().Add(-1 * time.Hour),
			Context: map[string]interface{}{
				"device": "desktop",
				"source": "direct",
			},
		},
	}
	
	for _, behavior := range behaviors {
		key := behavior.UserID
		if m.userBehaviors[key] == nil {
			m.userBehaviors[key] = make([]UserBehaviorRecord, 0)
		}
		m.userBehaviors[key] = append(m.userBehaviors[key], behavior)
	}
	
	// 初始化热门物品
	m.updatePopularItems()
}

// GetUserBehaviorData 获取用户行为数据
func (m *MemoryDataSource) GetUserBehaviorData(ctx context.Context, userID string, startTime, endTime time.Time) ([]UserBehaviorRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	behaviors, exists := m.userBehaviors[userID]
	if !exists {
		return []UserBehaviorRecord{}, nil
	}
	
	// 过滤时间范围
	var result []UserBehaviorRecord
	for _, behavior := range behaviors {
		if behavior.Timestamp.After(startTime) && behavior.Timestamp.Before(endTime) {
			result = append(result, behavior)
		}
	}
	
	m.log.WithFields(logrus.Fields{
		"user_id":   userID,
		"count":     len(result),
		"time_range": fmt.Sprintf("%s-%s", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339)),
	}).Info("获取用户行为数据成功")
	
	return result, nil
}

// GetItemData 获取物品数据
func (m *MemoryDataSource) GetItemData(ctx context.Context, itemIDs []string) ([]ItemRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var result []ItemRecord
	for _, itemID := range itemIDs {
		if item, exists := m.items[itemID]; exists {
			result = append(result, item)
		}
	}
	
	m.log.WithFields(logrus.Fields{
		"requested_count": len(itemIDs),
		"returned_count":  len(result),
	}).Info("获取物品数据成功")
	
	return result, nil
}

// GetUserData 获取用户数据
func (m *MemoryDataSource) GetUserData(ctx context.Context, userID string) (*UserRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	user, exists := m.users[userID]
	if !exists {
		return nil, fmt.Errorf("用户不存在: %s", userID)
	}
	
	m.log.WithFields(logrus.Fields{
		"user_id": userID,
	}).Info("获取用户数据成功")
	
	return &user, nil
}

// GetPopularItems 获取热门物品
func (m *MemoryDataSource) GetPopularItems(ctx context.Context, category string, limit int) ([]ItemRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if category == "" {
		// 获取所有类别的热门物品
		var allItems []ItemRecord
		for _, item := range m.items {
			allItems = append(allItems, item)
		}
		
		// 按热度排序
		m.sortItemsByPopularity(allItems)
		
		if limit > 0 && limit < len(allItems) {
			allItems = allItems[:limit]
		}
		
		return allItems, nil
	}
	
	// 获取指定类别的热门物品
	items, exists := m.popularItems[category]
	if !exists {
		return []ItemRecord{}, nil
	}
	
	if limit > 0 && limit < len(items) {
		items = items[:limit]
	}
	
	m.log.WithFields(logrus.Fields{
		"category": category,
		"limit":    limit,
		"count":    len(items),
	}).Info("获取热门物品成功")
	
	return items, nil
}

// GetSimilarUsers 获取相似用户
func (m *MemoryDataSource) GetSimilarUsers(ctx context.Context, userID string, limit int) ([]SimilarUserRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// 简单的相似用户计算（基于用户行为数量）
	targetUser, exists := m.users[userID]
	if !exists {
		return []SimilarUserRecord{}, fmt.Errorf("用户不存在: %s", userID)
	}
	
	targetBehaviors := len(m.userBehaviors[userID])
	
	var similarUsers []SimilarUserRecord
	for uid, user := range m.users {
		if uid == userID {
			continue
		}
		
		behaviors := len(m.userBehaviors[uid])
		if behaviors > 0 {
			similarity := 1.0 - float64(abs(targetBehaviors-behaviors))/float64(max(targetBehaviors, behaviors))
			similarUsers = append(similarUsers, SimilarUserRecord{
				UserID:     uid,
				Similarity: similarity,
			})
		}
	}
	
	// 按相似度排序
	m.sortSimilarUsers(similarUsers)
	
	if limit > 0 && limit < len(similarUsers) {
		similarUsers = similarUsers[:limit]
	}
	
	m.log.WithFields(logrus.Fields{
		"user_id": userID,
		"limit":   limit,
		"count":   len(similarUsers),
	}).Info("获取相似用户成功")
	
	return similarUsers, nil
}

// HealthCheck 健康检查
func (m *MemoryDataSource) HealthCheck(ctx context.Context) error {
	// 内存数据源总是健康的
	return nil
}

// GetName 获取数据源名称
func (m *MemoryDataSource) GetName() string {
	return m.name
}

// Close 关闭数据源
func (m *MemoryDataSource) Close() error {
	m.log.WithField("name", m.name).Info("关闭内存数据源")
	return nil
}

// 更新热门物品
func (m *MemoryDataSource) updatePopularItems() {
	categoryItems := make(map[string][]ItemRecord)
	
	for _, item := range m.items {
		categoryItems[item.Category] = append(categoryItems[item.Category], item)
	}
	
	// 对每个类别的物品按热度排序
	for category, items := range categoryItems {
		m.sortItemsByPopularity(items)
		m.popularItems[category] = items
	}
}

// 按热度排序物品
func (m *MemoryDataSource) sortItemsByPopularity(items []ItemRecord) {
	// 简单的冒泡排序
	for i := 0; i < len(items)-1; i++ {
		for j := 0; j < len(items)-i-1; j++ {
			if items[j].Popularity < items[j+1].Popularity {
				items[j], items[j+1] = items[j+1], items[j]
			}
		}
	}
}

// 按相似度排序用户
func (m *MemoryDataSource) sortSimilarUsers(users []SimilarUserRecord) {
	for i := 0; i < len(users)-1; i++ {
		for j := 0; j < len(users)-i-1; j++ {
			if users[j].Similarity < users[j+1].Similarity {
				users[j], users[j+1] = users[j+1], users[j]
			}
		}
	}
}

// 辅助函数
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}