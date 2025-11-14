//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sirupsen/logrus"
	
	"recommendation-system/internal/application"
	"recommendation-system/internal/datacollection"
	"recommendation-system/internal/datacollection/datasource"
	"recommendation-system/internal/dataprocessing"
	"recommendation-system/internal/dataprocessing/chain"
	"recommendation-system/internal/domain"
	"recommendation-system/internal/infra/config"
	"recommendation-system/internal/recommendation"
	"recommendation-system/internal/recommendation/strategy"
	"recommendation-system/pkg/plugin"
)

// ProviderSet 定义所有依赖提供者
var ProviderSet = wire.NewSet(
	// 基础设施层
	NewLogger,
	
	// 配置管理
	config.NewViperConfigManager,
	wire.Bind(new(config.ConfigManager), new(*config.ViperConfigManager)),
	
	// 数据收集层 - 工厂和适配器模式
	NewDataSourceFactory,
	NewMemoryDataSourceConfig,
	NewMultiDataSource,
	wire.Bind(new(datasource.DataSource), new(*datasource.MultiDataSource)),
	
	// 数据处理层 - 责任链模式
	dataprocessing.NewMemoryDataProcessor,
	wire.Bind(new(dataprocessing.DataProcessor), new(*dataprocessing.MemoryDataProcessor)),
	
	// 责任链构建器
	NewProcessingChainBuilder,
	
	// 推荐引擎层 - 策略模式
	NewRecommendationEngine,
	wire.Bind(new(domain.RecommendationService), new(*recommendation.SimpleRecommendationEngine)),
	
	// 排序策略
	NewRankingStrategies,
	
	// 责任链
	NewProcessingChain,
	
	// 应用服务
	application.NewRecommendationPresenter,
	wire.Bind(new(application.RecommendationUseCase), new(*application.RecommendationPresenter)),
	
	// 插件管理
	NewPluginManager,
	wire.Bind(new(plugin.PluginManager), new(*plugin.PluginManager)),
	
	// 应用程序
	NewApplication,
)

// NewLogger 创建日志记录器
func NewLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	return logger
}

// NewDataSourceFactory 创建数据源工厂
func NewDataSourceFactory(logger *logrus.Logger) *datasource.DataSourceFactory {
	return datasource.NewDataSourceFactory(logger)
}

// NewMemoryDataSourceConfig 创建内存数据源配置
func NewMemoryDataSourceConfig() datasource.DataSourceConfig {
	return datasource.DataSourceConfig{
		Type: datasource.DataSourceTypeMemory,
		Name: "memory_data_source",
		Options: map[string]interface{}{
			"max_items": 10000,
			"ttl":       3600,
		},
	}
}

// NewMultiDataSource 创建多数据源适配器
func NewMultiDataSource(factory *datasource.DataSourceFactory, config datasource.DataSourceConfig, logger *logrus.Logger) (*datasource.MultiDataSource, error) {
	// 创建内存数据源
	memorySource, err := factory.CreateDataSource(config)
	if err != nil {
		return nil, err
	}
	
	// 创建多数据源适配器
	return datasource.NewMultiDataSource([]datasource.DataSource{memorySource}, logger), nil
}

// NewProcessingChainBuilder 创建责任链构建器
func NewProcessingChainBuilder() *chain.ChainBuilder {
	return chain.NewChainBuilder()
}

// NewProcessingChain 创建数据处理责任链
func NewProcessingChain(builder *chain.ChainBuilder) *chain.ProcessingChain {
	return builder.
		WithValidation().
		WithNormalization().
		WithFeatureExtraction().
		WithQualityCheck().
		Build()
}

// NewRankingStrategies 创建排序策略
func NewRankingStrategies() []strategy.RankingStrategy {
	return strategy.BuildDefaultStrategies()
}

// NewRecommendationEngine 创建推荐引擎
func NewRecommendationEngine(
	logger *logrus.Logger,
	dataSource datasource.DataSource,
	dataProcessor dataprocessing.DataProcessor,
	rankingStrategies []strategy.RankingStrategy,
) *recommendation.SimpleRecommendationEngine {
	return recommendation.NewSimpleRecommendationEngine(
		logger,
		dataSource,
		dataProcessor,
		rankingStrategies,
	)
}

// NewPluginManager 创建插件管理器
func NewPluginManager(logger *logrus.Logger) *plugin.PluginManager {
	return plugin.NewPluginManager(logger)
}

// InitializeApp 初始化应用程序
func InitializeApp() (*Application, error) {
	wire.Build(ProviderSet)
	return nil, nil
}

// Application 应用程序容器
type Application struct {
	ConfigManager      config.ConfigManager
	DataSourceFactory  *datasource.DataSourceFactory
	ProcessingChain    *chain.ProcessingChain
	RankingStrategies  []strategy.RankingStrategy
	RecommendationSvc  domain.RecommendationService
	PluginManager      *plugin.PluginManager
	Logger             *logrus.Logger
}

// NewApplication 创建应用程序
func NewApplication(
	configManager config.ConfigManager,
	dataSourceFactory *datasource.DataSourceFactory,
	processingChain *chain.ProcessingChain,
	rankingStrategies []strategy.RankingStrategy,
	recommendationSvc domain.RecommendationService,
	pluginManager *plugin.PluginManager,
	logger *logrus.Logger,
) *Application {
	return &Application{
		ConfigManager:     configManager,
		DataSourceFactory: dataSourceFactory,
		ProcessingChain:   processingChain,
		RankingStrategies: rankingStrategies,
		RecommendationSvc: recommendationSvc,
		PluginManager:     pluginManager,
		Logger:            logger,
	}
}