package dataprocessing

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// 内存数据处理器实现
type MemoryDataProcessor struct {
	mu              sync.RWMutex
	log             *logrus.Logger
	normalizer      *DataNormalizer
	featureExtractor *FeatureExtractor
	qualityChecker  *DataQualityChecker
}

// 创建新的内存数据处理器
func NewMemoryDataProcessor(log *logrus.Logger) *MemoryDataProcessor {
	if log == nil {
		log = logrus.New()
	}
	
	return &MemoryDataProcessor{
		log:              log,
		normalizer:       NewDataNormalizer(),
		featureExtractor: NewFeatureExtractor(),
		qualityChecker:   NewDataQualityChecker(),
	}
}

// 清洗用户行为数据
func (m *MemoryDataProcessor) CleanUserBehaviorData(ctx context.Context, rawData interface{}) (*ProcessedUserBehavior, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	behavior, ok := rawData.(UserBehavior)
	if !ok {
		return nil, &DataProcessingError{Message: "无效的用户行为数据类型"}
	}
	
	// 数据验证
	if err := m.validateUserBehavior(behavior); err != nil {
		return nil, err
	}
	
	// 归一化行为值
	normalizedValue := m.normalizeBehaviorValue(behavior.Behavior, behavior.Value)
	
	// 计算行为权重
	weight := m.calculateBehaviorWeight(behavior.Behavior, behavior.Value)
	
	// 提取特征
	features := m.extractUserBehaviorFeatures(behavior)
	
	processed := &ProcessedUserBehavior{
		UserID:          behavior.UserID,
		ItemID:          behavior.ItemID,
		Behavior:        string(behavior.Behavior),
		NormalizedValue: normalizedValue,
		Timestamp:       behavior.Timestamp,
		Weight:          weight,
		Features:        features,
	}
	
	m.log.WithFields(logrus.Fields{
		"user_id": behavior.UserID,
		"item_id": behavior.ItemID,
		"behavior": behavior.Behavior,
	}).Info("用户行为数据清洗成功")
	
	return processed, nil
}

// 批量清洗用户行为数据
func (m *MemoryDataProcessor) CleanUserBehaviorDataBatch(ctx context.Context, rawData []interface{}) ([]ProcessedUserBehavior, error) {
	results := make([]ProcessedUserBehavior, 0, len(rawData))
	
	for _, data := range rawData {
		processed, err := m.CleanUserBehaviorData(ctx, data)
		if err != nil {
			m.log.WithError(err).Error("清洗用户行为数据失败")
			continue
		}
		results = append(results, *processed)
	}
	
	return results, nil
}

// 清洗物品数据
func (m *MemoryDataProcessor) CleanItemData(ctx context.Context, rawData interface{}) (*ProcessedItemData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	item, ok := rawData.(ItemData)
	if !ok {
		return nil, &DataProcessingError{Message: "无效的物品数据类型"}
	}
	
	// 数据验证
	if err := m.validateItemData(item); err != nil {
		return nil, err
	}
	
	// 提取特征向量
	features, err := m.featureExtractor.ExtractItemFeatures(item)
	if err != nil {
		return nil, err
	}
	
	// 计算数据质量
	quality := m.qualityChecker.CheckItemQuality(item)
	
	processed := &ProcessedItemData{
		ItemID:   item.ItemID,
		Category: item.Category,
		Features: features,
		Metadata: item.Metadata,
		Quality:  quality,
	}
	
	m.log.WithFields(logrus.Fields{
		"item_id": item.ItemID,
		"category": item.Category,
		"quality": quality,
	}).Info("物品数据清洗成功")
	
	return processed, nil
}

// 批量清洗物品数据
func (m *MemoryDataProcessor) CleanItemDataBatch(ctx context.Context, rawData []interface{}) ([]ProcessedItemData, error) {
	results := make([]ProcessedItemData, 0, len(rawData))
	
	for _, data := range rawData {
		processed, err := m.CleanItemData(ctx, data)
		if err != nil {
			m.log.WithError(err).Error("清洗物品数据失败")
			continue
		}
		results = append(results, *processed)
	}
	
	return results, nil
}

// 清洗用户数据
func (m *MemoryDataProcessor) CleanUserData(ctx context.Context, rawData interface{}) (*ProcessedUserData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	user, ok := rawData.(UserData)
	if !ok {
		return nil, &DataProcessingError{Message: "无效的用户数据类型"}
	}
	
	// 数据验证
	if err := m.validateUserData(user); err != nil {
		return nil, err
	}
	
	// 提取特征向量
	features, err := m.featureExtractor.ExtractUserFeatures(user)
	if err != nil {
		return nil, err
	}
	
	// 计算数据质量
	quality := m.qualityChecker.CheckUserQuality(user)
	
	processed := &ProcessedUserData{
		UserID:      user.UserID,
		Features:    features,
		Preferences: user.Preferences,
		Quality:     quality,
	}
	
	m.log.WithFields(logrus.Fields{
		"user_id": user.UserID,
		"quality": quality,
	}).Info("用户数据清洗成功")
	
	return processed, nil
}

