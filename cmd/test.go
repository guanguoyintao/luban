package main

import (
	"context"
	"fmt"
	"log"
	
	"recommendation-system/internal/infra/di"
)

func main() {
	fmt.Println("=== 推荐系统框架测试 ===")
	
	// 初始化应用程序
	app, err := di.InitializeApp()
	if err != nil {
		log.Fatalf("初始化应用程序失败: %v", err)
	}
	
	fmt.Println("应用程序初始化成功")
	
	// 测试推荐功能
	ctx := context.Background()
	userID := "user_123"
	count := 5
	
	fmt.Printf("\n为用户 %s 生成 %d 个推荐...\n", userID, count)
	
	recommendations, err := app.RecommendationSvc.GetRecommendations(ctx, userID, count)
	if err != nil {
		log.Printf("获取推荐失败: %v", err)
		return
	}
	
	fmt.Printf("成功生成 %d 个推荐:\n", len(recommendations))
	for i, rec := range recommendations {
		fmt.Printf("%d. 物品ID: %s, 得分: %.2f, 算法: %s, 原因: %s\n", 
			i+1, rec.ItemID, rec.Score, rec.Algorithm, rec.Reason)
	}
	
	// 测试按类别推荐
	category := "technology"
	fmt.Printf("\n为用户 %s 生成 %s 类别的推荐...\n", userID, category)
	
	categoryRecommendations, err := app.RecommendationSvc.GetRecommendationsByCategory(ctx, userID, category, 3)
	if err != nil {
		log.Printf("获取类别推荐失败: %v", err)
		return
	}
	
	fmt.Printf("成功生成 %d 个 %s 类别推荐:\n", len(categoryRecommendations), category)
	for i, rec := range categoryRecommendations {
		fmt.Printf("%d. 物品ID: %s, 得分: %.2f, 算法: %s\n", 
			i+1, rec.ItemID, rec.Score, rec.Algorithm)
	}
	
	fmt.Println("\n=== 测试完成 ===")
}