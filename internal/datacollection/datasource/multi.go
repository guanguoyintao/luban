// Package datasource 多数据源适配器（多路召回）
package datasource

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"github.com/sirupsen/logrus"
)

// MultiDataSource 多数据源适配器
type MultiDataSource struct {
	sources []DataSource
	log     *logrus.Logger
	mu      sync.RWMutex
}

// NewMultiDataSource 创建多数据源适配器
func NewMultiDataSource(sources []DataSource, log *logrus.Logger) *MultiDataSource {
	if log == nil {
		log = logrus.New()
	}
	
	return &MultiDataSource{
		sources: sources,
		log:     log,
	}
}

// MultiRecall 多路召回
type MultiRecall struct {
	results map[string]RecallResult
	mu      sync.RWMutex
}

// NewMultiRecall 创建多路召回
func NewMultiRecall() *MultiRecall {
	return &MultiRecall{
		results: make(map[string]RecallResult),
	}
}

// AddResult 添加召回结果
func (m *MultiRecall) AddResult(source string, result RecallResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.results[source] = result
}

// GetResults 获取所有召回结果
func (m *MultiRecall) GetResults() map[string]RecallResult {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	results := make(map[string]RecallResult)
	for k, v := range m.results {
		results[k] = v
	}
	return results
}

// MergeResults 合并召回结果
func (m *MultiRecall) MergeResults() []ItemRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	itemMap := make(map[string]ItemRecord)
	
	for source, result := range m.results {
		for _, item := range result.Items {
			if existing, exists := itemMap[item.ItemID]; exists {
				// 如果物品已存在，合并分数（取最大值）
				if result.Score > existing.Popularity {
					item.Popularity = result.Score
				}
				// 添加数据源信息
				if item.Metadata == nil {
					item.Metadata = make(map[string]interface{})
				}
				item.Metadata["sources"] = append(item.Metadata["sources"].([]string), source)
			} else {
				item.Popularity = result.Score
				if item.Metadata == nil {
					item.Metadata = make(map[string]interface{})
				}
				item.Metadata["sources"] = []string{source}
				itemMap[item.ItemID] = item
			}
		}
	}
	
	// 转换为切片
	result := make([]ItemRecord, 0, len(itemMap))
	for _, item := range itemMap {
		result = append(result, item)
	}
	
	return result
}

// ParallelRecall 并行多路召回
func (m *MultiDataSource) ParallelRecall(ctx context.Context, userID string, recallTypes []string) (*MultiRecall, error) {
	m.log.WithFields(logrus.Fields{
		"user_id":      userID,
		"recall_types": recallTypes,
		"source_count": len(m.sources),
	}).Info("开始并行多路召回")
	
	multiRecall := NewMultiRecall()
	
	// 创建错误通道和完成信号
	errChan := make(chan error, len(m.sources))
	doneChan := make(chan bool, len(m.sources))
	
	// 并行执行各路召回
	var wg sync.WaitGroup
	for i, source := range m.sources {
		wg.Add(1)
		go func(idx int, src DataSource) {
			defer wg.Done()
			
			result, err := m.executeRecall(ctx, src, userID, recallTypes)
			if err != nil {
				m.log.WithError(err).WithField("source", src.GetName()).Error("召回失败")
				errChan <- err
				return
			}
			
			if result != nil {
				multiRecall.AddResult(src.GetName(), *result)
			}
			doneChan <- true
		}(i, source)
	}
	
	// 等待所有召回完成
	go func() {
		wg.Wait()
		close(errChan)
		close(doneChan)
	}()
	
	// 收集结果
	successCount := 0
	errorCount := 0
	
	for done := range doneChan {
		if done {
			successCount++
		}
	}
	
	for err := range errChan {
		if err != nil {
			errorCount++
		}
	}
	
	m.log.WithFields(logrus.Fields{
		"success_count": successCount,
		"error_count":   errorCount,
		"total_results": len(multiRecall.GetResults()),
	}).Info("并行多路召回完成")
	
	return multiRecall, nil
}

