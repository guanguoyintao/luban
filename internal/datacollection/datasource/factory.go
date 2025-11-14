// Package datasource 数据源工厂模式实现
package datasource

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// DataSourceType 数据源类型
type DataSourceType string

const (
	DataSourceTypeMemory        DataSourceType = "memory"
	DataSourceTypeRedis         DataSourceType = "redis"
	DataSourceTypeMySQL         DataSourceType = "mysql"
	DataSourceTypeMongoDB       DataSourceType = "mongodb"
	DataSourceTypeElasticsearch DataSourceType = "elasticsearch"
)

// DataSourceConfig 数据源配置
type DataSourceConfig struct {
	Type     DataSourceType
	Name     string
	Address  string
	Port     int
	Username string
	Password string
	Database string
	Options  map[string]interface{}
}

// DataSourceFactory 数据源工厂
type DataSourceFactory struct {
	log      *logrus.Logger
	creators map[DataSourceType]DataSourceCreator
	sources  map[string]DataSource
	mu       sync.RWMutex
}

// DataSourceCreator 数据源创建器函数类型
type DataSourceCreator func(config DataSourceConfig, log *logrus.Logger) (DataSource, error)

// NewDataSourceFactory 创建数据源工厂
func NewDataSourceFactory(log *logrus.Logger) *DataSourceFactory {
	factory := &DataSourceFactory{
		log:      log,
		creators: make(map[DataSourceType]DataSourceCreator),
		sources:  make(map[string]DataSource),
	}

	// 注册默认的数据源创建器
	factory.registerDefaultCreators()

	return factory
}

// registerDefaultCreators 注册默认的数据源创建器
func (f *DataSourceFactory) registerDefaultCreators() {
	// 注册内存数据源
	f.RegisterCreator(DataSourceTypeMemory, func(config DataSourceConfig, log *logrus.Logger) (DataSource, error) {
		return NewMemoryDataSource(config, log), nil
	})

	// 这里可以注册其他数据源的创建器
	// 例如 Redis, MySQL, MongoDB, Elasticsearch 等
}

// RegisterCreator 注册数据源创建器
func (f *DataSourceFactory) RegisterCreator(dataSourceType DataSourceType, creator DataSourceCreator) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.creators[dataSourceType] = creator
	f.log.WithField("type", dataSourceType).Info("注册数据源创建器")
}

// CreateDataSource 创建数据源
func (f *DataSourceFactory) CreateDataSource(config DataSourceConfig) (DataSource, error) {
	f.mu.RLock()
	creator, exists := f.creators[config.Type]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("不支持的数据源类型: %s", config.Type)
	}

	source, err := creator(config, f.log)
	if err != nil {
		return nil, fmt.Errorf("创建数据源失败: %w", err)
	}

	// 缓存数据源实例
	f.mu.Lock()
	f.sources[config.Name] = source
	f.mu.Unlock()

	f.log.WithFields(logrus.Fields{
		"name": config.Name,
		"type": config.Type,
	}).Info("数据源创建成功")

	return source, nil
}

// GetDataSource 获取数据源
func (f *DataSourceFactory) GetDataSource(name string) (DataSource, error) {
	f.mu.RLock()
	source, exists := f.sources[name]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("数据源不存在: %s", name)
	}

	return source, nil
}

// CreateMultiDataSource 创建多数据源组合
func (f *DataSourceFactory) CreateMultiDataSource(configs []DataSourceConfig) (*MultiDataSource, error) {
	sources := make([]DataSource, 0, len(configs))

	for _, config := range configs {
		source, err := f.CreateDataSource(config)
		if err != nil {
			return nil, fmt.Errorf("创建数据源 %s 失败: %w", config.Name, err)
		}
		sources = append(sources, source)
	}

	return NewMultiDataSource(sources, f.log), nil
}

// GetAllDataSources 获取所有数据源
func (f *DataSourceFactory) GetAllDataSources() map[string]DataSource {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make(map[string]DataSource)
	for name, source := range f.sources {
		result[name] = source
	}
	return result
}

// Close 关闭所有数据源
func (f *DataSourceFactory) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	var lastErr error
	for name, source := range f.sources {
		if err := source.Close(); err != nil {
			f.log.WithError(err).WithField("source", name).Error("关闭数据源失败")
			lastErr = err
		}
	}

	// 清空数据源缓存
	f.sources = make(map[string]DataSource)

	return lastErr
}
