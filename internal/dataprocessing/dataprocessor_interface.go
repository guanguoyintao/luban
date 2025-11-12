package dataprocessing

import (
	"context"
	"time"
)

// 数据清洗状态
type DataCleaningStatus string

const (
	StatusClean      DataCleaningStatus = "clean"      // 数据已清洗
	StatusDirty      DataCleaningStatus = "dirty"      // 数据需要清洗
	StatusProcessing DataCleaningStatus = "processing" // 数据清洗中
	StatusError      DataCleaningStatus = "error"      // 数据清洗出错
)

// 数据质量指标
type DataQualityMetrics struct {
	Completeness float64 // 数据完整性
	Accuracy     float64 // 数据准确性
	Consistency  float64 // 数据一致性
	Timeliness   float64 // 数据时效性
	Validity     float64 // 数据有效性
}

// 处理后的用户行为数据
type ProcessedUserBehavior struct {
	UserID        string                 // 用户ID
	ItemID        string                 // 物品ID
	Behavior      string                 // 行为类型
	NormalizedValue float64              // 归一化后的行为值
	Timestamp     time.Time              // 时间戳
	Weight        float64                // 行为权重
	Features      map[string]interface{} // 特征数据
}

// 处理后的物品数据
type ProcessedItemData struct {
	ItemID      string                 // 物品ID
	Category    string                 // 类别
	Features    []float64              // 特征向量
	Metadata    map[string]interface{} // 元数据
	Quality     float64                // 数据质量评分
}

// 处理后的用户数据
type ProcessedUserData struct {
	UserID      string                 // 用户ID
	Features    []float64              // 用户特征向量
	Preferences map[string]interface{} // 用户偏好
	Quality     float64                // 数据质量评分
}

// 数据处理器接口
type DataProcessor interface {
	// 清洗用户行为数据
	CleanUserBehaviorData(ctx context.Context, rawData interface{}) (*ProcessedUserBehavior, error)
	
	// 批量清洗用户行为数据
	CleanUserBehaviorDataBatch(ctx context.Context, rawData []interface{}) ([]ProcessedUserBehavior, error)
	
	// 清洗物品数据
	CleanItemData(ctx context.Context, rawData interface{}) (*ProcessedItemData, error)
	
	// 批量清洗物品数据
	CleanItemDataBatch(ctx context.Context, rawData []interface{}) ([]ProcessedItemData, error)
	
	// 清洗用户数据
	CleanUserData(ctx context.Context, rawData interface{}) (*ProcessedUserData, error)
	
	// 批量清洗用户数据
	CleanUserDataBatch(ctx context.Context, rawData []interface{}) ([]ProcessedUserData, error)
	
	// 数据归一化处理
	NormalizeData(ctx context.Context, data []float64) ([]float64, error)
	
	// 特征提取
	ExtractFeatures(ctx context.Context, data interface{}) ([]float64, error)
	
	// 计算数据质量指标
	CalculateDataQuality(ctx context.Context, data interface{}) (*DataQualityMetrics, error)
	
	// 数据去重
	RemoveDuplicates(ctx context.Context, data []interface{}) ([]interface{}, error)
	
	// 处理缺失值
	HandleMissingValues(ctx context.Context, data interface{}) (interface{}, error)
	
	// 关闭处理器
	Close() error
}