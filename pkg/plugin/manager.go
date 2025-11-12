// Package plugin 插件系统
package plugin

import (
	"context"
	"fmt"
	"plugin"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Plugin 插件接口
type Plugin interface {
	// 插件信息
	GetInfo() PluginInfo
	
	// 初始化插件
	Init(ctx context.Context, config map[string]interface{}) error
	
	// 启动插件
	Start(ctx context.Context) error
	
	// 停止插件
	Stop(ctx context.Context) error
	
	// 获取插件状态
	GetStatus() PluginStatus
	
	// 健康检查
	HealthCheck(ctx context.Context) error
}

// PluginInfo 插件信息
type PluginInfo struct {
	ID          string
	Name        string
	Version     string
	Description string
	Author      string
	Type        PluginType
	Dependencies []string
}

// PluginType 插件类型
type PluginType string

const (
	PluginTypeAlgorithm PluginType = "algorithm"
	PluginTypeDataSource PluginType = "datasource"
	PluginTypeStorage    PluginType = "storage"
	PluginTypeService    PluginType = "service"
	PluginTypeMiddleware PluginType = "middleware"
)

// PluginStatus 插件状态
type PluginStatus struct {
	State      PluginState
	StartTime  *time.Time
	Error      error
	LastUpdate time.Time
}

// PluginState 插件状态枚举
type PluginState string

const (
	PluginStateIdle     PluginState = "idle"
	PluginStateLoading  PluginState = "loading"
	PluginStateRunning  PluginState = "running"
	PluginStateStopping PluginState = "stopping"
	PluginStateStopped  PluginState = "stopped"
	PluginStateError    PluginState = "error"
)

// PluginManager 插件管理器
type PluginManager struct {
	mu          sync.RWMutex
	plugins     map[string]Plugin
	registry    *PluginRegistry
	loader      *PluginLoader
	config      *PluginConfig
	log         *logrus.Logger
}

// PluginRegistry 插件注册表
type PluginRegistry struct {
	mu       sync.RWMutex
	factories map[string]PluginFactory
}

// PluginFactory 插件工厂接口
type PluginFactory interface {
	Create(info PluginInfo) (Plugin, error)
	GetType() PluginType
}

// PluginLoader 插件加载器
type PluginLoader struct {
	mu       sync.RWMutex
	plugins  map[string]*LoadedPlugin
	log      *logrus.Logger
}

// LoadedPlugin 已加载的插件
type LoadedPlugin struct {
	Info     PluginInfo
	Plugin   Plugin
	Path     string
	LoadedAt time.Time
}

// PluginConfig 插件配置
type PluginConfig struct {
	Enabled        bool
	PluginDir      string
	AutoLoad       bool
	LoadOrder      []string
	Configurations map[string]map[string]interface{}
}

// NewPluginManager 创建插件管理器
func NewPluginManager(config *PluginConfig, log *logrus.Logger) *PluginManager {
	if log == nil {
		log = logrus.New()
	}
	
	if config == nil {
		config = &PluginConfig{
			Enabled:        true,
			PluginDir:      "./plugins",
			AutoLoad:       true,
			LoadOrder:      []string{},
			Configurations: make(map[string]map[string]interface{}),
		}
	}
	
	return &PluginManager{
		plugins:  make(map[string]Plugin),
		registry: NewPluginRegistry(),
		loader:   NewPluginLoader(log),
		config:   config,
		log:      log,
	}
}

// NewPluginRegistry 创建插件注册表
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		factories: make(map[string]PluginFactory),
	}
}

// NewPluginLoader 创建插件加载器
func NewPluginLoader(log *logrus.Logger) *PluginLoader {
	if log == nil {
		log = logrus.New()
	}
	
	return &PluginLoader{
		plugins: make(map[string]*LoadedPlugin),
		log:     log,
	}
}

// RegisterFactory 注册插件工厂
func (r *PluginRegistry) RegisterFactory(factory PluginFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	factoryType := string(factory.GetType())
	if _, exists := r.factories[factoryType]; exists {
		return fmt.Errorf("插件工厂已存在: %s", factoryType)
	}
	
	r.factories[factoryType] = factory
	return nil
}

// GetFactory 获取插件工厂
func (r *PluginRegistry) GetFactory(pluginType string) (PluginFactory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	factory, exists := r.factories[pluginType]
	if !exists {
		return nil, fmt.Errorf("插件工厂不存在: %s", pluginType)
	}
	
	return factory, nil
}

// GetSupportedTypes 获取支持的插件类型
func (r *PluginRegistry) GetSupportedTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	types := make([]string, 0, len(r.factories))
	for pluginType := range r.factories {
		types = append(types, pluginType)
	}
	
	return types
}

