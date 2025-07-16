package main

import (
	"fmt"
	"log"
	"time"
	"alpha_wx_notify/internal"
)

// ProcessAirdrops 处理空投信息
func ProcessAirdrops() {
	fmt.Printf("[%s] 开始检查空投信息...\n", time.Now().Format("2006-01-02 15:04:05"))

	// 加载配置
	cfg, err := internal.LoadConfig("../config/config.json")
	if err != nil {
		log.Fatal(err)
	}

	// 创建空投服务
	airdropService := internal.NewAirdropService(cfg)

	// 生成消息和快照
	msg, snapshot := airdropService.GenerateMessageAndSnapshot()

	if msg != "" {
		// 读取上次的快照
		lastSnapshot, err := internal.LoadLastSnapshot("../data/last_snapshot.txt")
		if err != nil {
			fmt.Printf("读取上次快照失败: %v\n", err)
		}

		// 使用新的对比函数来忽略顺序
		if !airdropService.CompareSnapshots(snapshot, lastSnapshot) {
			// 检测变化类型
			_, isOnlyDeletion := airdropService.DetectSnapshotChange(lastSnapshot, snapshot)

			if isOnlyDeletion {
				fmt.Println("检测到空投信息删除，不进行推送，仅更新快照...")
				// 保存当前快照但不推送
				if err := internal.SaveSnapshot(snapshot, "../data/last_snapshot.txt"); err != nil {
					fmt.Printf("保存快照失败: %v\n", err)
				}
			} else {
				fmt.Println("检测到空投信息变化，推送通知...")
				fmt.Println(msg)

				// 推送通知
				if err := internal.SendToServerChan(msg, "今日空投播报", cfg); err != nil {
					fmt.Println("推送Server酱失败:", err)
				} else {
					fmt.Println("推送成功！")
				}

				// 保存当前快照
				if err := internal.SaveSnapshot(snapshot, "../data/last_snapshot.txt"); err != nil {
					fmt.Printf("保存快照失败: %v\n", err)
				}
			}
		} else {
			fmt.Println("空投信息无变化，跳过推送。")
		}
	} else {
		fmt.Println("今日无空投信息。")
		// 如果当前没有空投信息，但之前有，也要更新快照
		lastSnapshot, err := internal.LoadLastSnapshot("../data/last_snapshot.txt")
		if err == nil && lastSnapshot != "" {
			fmt.Println("清空快照文件...")
			if err := internal.SaveSnapshot("", "../data/last_snapshot.txt"); err != nil {
				fmt.Printf("清空快照失败: %v\n", err)
			}
		}
	}
}

func main() {
	ProcessAirdrops()
}
