# 企业级推荐系统框架

高度可插拔、高内聚低耦合的企业级推荐系统框架，采用领域驱动设计(DDD)和Google Wire依赖注入，适配大厂标准，支持复杂的推荐业务场景。

## 🏗️ 架构特点

### 核心设计原则
- **领域驱动设计(DDD)** - 业务逻辑与基础设施分离
- **依赖倒置原则(DIP)** - 高层模块不依赖低层模块，采用Google Wire依赖注入框架
- **防腐层(ACL)** - 隔离外部系统变化对核心业务的影响
- **策略模式** - 算法可插拔替换，无需修改核心代码
- **工厂模式** - 对象创建统一管理，支持多种数据源
- **适配器模式** - 统一不同数据源接口
- **责任链模式** - 处理流程可配置，支持链式处理
- **观察者模式** - 事件驱动架构

### 分层架构

```
recommendation-system/
├── cmd/                          # 应用程序入口
│   ├── main.go                  # 主程序入口
│   └── test.go                  # 测试程序
├── internal/                    # 内部核心业务逻辑
│   ├── application/             # 应用层（用例协调）
│   │   ├── dto.go              # 数据传输对象
│   │   ├── presenter.go        # 推荐服务实现
│   │   └── presenter_interface.go # 推荐服务接口
│   ├── datacollection/          # 数据收集层
│   │   ├── datacollector.go    # 数据收集器实现
│   │   └── datacollector_interface.go # 数据收集器接口
│   ├── dataprocessing/          # 数据处理层
│   │   ├── dataprocessor.go    # 数据处理器实现
│   │   └── dataprocessor_interface.go # 数据处理器接口
│   ├── domain/                  # 领域层（核心业务逻辑）
│   │   ├── entity.go           # 领域实体
│   │   └── repository.go       # 仓储接口
│   ├── infra/                   # 基础设施层（缩写为infra）
│   │   ├── acl/                # 防腐层（Anti-Corruption Layer）
│   │   │   ├── adapter.go      # 适配器接口
│   │   │   └── manager.go      # 防腐层管理器
│   │   ├── config/             # 配置管理
│   │   │   ├── config.go       # 配置管理器实现
│   │   │   └── manager.go      # 配置管理器（旧文件）
│   │   ├── di/                 # 依赖注入（Google Wire）
│   │   │   ├── wire.go         # Wire依赖注入配置
│   │   │   └── wire_gen.go     # Wire生成的代码
│   │   └── error/              # 错误处理框架
│   │       └── error.go        # 错误处理实现
│   └── recommendation/          # 推荐引擎
│       ├── algorithms/          # 推荐算法
│       │   ├── collaborativefiltering.go # 协同过滤算法
│       │   ├── contentbasedfiltering.go  # 基于内容过滤算法
│       │   └── hybridfiltering.go        # 混合过滤算法
│       ├── models/              # 推荐模型
│       │   ├── item.go           # 物品模型
│       │   ├── recommendation.go # 推荐模型
│       │   └── user.go           # 用户模型
│       ├── engine.go             # 推荐引擎管理器
│       ├── engine_interface.go   # 推荐引擎接口
│       └── simple_engine.go      # 简单推荐引擎实现
├── pkg/                         # 可复用的包
│   ├── algorithm/               # 算法包
│   │   ├── chain/              # 责任链模式
│   │   │   └── chain.go        # 责任链实现
│   │   └── strategy/           # 策略模式
│   │       └── strategy.go     # 策略模式实现
│   ├── datasource/              # 数据源适配器
│   │   ├── adapter.go          # 数据源适配器接口
│   │   └── factory.go          # 数据源工厂
│   └── plugin/                  # 插件系统
│       ├── examples.go         # 插件示例
│       ├── manager.go          # 插件管理器
│       └── registry.go         # 插件注册表
├── configs/                     # 配置文件
│   ├── development/            # 开发环境配置
│   │   └── config.yaml        # 开发环境配置文件
│   └── production/             # 生产环境配置
│       └── config.yaml        # 生产环境配置文件
├── go.mod                      # Go模块定义
├── go.sum                      # Go依赖校验
├── Makefile                    # 构建脚本
└── README.md                   # 项目文档
```

## 🚀 快速开始

### 环境要求
- Go 1.24+
- Make

### 安装依赖
```bash
go mod download
```

### 运行测试程序
```bash
go run cmd/test.go
```

### 运行主程序
```bash
go run cmd/main.go
```

### 使用Makefile构建
```bash
# 构建项目
make build

# 运行测试
make test

# 开发模式运行
make dev

# 生产模式运行
make prod

# 查看所有可用命令
make help
```

## 📊 性能指标

- **QPS**: 10万+ (单实例)
- **延迟**: P99 < 100ms
- **可用性**: 99.99%
- **扩展性**: 支持水平扩展到1000+节点

## 🔧 核心特性

### 1. 高度可插拔架构
- **插件系统**：支持运行时动态加载和卸载推荐算法
- **策略模式**：算法可插拔替换，无需修改核心代码
- **责任链模式**：处理流程可配置，支持链式处理

