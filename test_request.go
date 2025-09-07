package main

import (
	"fmt"
	"log"
	"alpha_wx_notify/internal"
)

func main() {
	fmt.Println("开始测试API请求...")

	// 加载配置
	cfg, err := internal.LoadConfig("config/config.json")
	if err != nil {
		log.Printf("加载配置失败: %v\n", err)
		// 即使配置加载失败，也继续测试API请求
	}

	// 创建空投服务实例
	airdropService := internal.NewAirdropService(cfg)

	// 获取空投数据
	apiResp := airdropService.GetAirdropData()
	if apiResp == nil {
		fmt.Println("获取空投数据失败")
		return
	}

	// 请求成功，打印数据
	fmt.Println("请求成功！获取到空投数据:")
	fmt.Printf("总共获取到 %d 个空投项目\n", len(apiResp.Airdrops))

	// 打印前5个项目的信息作为示例
	fmt.Println("\n前5个项目示例:")
	count := 0
	for _, item := range apiResp.Airdrops {
		if count >= 5 {
			break
		}
		fmt.Printf("项目: %s(%s), 日期: %s, 时间: %s, 数量: %s, 阶段: %d\n",
			item.Token, item.Name, item.Date, item.Time, item.Amount, item.Phase)
		count++
	}

	fmt.Println("\n测试完成")
}