// 批量清洗用户数据
func (m *MemoryDataProcessor) CleanUserDataBatch(ctx context.Context, rawData []interface{}) ([]ProcessedUserData, error) {
	results := make([]ProcessedUserData, 0, len(rawData))
	
	for _, data := range rawData {
		processed, err := m.CleanUserData(ctx, data)
		if err != nil {
			m.log.WithError(err).Error("清洗用户数据失败")
			continue
		}
		results = append(results, *processed)
	}
	
	return results, nil
}

// 数据归一化处理
func (m *MemoryDataProcessor) NormalizeData(ctx context.Context, data []float64) ([]float64, error) {
	return m.normalizer.Normalize(data), nil
}

// 特征提取
func (m *MemoryDataProcessor) ExtractFeatures(ctx context.Context, data interface{}) ([]float64, error) {
	return m.featureExtractor.ExtractGenericFeatures(data)
}

// 计算数据质量指标
func (m *MemoryDataProcessor) CalculateDataQuality(ctx context.Context, data interface{}) (*DataQualityMetrics, error) {
	return m.qualityChecker.CalculateQualityMetrics(data), nil
}

// 数据去重
func (m *MemoryDataProcessor) RemoveDuplicates(ctx context.Context, data []interface{}) ([]interface{}, error) {
	if len(data) == 0 {
		return data, nil
	}
	
	seen := make(map[string]bool)
	result := make([]interface{}, 0)
	
	for _, item := range data {
		key := m.generateItemKey(item)
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}
	
	return result, nil
}

// 处理缺失值
func (m *MemoryDataProcessor) HandleMissingValues(ctx context.Context, data interface{}) (interface{}, error) {
	// 简单的缺失值处理：使用默认值填充
	switch v := data.(type) {
	case UserBehavior:
		return m.handleMissingUserBehaviorValues(v), nil
	case ItemData:
		return m.handleMissingItemDataValues(v), nil
	case UserData:
		return m.handleMissingUserDataValues(v), nil
	default:
		return nil, &DataProcessingError{Message: "不支持的数据类型"}
	}
}

// 关闭处理器
func (m *MemoryDataProcessor) Close() error {
	m.log.Info("关闭内存数据处理器")
	return nil
}

// 验证用户行为数据
func (m *MemoryDataProcessor) validateUserBehavior(behavior UserBehavior) error {
	if behavior.UserID == "" {
		return &DataProcessingError{Message: "用户ID不能为空"}
	}
	if behavior.ItemID == "" {
		return &DataProcessingError{Message: "物品ID不能为空"}
	}
	if behavior.Value < 0 {
		return &DataProcessingError{Message: "行为值不能为负数"}
	}
	return nil
}

// 验证物品数据
func (m *MemoryDataProcessor) validateItemData(item ItemData) error {
	if item.ItemID == "" {
		return &DataProcessingError{Message: "物品ID不能为空"}
	}
	return nil
}

// 验证用户数据
func (m *MemoryDataProcessor) validateUserData(user UserData) error {
	if user.UserID == "" {
		return &DataProcessingError{Message: "用户ID不能为空"}
	}
	return nil
}

// 归一化行为值
func (m *MemoryDataProcessor) normalizeBehaviorValue(behavior string, value float64) float64 {
	// 根据不同行为类型进行归一化
	switch behavior {
	case "rating":
		// 评分值归一化到0-1范围
		return value / 5.0
	case "click", "view":
		// 点击和浏览行为，简单归一化
		return math.Min(value/10.0, 1.0)
	case "purchase":
		// 购买行为，给予较高权重
		return 1.0
	default:
		return value
	}
}

// 计算行为权重
func (m *MemoryDataProcessor) calculateBehaviorWeight(behavior string, value float64) float64 {
	// 不同行为的权重
	weights := map[string]float64{
		"purchase": 1.0,
		"rating":   0.8,
		"favorite": 0.7,
		"share":    0.6,
		"click":    0.4,
		"view":     0.2,
	}
	
	if weight, exists := weights[behavior]; exists {
		return weight
	}
	return 0.1
}

// 提取用户行为特征
func (m *MemoryDataProcessor) extractUserBehaviorFeatures(behavior UserBehavior) map[string]interface{} {
	features := make(map[string]interface{})
	
	// 时间特征
	hour := behavior.Timestamp.Hour()
	features["hour_of_day"] = float64(hour) / 24.0
	features["is_weekend"] = behavior.Timestamp.Weekday() == time.Saturday || behavior.Timestamp.Weekday() == time.Sunday
	
	// 行为特征
	features["behavior_type"] = string(behavior.Behavior)
	features["behavior_value"] = behavior.Value
	
	return features
}

