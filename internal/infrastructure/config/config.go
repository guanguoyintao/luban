// Package config 配置管理
package config

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// ConfigManager 配置管理器接口
type ConfigManager interface {
	// 加载配置
	Load(configPath string) error
	LoadFromBytes(data []byte, format string) error
	
	// 获取配置
	Get(key string) interface{}
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetFloat64(key string) float64
	GetStringSlice(key string) []string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	
	// 设置配置
	Set(key string, value interface{})
	
	// 监听配置变化
	Watch(key string, callback func(key string, value interface{}))
	
	// 配置验证
	Validate() error
	
	// 获取所有配置
	AllSettings() map[string]interface{}
	
	// 保存配置
	Save(configPath string) error
}

// ViperConfigManager Viper配置管理器
type ViperConfigManager struct {
	viper     *viper.Viper
	watchers  map[string][]ConfigWatcher
	mu        sync.RWMutex
	validator ConfigValidator
}

// ConfigWatcher 配置监听器
type ConfigWatcher struct {
	Key      string
	Callback func(key string, value interface{})
}

// ConfigValidator 配置验证器
type ConfigValidator interface {
	Validate(config map[string]interface{}) error
}

// NewViperConfigManager 创建Viper配置管理器
func NewViperConfigManager() *ViperConfigManager {
	v := viper.New()
	v.SetConfigType("yaml")
	
	return &ViperConfigManager{
		viper:    v,
		watchers: make(map[string][]ConfigWatcher),
	}
}

// Load 加载配置文件
func (m *ViperConfigManager) Load(configPath string) error {
	m.viper.SetConfigFile(configPath)
	
	if err := m.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}
	
	return nil
}

// LoadFromBytes 从字节数组加载配置
func (m *ViperConfigManager) LoadFromBytes(data []byte, format string) error {
	m.viper.SetConfigType(format)
	
	if err := m.viper.ReadConfig(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("读取配置数据失败: %w", err)
	}
	
	return nil
}

// Get 获取配置值
func (m *ViperConfigManager) Get(key string) interface{} {
	return m.viper.Get(key)
}

// GetString 获取字符串配置
func (m *ViperConfigManager) GetString(key string) string {
	return m.viper.GetString(key)
}

// GetInt 获取整数配置
func (m *ViperConfigManager) GetInt(key string) int {
	return m.viper.GetInt(key)
}

// GetBool 获取布尔配置
func (m *ViperConfigManager) GetBool(key string) bool {
	return m.viper.GetBool(key)
}

// GetFloat64 获取浮点数配置
func (m *ViperConfigManager) GetFloat64(key string) float64 {
	return m.viper.GetFloat64(key)
}

// GetStringSlice 获取字符串切片配置
func (m *ViperConfigManager) GetStringSlice(key string) []string {
	return m.viper.GetStringSlice(key)
}

// GetStringMap 获取字符串映射配置
func (m *ViperConfigManager) GetStringMap(key string) map[string]interface{} {
	return m.viper.GetStringMap(key)
}

// GetStringMapString 获取字符串映射配置
func (m *ViperConfigManager) GetStringMapString(key string) map[string]string {
	return m.viper.GetStringMapString(key)
}

// Set 设置配置值
func (m *ViperConfigManager) Set(key string, value interface{}) {
	m.viper.Set(key, value)
	m.notifyWatchersForKey(key, value)
}

// Watch 监听配置变化
func (m *ViperConfigManager) Watch(key string, callback func(key string, value interface{})) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	watcher := ConfigWatcher{
		Key:      key,
		Callback: callback,
	}
	
	m.watchers[key] = append(m.watchers[key], watcher)
}

// Validate 验证配置
func (m *ViperConfigManager) Validate() error {
	if m.validator != nil {
		return m.validator.Validate(m.AllSettings())
	}
	return nil
}

// AllSettings 获取所有配置
func (m *ViperConfigManager) AllSettings() map[string]interface{} {
	return m.viper.AllSettings()
}

// Save 保存配置
func (m *ViperConfigManager) Save(configPath string) error {
	return m.viper.WriteConfigAs(configPath)
}