// executeRecall 执行单路召回
func (m *MultiDataSource) executeRecall(ctx context.Context, source DataSource, userID string, recallTypes []string) (*RecallResult, error) {
	sourceName := source.GetName()
	
	// 根据召回类型执行不同的召回策略
	for _, recallType := range recallTypes {
		switch recallType {
		case "popular":
			return m.recallPopularItems(ctx, source, userID)
		case "similar_users":
			return m.recallSimilarUsersItems(ctx, source, userID)
		case "recent_behavior":
			return m.recallRecentBehaviorItems(ctx, source, userID)
		case "category_preference":
			return m.recallCategoryPreferenceItems(ctx, source, userID)
		default:
			m.log.WithFields(logrus.Fields{
				"source":      sourceName,
				"recall_type": recallType,
			}).Warn("不支持的召回类型")
		}
	}
	
	return nil, fmt.Errorf("没有可用的召回策略")
}

// recallPopularItems 热门物品召回
func (m *MultiDataSource) recallPopularItems(ctx context.Context, source DataSource, userID string) (*RecallResult, error) {
	// 获取用户数据以了解用户偏好
	userData, err := source.GetUserData(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户数据失败: %w", err)
	}
	
	// 获取用户偏好的类别
	var categories []string
	if prefs, ok := userData.Preferences["categories"].([]string); ok {
		categories = prefs
	}
	
	var items []ItemRecord
	if len(categories) > 0 {
		// 根据用户偏好类别召回热门物品
		for _, category := range categories {
			popularItems, err := source.GetPopularItems(ctx, category, 10)
			if err != nil {
				m.log.WithError(err).WithField("category", category).Error("获取热门物品失败")
				continue
			}
			items = append(items, popularItems...)
		}
	} else {
		// 如果没有偏好类别，获取所有热门物品
		allPopularItems, err := source.GetPopularItems(ctx, "", 20)
		if err != nil {
			return nil, fmt.Errorf("获取热门物品失败: %w", err)
		}
		items = allPopularItems
	}
	
	return &RecallResult{
		Items:    items,
		Score:    0.8, // 热门物品的基础分数
		Source:   "popular_items",
		Metadata: map[string]interface{}{
			"strategy":  "popular",
			"timestamp": time.Now(),
		},
	}, nil
}

// recallSimilarUsersItems 相似用户召回
func (m *MultiDataSource) recallSimilarUsersItems(ctx context.Context, source DataSource, userID string) (*RecallResult, error) {
	// 获取相似用户
	similarUsers, err := source.GetSimilarUsers(ctx, userID, 10)
	if err != nil {
		return nil, fmt.Errorf("获取相似用户失败: %w", err)
	}
	
	var items []ItemRecord
	userItemMap := make(map[string]bool) // 避免重复物品
	
	for _, similarUser := range similarUsers {
		// 获取相似用户的行为数据
		behaviors, err := source.GetUserBehaviorData(ctx, similarUser.UserID, time.Now().Add(-30*24*time.Hour), time.Now())
		if err != nil {
			m.log.WithError(err).WithField("similar_user", similarUser.UserID).Error("获取相似用户行为数据失败")
			continue
		}
		
		// 获取相似用户交互过的物品
		itemIDs := make([]string, 0)
		for _, behavior := range behaviors {
			if !userItemMap[behavior.ItemID] {
				itemIDs = append(itemIDs, behavior.ItemID)
				userItemMap[behavior.ItemID] = true
			}
		}
		
		if len(itemIDs) > 0 {
			userItems, err := source.GetItemData(ctx, itemIDs)
			if err != nil {
				m.log.WithError(err).WithField("similar_user", similarUser.UserID).Error("获取相似用户物品数据失败")
				continue
			}
			items = append(items, userItems...)
		}
	}
	
	return &RecallResult{
		Items:    items,
		Score:    0.7, // 相似用户召回的基础分数
		Source:   "similar_users",
		Metadata: map[string]interface{}{
			"strategy":       "collaborative",
			"similar_users":  len(similarUsers),
			"timestamp":      time.Now(),
		},
	}, nil
}

