package recommendation

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// 推荐引擎管理器
type RecommendationEngineManager struct {
	mu        sync.RWMutex
	engines   map[AlgorithmType]RecommendationEngine // 算法引擎映射
	log       *logrus.Logger
	config    *EngineConfig
}

// 引擎配置
type EngineConfig struct {
	DefaultAlgorithm      AlgorithmType
	MaxRecommendations    int
	MinConfidenceScore    float64
	EnableFallback        bool
	FallbackAlgorithm     AlgorithmType
}

// 创建新的推荐引擎管理器
func NewRecommendationEngineManager(log *logrus.Logger) *RecommendationEngineManager {
	if log == nil {
		log = logrus.New()
	}
	
	config := &EngineConfig{
		DefaultAlgorithm:   AlgorithmCollaborativeFiltering,
		MaxRecommendations: 50,
		MinConfidenceScore: 0.1,
		EnableFallback:     true,
		FallbackAlgorithm:  AlgorithmContentBasedFiltering,
	}
	
	manager := &RecommendationEngineManager{
		engines: make(map[AlgorithmType]RecommendationEngine),
		log:     log,
		config:  config,
	}
	
	// 注册默认算法引擎
	manager.registerDefaultEngines()
	
	return manager
}

// 注册默认算法引擎
func (m *RecommendationEngineManager) registerDefaultEngines() {
	// 这里可以注册具体的算法实现
	// 例如：m.RegisterEngine(AlgorithmCollaborativeFiltering, NewCollaborativeFilteringEngine(m.log))
	m.log.Info("注册默认推荐算法引擎")
}

// 注册算法引擎
func (m *RecommendationEngineManager) RegisterEngine(algorithm AlgorithmType, engine RecommendationEngine) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.engines[algorithm] = engine
	m.log.WithField("algorithm", algorithm).Info("注册推荐算法引擎成功")
}

// 生成推荐
func (m *RecommendationEngineManager) Recommend(ctx context.Context, request RecommendationRequest) (*RecommendationResponse, error) {
	startTime := time.Now()
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// 确定使用的算法
	algorithm := request.Algorithm
	if algorithm == "" {
		algorithm = m.config.DefaultAlgorithm
	}
	
	// 获取对应的引擎
	engine, exists := m.engines[algorithm]
	if !exists {
		return nil, &RecommendationError{Message: fmt.Sprintf("算法引擎不存在: %s", algorithm)}
	}
	
	// 生成推荐
	response, err := engine.Recommend(ctx, request)
	if err != nil {
		m.log.WithError(err).WithField("algorithm", algorithm).Error("推荐生成失败")
		
		// 如果启用回退算法，尝试使用回退算法
		if m.config.EnableFallback && algorithm != m.config.FallbackAlgorithm {
			m.log.WithField("fallback_algorithm", m.config.FallbackAlgorithm).Info("使用回退算法")
			fallbackEngine, fallbackExists := m.engines[m.config.FallbackAlgorithm]
			if fallbackExists {
				request.Algorithm = m.config.FallbackAlgorithm
				return fallbackEngine.Recommend(ctx, request)
			}
		}
		
		return nil, err
	}
	
	// 过滤低置信度推荐
	filteredRecommendations := m.filterLowConfidenceRecommendations(response.Recommendations)
	
	// 限制推荐数量
	if len(filteredRecommendations) > request.Limit && request.Limit > 0 {
		filteredRecommendations = filteredRecommendations[:request.Limit]
	}
	
	response.Recommendations = filteredRecommendations
	response.TotalCount = len(filteredRecommendations)
	response.ProcessingTime = time.Since(startTime).Milliseconds()
	
	m.log.WithFields(logrus.Fields{
		"user_id":      request.UserID,
		"algorithm":    algorithm,
		"recommendations": len(filteredRecommendations),
		"processing_time": response.ProcessingTime,
	}).Info("推荐生成成功")
	
	return response, nil
}

// 批量生成推荐
func (m *RecommendationEngineManager) RecommendBatch(ctx context.Context, requests []RecommendationRequest) ([]*RecommendationResponse, error) {
	results := make([]*RecommendationResponse, len(requests))
	
	for i, request := range requests {
		response, err := m.Recommend(ctx, request)
		if err != nil {
			m.log.WithError(err).WithField("user_id", request.UserID).Error("批量推荐生成失败")
			results[i] = &RecommendationResponse{
				UserID: request.UserID,
				Metadata: map[string]interface{}{
					"error": err.Error(),
				},
			}
		} else {
			results[i] = response
		}
	}
	
	return results, nil
}