// 生成项目键用于去重
func (m *MemoryDataProcessor) generateItemKey(item interface{}) string {
	return fmt.Sprintf("%v", item)
}

// 处理缺失的用户行为值
func (m *MemoryDataProcessor) handleMissingUserBehaviorValues(behavior UserBehavior) UserBehavior {
	if behavior.Context == nil {
		behavior.Context = make(map[string]interface{})
	}
	return behavior
}

// 处理缺失的物品数据值
func (m *MemoryDataProcessor) handleMissingItemDataValues(item ItemData) ItemData {
	if item.Features == nil {
		item.Features = make(map[string]interface{})
	}
	if item.Metadata == nil {
		item.Metadata = make(map[string]interface{})
	}
	return item
}

// 处理缺失的用户数据值
func (m *MemoryDataProcessor) handleMissingUserDataValues(user UserData) UserData {
	if user.Demographics == nil {
		user.Demographics = make(map[string]interface{})
	}
	if user.Preferences == nil {
		user.Preferences = make(map[string]interface{})
	}
	if user.Metadata == nil {
		user.Metadata = make(map[string]interface{})
	}
	return user
}

// 数据归一化器
type DataNormalizer struct{}

func NewDataNormalizer() *DataNormalizer {
	return &DataNormalizer{}
}

func (d *DataNormalizer) Normalize(data []float64) []float64 {
	if len(data) == 0 {
		return data
	}
	
	// 找到最小值和最大值
	min, max := data[0], data[0]
	for _, value := range data {
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}
	
	// 如果所有值都相同，返回全1数组
	if max == min {
		result := make([]float64, len(data))
		for i := range result {
			result[i] = 1.0
		}
		return result
	}
	
	// 归一化到0-1范围
	result := make([]float64, len(data))
	for i, value := range data {
		result[i] = (value - min) / (max - min)
	}
	
	return result
}

// 特征提取器
type FeatureExtractor struct{}

func NewFeatureExtractor() *FeatureExtractor {
	return &FeatureExtractor{}
}

func (f *FeatureExtractor) ExtractItemFeatures(item ItemData) ([]float64, error) {
	features := make([]float64, 0)
	
	// 类别特征（简单的哈希编码）
	categoryHash := float64(hashString(item.Category))
	features = append(features, float64(categoryHash%1000)/1000.0)
	
	// 标题长度特征
	features = append(features, float64(len(item.Title))/100.0)
	
	// 描述长度特征
	features = append(features, float64(len(item.Description))/500.0)
	
	return features, nil
}

func (f *FeatureExtractor) ExtractUserFeatures(user UserData) ([]float64, error) {
	features := make([]float64, 0)
	
	// 偏好特征数量
	features = append(features, float64(len(user.Preferences)))
	
	// 人口统计学特征
	features = append(features, float64(len(user.Demographics)))
	
	return features, nil
}

func (f *FeatureExtractor) ExtractGenericFeatures(data interface{}) ([]float64, error) {
	return []float64{1.0}, nil // 简单的默认特征
}

// 数据质量检查器
type DataQualityChecker struct{}

func NewDataQualityChecker() *DataQualityChecker {
	return &DataQualityChecker{}
}

func (d *DataQualityChecker) CheckItemQuality(item ItemData) float64 {
	score := 1.0
	
	// 检查必填字段
	if item.ItemID == "" {
		score -= 0.3
	}
	if item.Title == "" {
		score -= 0.2
	}
	
	return math.Max(score, 0.0)
}

func (d *DataQualityChecker) CheckUserQuality(user UserData) float64 {
	score := 1.0
	
	// 检查必填字段
	if user.UserID == "" {
		score -= 0.3
	}
	
	return math.Max(score, 0.0)
}

func (d *DataQualityChecker) CalculateQualityMetrics(data interface{}) *DataQualityMetrics {
	return &DataQualityMetrics{
		Completeness: 0.8,
		Accuracy:     0.9,
		Consistency:  0.85,
		Timeliness:   0.9,
		Validity:     0.95,
	}
}

// 辅助函数
func hashString(s string) int {
	h := 0
	for _, c := range s {
		h = h*31 + int(c)
	}
	return h
}

// 数据处理错误
type DataProcessingError struct {
	Message string
}

func (e *DataProcessingError) Error() string {
	return "数据处理错误: " + e.Message
}

// 导入数据采集层的类型
type UserBehavior struct {
	UserID    string
	ItemID    string
	Behavior  string
	Value     float64
	Timestamp time.Time
	Context   map[string]interface{}
}

type ItemData struct {
	ItemID      string
	Category    string
	Title       string
	Description string
	Features    map[string]interface{}
	Metadata    map[string]interface{}
}

type UserData struct {
	UserID        string
	Demographics map[string]interface{}
	Preferences  map[string]interface{}
	Metadata     map[string]interface{}
}