// SetValidator 设置验证器
func (m *ViperConfigManager) SetValidator(validator ConfigValidator) {
	m.validator = validator
}

// notifyWatchers 通知所有监听器
func (m *ViperConfigManager) notifyWatchers() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for key, watchers := range m.watchers {
		value := m.Get(key)
		for _, watcher := range watchers {
			if watcher.Callback != nil {
				watcher.Callback(key, value)
			}
		}
	}
}

// notifyWatchersForKey 通知指定key的监听器
func (m *ViperConfigManager) notifyWatchersForKey(key string, value interface{}) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if watchers, exists := m.watchers[key]; exists {
		for _, watcher := range watchers {
			if watcher.Callback != nil {
				watcher.Callback(key, value)
			}
		}
	}
}

// BaseConfigValidator 基础配置验证器
type BaseConfigValidator struct {
	requiredFields  []string
	fieldValidators map[string]FieldValidator
}

// FieldValidator 字段验证器
type FieldValidator func(value interface{}) error

// NewBaseConfigValidator 创建基础配置验证器
func NewBaseConfigValidator() *BaseConfigValidator {
	return &BaseConfigValidator{
		requiredFields:  []string{},
		fieldValidators: make(map[string]FieldValidator),
	}
}

// AddRequiredField 添加必填字段
func (v *BaseConfigValidator) AddRequiredField(field string) {
	v.requiredFields = append(v.requiredFields, field)
}

// AddFieldValidator 添加字段验证器
func (v *BaseConfigValidator) AddFieldValidator(field string, validator FieldValidator) {
	v.fieldValidators[field] = validator
}

// Validate 验证配置
func (v *BaseConfigValidator) Validate(config map[string]interface{}) error {
	// 检查必填字段
	for _, field := range v.requiredFields {
		if _, exists := config[field]; !exists {
			return fmt.Errorf("缺少必填字段: %s", field)
		}
	}
	
	// 验证字段值
	for field, validator := range v.fieldValidators {
		if value, exists := config[field]; exists {
			if err := validator(value); err != nil {
				return fmt.Errorf("字段 %s 验证失败: %w", field, err)
			}
		}
	}
	
	return nil
}

// RecommendationConfigValidator 推荐系统配置验证器
type RecommendationConfigValidator struct {
	*BaseConfigValidator
}

// NewRecommendationConfigValidator 创建推荐系统配置验证器
func NewRecommendationConfigValidator() *RecommendationConfigValidator {
	validator := &RecommendationConfigValidator{
		BaseConfigValidator: NewBaseConfigValidator(),
	}
	
	// 添加必填字段
	validator.AddRequiredField("server")
	validator.AddRequiredField("database")
	validator.AddRequiredField("redis")
	validator.AddRequiredField("algorithms")
	
	// 添加字段验证器
	validator.AddFieldValidator("server.port", validatePort)
	validator.AddFieldValidator("algorithms.enabled", validateAlgorithmList)
	
	return validator
}

// validatePort 验证端口
func validatePort(value interface{}) error {
	port, ok := value.(int)
	if !ok {
		return fmt.Errorf("端口必须是整数")
	}
	
	if port < 1 || port > 65535 {
		return fmt.Errorf("端口必须在1-65535之间")
	}
	
	return nil
}

// validateAlgorithmList 验证算法列表
func validateAlgorithmList(value interface{}) error {
	algorithms, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("算法列表必须是数组")
	}
	
	if len(algorithms) == 0 {
		return fmt.Errorf("至少需要启用一个算法")
	}
	
	validAlgorithms := []string{"collaborative_filtering", "content_based_filtering", "hybrid_filtering", "deep_learning", "popularity", "rule_based"}
	
	for _, algorithm := range algorithms {
		algStr, ok := algorithm.(string)
		if !ok {
			return fmt.Errorf("算法名称必须是字符串")
		}
		
		valid := false
		for _, validAlg := range validAlgorithms {
			if algStr == validAlg {
				valid = true
				break
			}
		}
		
		if !valid {
			return fmt.Errorf("无效的算法: %s", algStr)
		}
	}
	
	return nil
}