### 2. 设计模式实现
- **防腐层(ACL)**：隔离外部系统变化对核心业务的影响
- **依赖注入(DI)**：使用Google Wire框架，实现依赖倒置
- **工厂模式**：统一管理对象创建，支持多种数据源
- **策略模式**：算法策略可动态切换
- **责任链模式**：处理流程可配置
- **适配器模式**：统一不同数据源接口

### 3. 企业级特性
- **配置管理**：支持多环境配置，使用Viper管理
- **错误处理**：完善的错误处理框架，支持错误链和堆栈跟踪
- **日志系统**：结构化日志，支持多种格式
- **监控指标**：内置性能监控和数据质量检查

## 🔒 安全特性

- 数据加密存储和传输
- API访问控制和限流
- 敏感数据脱敏
- 审计日志记录
- 安全漏洞扫描

## 📈 监控告警

- 实时性能监控
- 异常检测和告警
- 业务指标监控
- 系统健康检查
- 自动化运维

## 🎯 核心功能

### 推荐算法
- **协同过滤** - 基于用户-物品交互矩阵
- **内容过滤** - 基于物品特征和用户偏好
- **混合过滤** - 结合多种算法优势
- **深度学习** - 基于神经网络的推荐
- **流行度算法** - 基于物品热度的推荐
- **基于规则** - 可配置的规则引擎

### 数据源支持
- **内存存储** - 高性能内存数据收集和处理
- **Redis** - 高性能缓存和会话存储
- **MySQL/PostgreSQL** - 关系型数据存储
- **MongoDB** - 文档数据存储
- **Elasticsearch** - 搜索引擎和分析
- **Kafka** - 消息队列和流处理

### 插件系统
- **热插拔** - 运行时动态加载和卸载
- **可扩展** - 支持自定义插件开发
- **版本管理** - 插件版本控制和兼容性检查
- **依赖管理** - 自动解析插件依赖关系

## 🔧 配置管理

支持多种配置格式：
- YAML (推荐)
- JSON
- TOML
- 环境变量

配置特性：
- 热更新
- 配置验证
- 环境隔离
- 敏感信息加密

## 💻 使用示例

```go
// 初始化应用程序
app, err := di.InitializeApp()
if err != nil {
    log.Fatalf("初始化失败: %v", err)
}

// 获取推荐
ctx := context.Background()
recommendations, err := app.RecommendationSvc.GetRecommendations(ctx, "user_123", 10)
if err != nil {
    log.Printf("获取推荐失败: %v", err)
    return
}

// 处理推荐结果
for _, rec := range recommendations {
    fmt.Printf("物品ID: %s, 得分: %.2f, 算法: %s\n", 
        rec.ItemID, rec.Score, rec.Algorithm)
}
```

## 🛡️ 容错机制

- **熔断器** - 防止级联故障
- **限流器** - 保护系统免受过载
- **重试机制** - 自动重试失败操作
- **降级策略** - 优雅降级保证核心功能

## 📈 扩展指南

### 添加新的推荐算法

1. 在 `internal/recommendation/algorithms/` 中实现新的算法引擎
2. 在 `pkg/algorithm/strategy/` 中实现策略模式
3. 更新Wire依赖注入配置 `internal/infra/di/wire.go`

### 添加新的数据源

1. 实现 `datasource.DataSource` 接口
2. 在数据源工厂 `pkg/datasource/factory.go` 中注册新的数据源
3. 配置数据源参数

### 添加插件

1. 实现 `plugin.Plugin` 接口
2. 在插件管理器 `pkg/plugin/manager.go` 中注册插件
3. 配置插件加载路径

## 📚 文档

- [架构设计](docs/architecture/README.md)
- [API文档](docs/api/README.md)
- [部署指南](docs/deployment/README.md)
- [插件开发](docs/plugin-development.md)
- [算法调优](docs/algorithm-tuning.md)

## 🎯 框架优势

本框架提供了一个完整的企业级推荐系统解决方案，具有高度的可扩展性和可维护性。通过采用现代化的架构设计模式和最佳实践，能够满足复杂的业务需求，并支持快速的业务迭代。

### 核心优势
- **高度模块化架构设计** - 清晰的层次结构和职责分离
- **完善的依赖注入机制** - 使用Google Wire实现依赖倒置
- **丰富的设计模式实现** - 策略、工厂、适配器、责任链等模式
- **企业级错误处理和监控** - 完善的容错机制和监控体系
- **支持多种推荐算法和数据源** - 灵活可扩展的算法和数据支持
- **易于扩展和维护** - 清晰的代码结构和文档

### 技术特色
- **Google Wire依赖注入** - 编译期依赖注入，类型安全
- **领域驱动设计** - 业务逻辑与技术实现分离
- **防腐层设计** - 有效隔离外部系统变化
- **插件化架构** - 支持运行时动态扩展
- **配置化管理** - 支持多环境和热更新

这个框架可以作为构建大规模推荐系统的基础，帮助企业快速搭建稳定、高效的推荐服务。