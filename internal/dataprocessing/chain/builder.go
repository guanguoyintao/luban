// Package chain 数据处理责任链构建器
package chain

import (
	"github.com/guanguoyintao/luban/internal/datacollection"
)

// ChainBuilder 责任链构建器
type ChainBuilder struct {
	processors []DataProcessor
}

// NewChainBuilder 创建责任链构建器
func NewChainBuilder() *ChainBuilder {
	return &ChainBuilder{
		processors: make([]DataProcessor, 0),
	}
}

// WithValidation 添加验证处理器
func (b *ChainBuilder) WithValidation() *ChainBuilder {
	b.processors = append(b.processors, NewValidationProcessor())
	return b
}

// WithNormalization 添加归一化处理器
func (b *ChainBuilder) WithNormalization() *ChainBuilder {
	b.processors = append(b.processors, NewNormalizationProcessor())
	return b
}

// WithFeatureExtraction 添加特征提取处理器
func (b *ChainBuilder) WithFeatureExtraction() *ChainBuilder {
	b.processors = append(b.processors, NewFeatureExtractionProcessor())
	return b
}

// WithQualityCheck 添加质量检查处理器
func (b *ChainBuilder) WithQualityCheck() *ChainBuilder {
	b.processors = append(b.processors, NewQualityCheckProcessor())
	return b
}

// Build 构建处理链
func (b *ChainBuilder) Build() *ProcessingChain {
	return NewProcessingChain(b.processors...)
}

// BuildDefaultChain 构建默认处理链
func BuildDefaultChain() *ProcessingChain {
	return NewChainBuilder().
		WithValidation().
		WithNormalization().
		WithFeatureExtraction().
		WithQualityCheck().
		Build()
}

// BuildUserBehaviorChain 构建用户行为数据处理链
func BuildUserBehaviorChain() *ProcessingChain {
	return NewChainBuilder().
		WithValidation().
		WithNormalization().
		WithFeatureExtraction().
		Build()
}

// BuildItemDataChain 构建物品数据处理链
func BuildItemDataChain() *ProcessingChain {
	return NewChainBuilder().
		WithValidation().
		WithFeatureExtraction().
		WithQualityCheck().
		Build()
}

// BuildUserDataChain 构建用户数据处理链
func BuildUserDataChain() *ProcessingChain {
	return NewChainBuilder().
		WithValidation().
		WithFeatureExtraction().
		WithQualityCheck().
		Build()
}

// ConvertUserBehavior 转换用户行为数据类型
func ConvertUserBehavior(behavior datacollection.UserBehavior) UserBehaviorData {
	return UserBehaviorData{
		UserID:    behavior.UserID,
		ItemID:    behavior.ItemID,
		Behavior:  string(behavior.Behavior),
		Value:     behavior.Value,
		Timestamp: behavior.Timestamp,
	}
}

// ConvertItemData 转换物品数据类型
func ConvertItemData(item datacollection.ItemData) ItemData {
	return ItemData{
		ItemID:      item.ItemID,
		Category:    item.Category,
		Title:       item.Title,
		Description: item.Description,
		Metadata:    item.Metadata,
	}
}

// ConvertUserData 转换用户数据类型
func ConvertUserData(user datacollection.UserData) UserData {
	return UserData{
		UserID:       user.UserID,
		Demographics: user.Demographics,
		Preferences:  user.Preferences,
		Metadata:     user.Metadata,
	}
}
