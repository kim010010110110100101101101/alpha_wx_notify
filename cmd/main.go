// Package main 是程序的入口包，负责启动空投信息检查和推送服务
// 该程序主要功能是定期检查空投信息，当有新的空投信息时，通过Server酱推送通知
package main

import (
	"fmt"
	"log"
	"time"
	"alpha_wx_notify/internal" // 导入内部包，包含空投服务和工具函数
)

// ProcessAirdrops 处理空投信息的主要逻辑
// 该函数负责:
// 1. 加载配置文件
// 2. 创建空投服务实例
// 3. 获取并生成空投消息和快照
// 4. 检测空投信息变化并决定是否推送通知
// 5. 保存当前快照以便下次比较
func ProcessAirdrops() {
	fmt.Printf("[%s] 开始检查空投信息...\n", time.Now().Format("2006-01-02 15:04:05"))

	// 加载配置文件
	// 配置文件包含Server酱的SendKey、检查间隔和是否过滤TGE项目等设置
	cfg, err := internal.LoadConfig("../config/config.json")
	if err != nil {
		log.Fatal(err) // 如果配置加载失败，程序无法继续运行，直接终止
	}

	// 创建空投服务实例
	// 空投服务负责获取空投数据、生成消息和快照、比较快照等核心功能
	airdropService := internal.NewAirdropService(cfg)

	// 生成消息和快照
	// msg: 格式化的消息内容，用于推送通知
	// snapshot: 当前空投信息的快照，用于与上次快照比较检测变化
	msg, snapshot := airdropService.GenerateMessageAndSnapshot()

	if msg != "" { // 如果有空投信息（消息不为空）
		// 读取上次保存的快照文件
		// 快照文件记录了上次检查时的空投信息，用于与当前信息比较
		lastSnapshot, err := internal.LoadLastSnapshot("../data/last_snapshot.txt")
		if err != nil {
			fmt.Printf("读取上次快照失败: %v\n", err) // 读取失败时记录错误但继续执行
		}

		// 使用新的对比函数来忽略顺序比较两个快照是否相同
		// 如果快照不同，说明空投信息有变化
		if !airdropService.CompareSnapshots(snapshot, lastSnapshot) {
			// 检测变化类型：是新增了项目还是只是删除了项目
			// isOnlyDeletion为true表示只有删除操作，没有新增项目
			_, isOnlyDeletion := airdropService.DetectSnapshotChange(lastSnapshot, snapshot)

			if isOnlyDeletion {
				// 如果只是删除了项目，不进行推送，只更新快照
				fmt.Println("检测到空投信息删除，不进行推送，仅更新快照...")
				// 保存当前快照但不推送通知
				if err := internal.SaveSnapshot(snapshot, "../data/last_snapshot.txt"); err != nil {
					fmt.Printf("保存快照失败: %v\n", err)
				}
			} else {
				// 如果有新增项目或其他变化，推送通知
				fmt.Println("检测到空投信息变化，推送通知...")
				fmt.Println(msg) // 打印消息内容用于调试

				// 通过Server酱推送通知
				// 标题固定为"今日空投播报"
				if err := internal.SendToServerChan(msg, "今日空投播报", cfg); err != nil {
					fmt.Println("推送Server酱失败:", err)
				} else {
					fmt.Println("推送成功！")
				}

				// 保存当前快照，用于下次比较
				if err := internal.SaveSnapshot(snapshot, "../data/last_snapshot.txt"); err != nil {
					fmt.Printf("保存快照失败: %v\n", err)
				}
			}
		} else {
			// 如果快照相同，说明空投信息没有变化，不需要推送
			fmt.Println("空投信息无变化，跳过推送。")
		}
	} else {
		// 如果没有空投信息（消息为空）
		fmt.Println("今日无空投信息。")
		
		// 如果当前没有空投信息，但之前有，需要清空快照文件
		// 这样可以避免下次检查时与空的当前状态比较导致误判
		lastSnapshot, err := internal.LoadLastSnapshot("../data/last_snapshot.txt")
		if err == nil && lastSnapshot != "" { // 如果上次快照存在且不为空
			fmt.Println("清空快照文件...")
			// 写入空字符串到快照文件，相当于清空文件
			if err := internal.SaveSnapshot("", "../data/last_snapshot.txt"); err != nil {
				fmt.Printf("清空快照失败: %v\n", err)
			}
		}
	}
}

// main 程序入口函数
// 调用ProcessAirdrops函数开始处理空投信息
// 添加测试模式，直接输出API请求结果，验证请求头修改是否有效
func main() {
	// 测试模式：直接获取API数据并输出结果，验证请求头修改是否有效
	fmt.Println("=== 测试模式：验证API请求 ===")
	
	// 加载配置
	cfg, err := internal.LoadConfig("../config/config.json")
	if err != nil {
		log.Printf("加载配置失败: %v\n", err)
		// 即使配置加载失败，也继续测试API请求
	}

	// 创建空投服务实例
	airdropService := internal.NewAirdropService(cfg)

	// 获取空投数据
	apiResp := airdropService.GetAirdropData()
	if apiResp == nil {
		fmt.Println("获取空投数据失败，请求可能仍然返回403错误")
		return
	} else {
		fmt.Println("成功获取API响应，状态正常")
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

	fmt.Println("\n测试完成，请求头修改有效！")
	
	// 注释掉正常处理流程，仅进行测试
	// 记录开始时间，便于调试和性能分析
	// start := time.Now()
	
	// 调用主要处理函数
	// ProcessAirdrops()
	
	// 记录执行耗时
	// fmt.Printf("处理完成，耗时: %v\n", time.Since(start))
}