// recallRecentBehaviorItems 近期行为召回
func (m *MultiDataSource) recallRecentBehaviorItems(ctx context.Context, source DataSource, userID string) (*RecallResult, error) {
	// 获取用户近期行为数据
	behaviors, err := source.GetUserBehaviorData(ctx, userID, time.Now().Add(-7*24*time.Hour), time.Now())
	if err != nil {
		return nil, fmt.Errorf("获取用户行为数据失败: %w", err)
	}
	
	if len(behaviors) == 0 {
		return &RecallResult{
			Items:    []ItemRecord{},
			Score:    0.0,
			Source:   "recent_behavior",
			Metadata: map[string]interface{}{
				"strategy":  "recent_behavior",
				"reason":    "no_recent_behavior",
				"timestamp": time.Now(),
			},
		}, nil
	}
	
	// 获取用户近期交互过的物品
	itemIDs := make([]string, 0)
	for _, behavior := range behaviors {
		itemIDs = append(itemIDs, behavior.ItemID)
	}
	
	items, err := source.GetItemData(ctx, itemIDs)
	if err != nil {
		return nil, fmt.Errorf("获取物品数据失败: %w", err)
	}
	
	return &RecallResult{
		Items:    items,
		Score:    0.6, // 近期行为召回的基础分数
		Source:   "recent_behavior",
		Metadata: map[string]interface{}{
			"strategy":     "recent_behavior",
			"behavior_count": len(behaviors),
			"time_range":   "7d",
			"timestamp":    time.Now(),
		},
	}, nil
}

// recallCategoryPreferenceItems 类别偏好召回
func (m *MultiDataSource) recallCategoryPreferenceItems(ctx context.Context, source DataSource, userID string) (*RecallResult, error) {
	// 获取用户数据
	userData, err := source.GetUserData(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户数据失败: %w", err)
	}
	
	// 获取用户偏好的类别
	var categories []string
	if prefs, ok := userData.Preferences["categories"].([]string); ok {
		categories = prefs
	}
	
	if len(categories) == 0 {
		return &RecallResult{
			Items:    []ItemRecord{},
			Score:    0.0,
			Source:   "category_preference",
			Metadata: map[string]interface{}{
				"strategy": "category_preference",
				"reason":   "no_category_preference",
				"timestamp": time.Now(),
			},
		}, nil
	}
	
	var items []ItemRecord
	for _, category := range categories {
		categoryItems, err := source.GetPopularItems(ctx, category, 5)
		if err != nil {
			m.log.WithError(err).WithField("category", category).Error("获取类别偏好物品失败")
			continue
		}
		items = append(items, categoryItems...)
	}
	
	return &RecallResult{
		Items:    items,
		Score:    0.75, // 类别偏好召回的基础分数
		Source:   "category_preference",
		Metadata: map[string]interface{}{
			"strategy":   "category_preference",
			"categories": categories,
			"timestamp":  time.Now(),
		},
	}, nil
}

// GetName 获取数据源名称
func (m *MultiDataSource) GetName() string {
	return "multi_data_source"
}

// HealthCheck 健康检查
func (m *MultiDataSource) HealthCheck(ctx context.Context) error {
	for _, source := range m.sources {
		if err := source.HealthCheck(ctx); err != nil {
			return fmt.Errorf("数据源 %s 健康检查失败: %w", source.GetName(), err)
		}
	}
	return nil
}

// Close 关闭所有数据源
func (m *MultiDataSource) Close() error {
	var lastErr error
	for _, source := range m.sources {
		if err := source.Close(); err != nil {
			m.log.WithError(err).WithField("source", source.GetName()).Error("关闭数据源失败")
			lastErr = err
		}
	}
	return lastErr
}