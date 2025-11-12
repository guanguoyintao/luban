package datacollection

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// 内存数据采集器实现
type MemoryDataCollector struct {
	mu              sync.RWMutex
	userBehaviors   map[string][]UserBehavior // 用户行为数据，按用户ID分组
	itemsData       map[string]ItemData       // 物品数据
	usersData       map[string]UserData       // 用户数据
	log             *logrus.Logger
	maxHistorySize  int                       // 每个用户最大历史记录数
}

// 创建新的内存数据采集器
func NewMemoryDataCollector(log *logrus.Logger) *MemoryDataCollector {
	if log == nil {
		log = logrus.New()
	}
	
	return &MemoryDataCollector{
		userBehaviors:  make(map[string][]UserBehavior),
		itemsData:      make(map[string]ItemData),
		usersData:      make(map[string]UserData),
		log:            log,
		maxHistorySize: 1000, // 默认每个用户最多保存1000条历史记录
	}
}

// 收集用户行为数据
func (m *MemoryDataCollector) CollectUserBehavior(ctx context.Context, behavior UserBehavior) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 如果行为ID为空，生成新的UUID
	if behavior.UserID == "" {
		return &DataCollectionError{Message: "用户ID不能为空"}
	}
	
	// 设置时间戳
	if behavior.Timestamp.IsZero() {
		behavior.Timestamp = time.Now()
	}
	
	// 添加到用户行为历史
	history := m.userBehaviors[behavior.UserID]
	history = append(history, behavior)
	
	// 限制历史记录数量
	if len(history) > m.maxHistorySize {
		history = history[len(history)-m.maxHistorySize:]
	}
	
	m.userBehaviors[behavior.UserID] = history
	
	m.log.WithFields(logrus.Fields{
		"user_id":  behavior.UserID,
		"item_id":  behavior.ItemID,
		"behavior": behavior.Behavior,
		"value":    behavior.Value,
	}).Info("收集用户行为数据成功")
	
	return nil
}

// 批量收集用户行为数据
func (m *MemoryDataCollector) CollectUserBehaviors(ctx context.Context, behaviors []UserBehavior) error {
	for _, behavior := range behaviors {
		if err := m.CollectUserBehavior(ctx, behavior); err != nil {
			return err
		}
	}
	return nil
}

// 收集物品数据
func (m *MemoryDataCollector) CollectItemData(ctx context.Context, item ItemData) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if item.ItemID == "" {
		item.ItemID = uuid.New().String()
	}
	
	m.itemsData[item.ItemID] = item
	
	m.log.WithFields(logrus.Fields{
		"item_id":  item.ItemID,
		"category": item.Category,
		"title":    item.Title,
	}).Info("收集物品数据成功")
	
	return nil
}

// 批量收集物品数据
func (m *MemoryDataCollector) CollectItemsData(ctx context.Context, items []ItemData) error {
	for _, item := range items {
		if err := m.CollectItemData(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

// 收集用户数据
func (m *MemoryDataCollector) CollectUserData(ctx context.Context, user UserData) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if user.UserID == "" {
		user.UserID = uuid.New().String()
	}
	
	m.usersData[user.UserID] = user
	
	m.log.WithFields(logrus.Fields{
		"user_id": user.UserID,
	}).Info("收集用户数据成功")
	
	return nil
}

// 批量收集用户数据
func (m *MemoryDataCollector) CollectUsersData(ctx context.Context, users []UserData) error {
	for _, user := range users {
		if err := m.CollectUserData(ctx, user); err != nil {
			return err
		}
	}
	return nil
}

// 获取用户行为历史
func (m *MemoryDataCollector) GetUserBehaviorHistory(ctx context.Context, userID string, limit int) ([]UserBehavior, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	history, exists := m.userBehaviors[userID]
	if !exists {
		return []UserBehavior{}, nil
	}
	
	// 按时间戳排序（最新的在前）
	sortedHistory := make([]UserBehavior, len(history))
	copy(sortedHistory, history)
	
	// 简单的冒泡排序，可以优化为更快的排序算法
	for i := 0; i < len(sortedHistory)-1; i++ {
		for j := 0; j < len(sortedHistory)-i-1; j++ {
			if sortedHistory[j].Timestamp.Before(sortedHistory[j+1].Timestamp) {
				sortedHistory[j], sortedHistory[j+1] = sortedHistory[j+1], sortedHistory[j]
			}
		}
	}
	
	// 限制返回数量
	if limit > 0 && limit < len(sortedHistory) {
		return sortedHistory[:limit], nil
	}
	
	return sortedHistory, nil
}

// 获取物品数据
func (m *MemoryDataCollector) GetItemData(ctx context.Context, itemID string) (*ItemData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	item, exists := m.itemsData[itemID]
	if !exists {
		return nil, &DataCollectionError{Message: "物品不存在: " + itemID}
	}
	
	return &item, nil
}

// 获取用户数据
func (m *MemoryDataCollector) GetUserData(ctx context.Context, userID string) (*UserData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	user, exists := m.usersData[userID]
	if !exists {
		return nil, &DataCollectionError{Message: "用户不存在: " + userID}
	}
	
	return &user, nil
}

// 关闭采集器
func (m *MemoryDataCollector) Close() error {
	m.log.Info("关闭内存数据采集器")
	return nil
}

// 设置最大历史记录数
func (m *MemoryDataCollector) SetMaxHistorySize(size int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maxHistorySize = size
}

// 数据采集错误
type DataCollectionError struct {
	Message string
}

func (e *DataCollectionError) Error() string {
	return "数据采集错误: " + e.Message
}