package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"recommendation-system/internal/infra/di"
)

func main() {
	fmt.Println("推荐系统框架已启动")
	
	// 使用Wire初始化应用程序
	app, err := di.InitializeApp()
	if err != nil {
		log.Fatalf("初始化应用程序失败: %v", err)
	}
	
	fmt.Println("应用程序初始化成功")
	
	// 创建上下文用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// 启动应用程序
	go func() {
		if err := startApp(ctx, app); err != nil {
			log.Printf("应用程序运行错误: %v", err)
			cancel()
		}
	}()
	
	// 等待信号或错误
	select {
	case sig := <-sigChan:
		fmt.Printf("接收到信号: %v，正在关闭应用...\n", sig)
	case <-ctx.Done():
		fmt.Println("应用程序上下文已取消")
	}
	
	// 优雅关闭
	shutdownApp(app)
	fmt.Println("应用程序已关闭")
}

// startApp 启动应用程序
func startApp(ctx context.Context, app *di.Application) error {
	app.Logger.Info("开始启动推荐系统框架")
	
	// 加载配置
	configPath := "configs/development/config.yaml"
	if err := app.ConfigManager.Load(configPath); err != nil {
		return fmt.Errorf("加载配置文件失败: %w", err)
	}
	
	app.Logger.Info("配置加载成功")
	
	// 验证配置
	if err := app.ConfigManager.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}
	
	app.Logger.Info("配置验证成功")
	
	// 插件管理器将在后续版本中实现
	app.Logger.Info("插件系统准备就绪")
	
	// 启动推荐服务
	app.Logger.Info("推荐系统框架启动完成，等待请求...")
	
	// 模拟服务运行
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			app.Logger.Debug("推荐系统运行中...")
		}
	}
}

// shutdownApp 关闭应用程序
func shutdownApp(app *di.Application) {
	app.Logger.Info("开始关闭应用程序")
	
	// 插件关闭将在后续版本中实现
	app.Logger.Info("插件系统关闭完成")
	
	app.Logger.Info("应用程序关闭完成")
}