// 获取推荐解释
func (m *RecommendationEngineManager) ExplainRecommendation(ctx context.Context, userID string, itemID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// 使用默认算法引擎获取解释
	engine, exists := m.engines[m.config.DefaultAlgorithm]
	if !exists {
		return "", &RecommendationError{Message: "默认算法引擎不存在"}
	}
	
	return engine.ExplainRecommendation(ctx, userID, itemID)
}

// 更新推荐模型
func (m *RecommendationEngineManager) UpdateModel(ctx context.Context, data interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// 更新所有注册的算法模型
	var lastError error
	for algorithm, engine := range m.engines {
		if err := engine.UpdateModel(ctx, data); err != nil {
			m.log.WithError(err).WithField("algorithm", algorithm).Error("更新推荐模型失败")
			lastError = err
		}
	}
	
	if lastError != nil {
		return lastError
	}
	
	m.log.Info("更新推荐模型成功")
	return nil
}

// 获取推荐算法列表
func (m *RecommendationEngineManager) GetAvailableAlgorithms(ctx context.Context) ([]AlgorithmType, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	algorithms := make([]AlgorithmType, 0, len(m.engines))
	for algorithm := range m.engines {
		algorithms = append(algorithms, algorithm)
	}
	
	sort.Slice(algorithms, func(i, j int) bool {
		return string(algorithms[i]) < string(algorithms[j])
	})
	
	return algorithms, nil
}

// 获取算法参数
func (m *RecommendationEngineManager) GetAlgorithmParameters(ctx context.Context, algorithm AlgorithmType) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	engine, exists := m.engines[algorithm]
	if !exists {
		return nil, &RecommendationError{Message: fmt.Sprintf("算法引擎不存在: %s", algorithm)}
	}
	
	return engine.GetAlgorithmParameters(ctx, algorithm)
}

// 设置算法参数
func (m *RecommendationEngineManager) SetAlgorithmParameters(ctx context.Context, algorithm AlgorithmType, parameters map[string]interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	engine, exists := m.engines[algorithm]
	if !exists {
		return &RecommendationError{Message: fmt.Sprintf("算法引擎不存在: %s", algorithm)}
	}
	
	return engine.SetAlgorithmParameters(ctx, algorithm, parameters)
}

// 获取推荐统计信息
func (m *RecommendationEngineManager) GetRecommendationStats(ctx context.Context, userID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 收集所有算法的统计信息
	for algorithm, engine := range m.engines {
		engineStats, err := engine.GetRecommendationStats(ctx, userID)
		if err != nil {
			m.log.WithError(err).WithField("algorithm", algorithm).Error("获取推荐统计信息失败")
			continue
		}
		stats[string(algorithm)] = engineStats
	}
	
	return stats, nil
}

// 记录用户反馈
func (m *RecommendationEngineManager) RecordFeedback(ctx context.Context, userID string, itemID string, feedback interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// 记录到所有算法引擎
	var lastError error
	for algorithm, engine := range m.engines {
		if err := engine.RecordFeedback(ctx, userID, itemID, feedback); err != nil {
			m.log.WithError(err).WithFields(logrus.Fields{
				"algorithm": algorithm,
				"user_id":   userID,
				"item_id":   itemID,
			}).Error("记录用户反馈失败")
			lastError = err
		}
	}
	
	if lastError != nil {
		return lastError
	}
	
	m.log.WithFields(logrus.Fields{
		"user_id":  userID,
		"item_id":  itemID,
	}).Info("记录用户反馈成功")
	
	return nil
}

// 关闭推荐引擎
func (m *RecommendationEngineManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 关闭所有算法引擎
	var lastError error
	for algorithm, engine := range m.engines {
		if err := engine.Close(); err != nil {
			m.log.WithError(err).WithField("algorithm", algorithm).Error("关闭算法引擎失败")
			lastError = err
		}
	}
	
	m.log.Info("关闭推荐引擎管理器")
	
	if lastError != nil {
		return lastError
	}
	return nil
}

// 过滤低置信度推荐
func (m *RecommendationEngineManager) filterLowConfidenceRecommendations(recommendations []RecommendationResult) []RecommendationResult {
	filtered := make([]RecommendationResult, 0)
	
	for _, rec := range recommendations {
		if rec.Confidence >= m.config.MinConfidenceScore {
			filtered = append(filtered, rec)
		}
	}
	
	// 按得分排序
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Score > filtered[j].Score
	})
	
	return filtered
}

// 设置引擎配置
func (m *RecommendationEngineManager) SetConfig(config *EngineConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = config
}

// 获取引擎配置
func (m *RecommendationEngineManager) GetConfig() *EngineConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// 推荐错误
type RecommendationError struct {
	Message string
}

func (e *RecommendationError) Error() string {
	return "推荐引擎错误: " + e.Message
}