// LoadPlugin 加载插件
func (l *PluginLoader) LoadPlugin(path string) (*LoadedPlugin, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// 打开插件文件
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开插件文件失败: %w", err)
	}
	
	// 查找插件信息符号
	infoSym, err := p.Lookup("PluginInfo")
	if err != nil {
		return nil, fmt.Errorf("未找到插件信息: %w", err)
	}
	
	info, ok := infoSym.(*PluginInfo)
	if !ok {
		return nil, fmt.Errorf("插件信息类型错误")
	}
	
	// 查找插件符号
	pluginSym, err := p.Lookup("Plugin")
	if err != nil {
		return nil, fmt.Errorf("未找到插件: %w", err)
	}
	
	plugin, ok := pluginSym.(Plugin)
	if !ok {
		return nil, fmt.Errorf("插件类型错误")
	}
	
	loadedPlugin := &LoadedPlugin{
		Info:     *info,
		Plugin:   plugin,
		Path:     path,
		LoadedAt: time.Now(),
	}
	
	l.plugins[info.ID] = loadedPlugin
	l.log.WithField("plugin_id", info.ID).Info("插件加载成功")
	
	return loadedPlugin, nil
}

// UnloadPlugin 卸载插件
func (l *PluginLoader) UnloadPlugin(pluginID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if _, exists := l.plugins[pluginID]; !exists {
		return fmt.Errorf("插件未加载: %s", pluginID)
	}
	
	delete(l.plugins, pluginID)
	l.log.WithField("plugin_id", pluginID).Info("插件卸载成功")
	
	return nil
}

// GetLoadedPlugin 获取已加载的插件
func (l *PluginLoader) GetLoadedPlugin(pluginID string) (*LoadedPlugin, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	plugin, exists := l.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("插件未加载: %s", pluginID)
	}
	
	return plugin, nil
}

// GetAllLoadedPlugins 获取所有已加载的插件
func (l *PluginLoader) GetAllLoadedPlugins() map[string]*LoadedPlugin {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	result := make(map[string]*LoadedPlugin)
	for id, plugin := range l.plugins {
		result[id] = plugin
	}
	
	return result
}

// RegisterPlugin 注册插件
func (m *PluginManager) RegisterPlugin(plugin Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	info := plugin.GetInfo()
	if _, exists := m.plugins[info.ID]; exists {
		return fmt.Errorf("插件已存在: %s", info.ID)
	}
	
	m.plugins[info.ID] = plugin
	m.log.WithField("plugin_id", info.ID).Info("插件注册成功")
	
	return nil
}

// UnregisterPlugin 注销插件
func (m *PluginManager) UnregisterPlugin(pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.plugins[pluginID]; !exists {
		return fmt.Errorf("插件不存在: %s", pluginID)
	}
	
	delete(m.plugins, pluginID)
	m.log.WithField("plugin_id", pluginID).Info("插件注销成功")
	
	return nil
}

// LoadPlugin 加载并注册插件
func (m *PluginManager) LoadPlugin(path string) error {
	// 使用插件加载器加载插件
	loadedPlugin, err := m.loader.LoadPlugin(path)
	if err != nil {
		return fmt.Errorf("加载插件失败: %w", err)
	}
	
	// 注册插件
	return m.RegisterPlugin(loadedPlugin.Plugin)
}

// InitPlugin 初始化插件
func (m *PluginManager) InitPlugin(ctx context.Context, pluginID string) error {
	m.mu.RLock()
	plugin, exists := m.plugins[pluginID]
	m.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("插件不存在: %s", pluginID)
	}
	
	config, exists := m.config.Configurations[pluginID]
	if !exists {
		config = make(map[string]interface{})
	}
	
	if err := plugin.Init(ctx, config); err != nil {
		return fmt.Errorf("初始化插件失败: %w", err)
	}
	
	m.log.WithField("plugin_id", pluginID).Info("插件初始化成功")
	
	return nil
}

// StartPlugin 启动插件
func (m *PluginManager) StartPlugin(ctx context.Context, pluginID string) error {
	m.mu.RLock()
	plugin, exists := m.plugins[pluginID]
	m.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("插件不存在: %s", pluginID)
	}
	
	if err := plugin.Start(ctx); err != nil {
		return fmt.Errorf("启动插件失败: %w", err)
	}
	
	m.log.WithField("plugin_id", pluginID).Info("插件启动成功")
	
	return nil
}

