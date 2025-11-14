// Package chain 数据处理责任链具体实现
package chain

import (
	"context"
	"fmt"
	"time"
)

// ValidationProcessor 数据验证处理器
type ValidationProcessor struct {
	name string
}

// NewValidationProcessor 创建数据验证处理器
func NewValidationProcessor() *ValidationProcessor {
	return &ValidationProcessor{
		name: "validation",
	}
}

func (v *ValidationProcessor) Process(ctx context.Context, data interface{}) (interface{}, error) {
	switch d := data.(type) {
	case UserBehaviorData:
		if d.UserID == "" || d.ItemID == "" {
			return nil, fmt.Errorf("用户ID和物品ID不能为空")
		}
		if d.Value < 0 {
			return nil, fmt.Errorf("行为值不能为负数")
		}
	case ItemData:
		if d.ItemID == "" {
			return nil, fmt.Errorf("物品ID不能为空")
		}
	case UserData:
		if d.UserID == "" {
			return nil, fmt.Errorf("用户ID不能为空")
		}
	default:
		return nil, fmt.Errorf("不支持的数据类型")
	}
	return data, nil
}

func (v *ValidationProcessor) CanProcess(data interface{}) bool {
	return true // 可以处理所有类型的数据
}

func (v *ValidationProcessor) GetName() string {
	return v.name
}

// NormalizationProcessor 数据归一化处理器
type NormalizationProcessor struct {
	name string
}

// NewNormalizationProcessor 创建数据归一化处理器
func NewNormalizationProcessor() *NormalizationProcessor {
	return &NormalizationProcessor{
		name: "normalization",
	}
}

func (n *NormalizationProcessor) Process(ctx context.Context, data interface{}) (interface{}, error) {
	switch d := data.(type) {
	case UserBehaviorData:
		// 归一化行为值
		d.NormalizedValue = n.normalizeBehaviorValue(d.Behavior, d.Value)
		return d, nil
	default:
		return data, nil // 其他类型不需要归一化
	}
}

func (n *NormalizationProcessor) CanProcess(data interface{}) bool {
	_, ok := data.(UserBehaviorData)
	return ok
}

func (n *NormalizationProcessor) GetName() string {
	return n.name
}

func (n *NormalizationProcessor) normalizeBehaviorValue(behavior string, value float64) float64 {
	switch behavior {
	case "rating":
		return value / 5.0 // 评分值归一化到0-1范围
	case "click", "view":
		return min(value/10.0, 1.0) // 点击和浏览行为归一化
	case "purchase":
		return 1.0 // 购买行为给予最高权重
	default:
		return value
	}
}

// FeatureExtractionProcessor 特征提取处理器
type FeatureExtractionProcessor struct {
	name string
}

// NewFeatureExtractionProcessor 创建特征提取处理器
func NewFeatureExtractionProcessor() *FeatureExtractionProcessor {
	return &FeatureExtractionProcessor{
		name: "feature_extraction",
	}
}

func (f *FeatureExtractionProcessor) Process(ctx context.Context, data interface{}) (interface{}, error) {
	switch d := data.(type) {
	case UserBehaviorData:
		d.Features = f.extractUserBehaviorFeatures(d)
		return d, nil
	case ItemData:
		features, err := f.extractItemFeatures(d)
		if err != nil {
			return nil, err
		}
		d.Features = features
		return d, nil
	case UserData:
		features, err := f.extractUserFeatures(d)
		if err != nil {
			return nil, err
		}
		d.Features = features
		return d, nil
	default:
		return data, nil
	}
}

func (f *FeatureExtractionProcessor) CanProcess(data interface{}) bool {
	return true // 可以处理所有类型的数据
}

func (f *FeatureExtractionProcessor) GetName() string {
	return f.name
}

func (f *FeatureExtractionProcessor) extractUserBehaviorFeatures(data UserBehaviorData) map[string]interface{} {
	features := make(map[string]interface{})
	
	// 时间特征
	hour := data.Timestamp.Hour()
	features["hour_of_day"] = float64(hour) / 24.0
	features["is_weekend"] = data.Timestamp.Weekday() == time.Saturday || data.Timestamp.Weekday() == time.Sunday
	
	// 行为特征
	features["behavior_type"] = data.Behavior
	features["behavior_value"] = data.Value
	features["normalized_value"] = data.NormalizedValue
	
	return features
}

func (f *FeatureExtractionProcessor) extractItemFeatures(data ItemData) ([]float64, error) {
	features := make([]float64, 0)
	
	// 类别特征（简单的哈希编码）
	categoryHash := float64(hashString(data.Category))
	features = append(features, float64(int(categoryHash)%1000)/1000.0)
	
	// 标题长度特征
	features = append(features, float64(len(data.Title))/100.0)
	
	// 描述长度特征
	features = append(features, float64(len(data.Description))/500.0)
	
	return features, nil
}

func (f *FeatureExtractionProcessor) extractUserFeatures(data UserData) ([]float64, error) {
	features := make([]float64, 0)
	
	// 偏好特征数量
	features = append(features, float64(len(data.Preferences)))
	
	// 人口统计学特征
	features = append(features, float64(len(data.Demographics)))
	
	return features, nil
}

// QualityCheckProcessor 数据质量检查处理器
type QualityCheckProcessor struct {
	name string
}

// NewQualityCheckProcessor 创建数据质量检查处理器
func NewQualityCheckProcessor() *QualityCheckProcessor {
	return &QualityCheckProcessor{
		name: "quality_check",
	}
}

func (q *QualityCheckProcessor) Process(ctx context.Context, data interface{}) (interface{}, error) {
	switch d := data.(type) {
	case ItemData:
		quality := q.checkItemQuality(d)
		d.Quality = quality
		if quality < 0.5 {
			return nil, fmt.Errorf("物品数据质量过低: %.2f", quality)
		}
		return d, nil
	case UserData:
		quality := q.checkUserQuality(d)
		d.Quality = quality
		if quality < 0.5 {
			return nil, fmt.Errorf("用户数据质量过低: %.2f", quality)
		}
		return d, nil
	default:
		return data, nil
	}
}

func (q *QualityCheckProcessor) CanProcess(data interface{}) bool {
	return true // 可以处理所有类型的数据
}

func (q *QualityCheckProcessor) GetName() string {
	return q.name
}

func (q *QualityCheckProcessor) checkItemQuality(data ItemData) float64 {
	score := 1.0
	
	// 检查必填字段
	if data.ItemID == "" {
		score -= 0.3
	}
	if data.Title == "" {
		score -= 0.2
	}
	
	return max(score, 0.0)
}

func (q *QualityCheckProcessor) checkUserQuality(data UserData) float64 {
	score := 1.0
	
	// 检查必填字段
	if data.UserID == "" {
		score -= 0.3
	}
	
	return max(score, 0.0)
}

// 数据类型定义
type UserBehaviorData struct {
	UserID          string
	ItemID          string
	Behavior        string
	Value           float64
	NormalizedValue float64
	Timestamp       time.Time
	Features        map[string]interface{}
}

type ItemData struct {
	ItemID      string
	Category    string
	Title       string
	Description string
	Features    []float64
	Metadata    map[string]interface{}
	Quality     float64
}

type UserData struct {
	UserID        string
	Demographics map[string]interface{}
	Preferences  map[string]interface{}
	Features     []float64
	Metadata     map[string]interface{}
	Quality      float64
}

// 辅助函数
func hashString(s string) int {
	h := 0
	for _, c := range s {
		h = h*31 + int(c)
	}
	return h
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}