// StopPlugin 停止插件
func (m *PluginManager) StopPlugin(ctx context.Context, pluginID string) error {
	m.mu.RLock()
	plugin, exists := m.plugins[pluginID]
	m.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("插件不存在: %s", pluginID)
	}
	
	if err := plugin.Stop(ctx); err != nil {
		return fmt.Errorf("停止插件失败: %w", err)
	}
	
	m.log.WithField("plugin_id", pluginID).Info("插件停止成功")
	
	return nil
}

// GetPlugin 获取插件
func (m *PluginManager) GetPlugin(pluginID string) (Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	plugin, exists := m.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("插件不存在: %s", pluginID)
	}
	
	return plugin, nil
}

// GetAllPlugins 获取所有插件
func (m *PluginManager) GetAllPlugins() map[string]Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]Plugin)
	for id, plugin := range m.plugins {
		result[id] = plugin
	}
	
	return result
}

// GetPluginsByType 按类型获取插件
func (m *PluginManager) GetPluginsByType(pluginType PluginType) map[string]Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]Plugin)
	for id, plugin := range m.plugins {
		if plugin.GetInfo().Type == pluginType {
			result[id] = plugin
		}
	}
	
	return result
}

// HealthCheck 健康检查
func (m *PluginManager) HealthCheck(ctx context.Context) map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	results := make(map[string]error)
	
	for id, plugin := range m.plugins {
		if err := plugin.HealthCheck(ctx); err != nil {
			results[id] = err
			m.log.WithError(err).WithField("plugin_id", id).Error("插件健康检查失败")
		} else {
			results[id] = nil
		}
	}
	
	return results
}

// InitAllPlugins 初始化所有插件
func (m *PluginManager) InitAllPlugins(ctx context.Context) error {
	m.mu.RLock()
	plugins := m.plugins
	m.mu.RUnlock()
	
	var lastError error
	for id := range plugins {
		if err := m.InitPlugin(ctx, id); err != nil {
			m.log.WithError(err).WithField("plugin_id", id).Error("插件初始化失败")
			lastError = err
		}
	}
	
	return lastError
}

// StartAllPlugins 启动所有插件
func (m *PluginManager) StartAllPlugins(ctx context.Context) error {
	m.mu.RLock()
	plugins := m.plugins
	m.mu.RUnlock()
	
	var lastError error
	for id := range plugins {
		if err := m.StartPlugin(ctx, id); err != nil {
			m.log.WithError(err).WithField("plugin_id", id).Error("插件启动失败")
			lastError = err
		}
	}
	
	return lastError
}

// StopAllPlugins 停止所有插件
func (m *PluginManager) StopAllPlugins(ctx context.Context) error {
	m.mu.RLock()
	plugins := m.plugins
	m.mu.RUnlock()
	
	var lastError error
	for id := range plugins {
		if err := m.StopPlugin(ctx, id); err != nil {
			m.log.WithError(err).WithField("plugin_id", id).Error("插件停止失败")
			lastError = err
		}
	}
	
	return lastError
}

// BasePlugin 基础插件
type BasePlugin struct {
	info   PluginInfo
	status PluginStatus
	mu     sync.RWMutex
}

// NewBasePlugin 创建基础插件
func NewBasePlugin(info PluginInfo) *BasePlugin {
	return &BasePlugin{
		info: info,
		status: PluginStatus{
			State:      PluginStateIdle,
			LastUpdate: time.Now(),
		},
	}
}

// GetInfo 获取插件信息
func (p *BasePlugin) GetInfo() PluginInfo {
	return p.info
}

// GetStatus 获取插件状态
func (p *BasePlugin) GetStatus() PluginStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

// SetStatus 设置插件状态
func (p *BasePlugin) SetStatus(state PluginState, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	now := time.Now()
	p.status.State = state
	p.status.Error = err
	p.status.LastUpdate = now
	
	if state == PluginStateRunning && p.status.StartTime == nil {
		p.status.StartTime = &now
	}
}

// Init 初始化插件（默认实现）
func (p *BasePlugin) Init(ctx context.Context, config map[string]interface{}) error {
	p.SetStatus(PluginStateIdle, nil)
	return nil
}

// Start 启动插件（默认实现）
func (p *BasePlugin) Start(ctx context.Context) error {
	p.SetStatus(PluginStateRunning, nil)
	return nil
}

// Stop 停止插件（默认实现）
func (p *BasePlugin) Stop(ctx context.Context) error {
	p.SetStatus(PluginStateStopped, nil)
	return nil
}

// HealthCheck 健康检查（默认实现）
func (p *BasePlugin) HealthCheck(ctx context.Context) error {
	status := p.GetStatus()
	if status.State == PluginStateError {
		return fmt.Errorf("插件处于错误状态: %v", status.Error)
	}
	return nil
}