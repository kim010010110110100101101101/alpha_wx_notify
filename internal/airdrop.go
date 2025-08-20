package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Airdrop 空投结构体
type Airdrop struct {
	Token           string `json:"token"`
	Name            string `json:"name"`
	Date            string `json:"date"`
	Time            string `json:"time"`
	Points          string `json:"points"`
	Amount          string `json:"amount"`
	Type            string `json:"type"`
	Phase           int    `json:"phase"`
	Status          string `json:"status"`
	SystemTimestamp int64  `json:"system_timestamp"`
	Completed       bool   `json:"completed"`
	ContractAddress string `json:"contract_address"`
	ChainID         string `json:"chain_id"`
}

// ApiResponse API响应结构体
type ApiResponse struct {
	Airdrops []Airdrop `json:"airdrops"`
}

// SnapshotItem 快照项结构体
type SnapshotItem struct {
	Token  string
	Name   string
	Date   string
	Time   string
	Amount string
	Phase  int
}

// AirdropService 空投服务
type AirdropService struct {
	config *Config
}

// NewAirdropService 创建空投服务实例
func NewAirdropService(config *Config) *AirdropService {
	return &AirdropService{
		config: config,
	}
}

// GetAirdropData 获取空投数据
func (s *AirdropService) GetAirdropData() *ApiResponse {
	// 使用当前时间戳避免缓存
	url := fmt.Sprintf("https://alpha123.uk/api/data?t=%d&fresh=1", time.Now().UnixMilli())

	// 重试机制
	for attempt := 1; attempt <= 3; attempt++ {
		fmt.Printf("尝试第 %d 次请求...\n", attempt)
		log.Printf("请求地址: %s", url)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Println(err)
			continue
		}

		// 设置更完整的浏览器请求头
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("DNT", "1")
		req.Header.Set("Origin", "https://alpha123.uk")
		req.Header.Set("Pragma", "no-cache")
		req.Header.Set("Referer", "https://alpha123.uk/")
		req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
		req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		req.Header.Set("Sec-Ch-Ua-Platform", `"Linux"`)
		req.Header.Set("Sec-Fetch-Dest", "empty")
		req.Header.Set("Sec-Fetch-Mode", "cors")
		req.Header.Set("Sec-Fetch-Site", "same-origin")
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")

		// 设置超时时间
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("请求失败 (尝试 %d/3): %v\n", attempt, err)
			if attempt < 3 {
				// 随机延迟 2-5 秒后重试
				delay := time.Duration(2+attempt) * time.Second
				fmt.Printf("等待 %v 后重试...\n", delay)
				time.Sleep(delay)
			}
			continue
		}
		defer resp.Body.Close()

		body, err := readResponseBody(resp)
		if err != nil {
			log.Printf("读取响应体失败 (尝试 %d/3): %v\n", attempt, err)
			if attempt < 3 {
				delay := time.Duration(2+attempt) * time.Second
				fmt.Printf("等待 %v 后重试...\n", delay)
				time.Sleep(delay)
			}
			continue
		}

		// 打印响应状态码和前100个字符用于调试
		fmt.Printf("HTTP状态码: %d\n", resp.StatusCode)
		fmt.Printf("响应内容: %s\n", string(body))

		// 检查HTTP状态码
		if resp.StatusCode == 403 {
			fmt.Printf("遇到403错误，可能被反爬虫拦截 (尝试 %d/3)\n", attempt)
			if attempt < 3 {
				// 403错误时延迟更长时间
				delay := time.Duration(5+attempt*2) * time.Second
				fmt.Printf("等待 %v 后重试...\n", delay)
				time.Sleep(delay)
			}
			continue
		}

		if resp.StatusCode != 200 {
			fmt.Printf("API请求失败，状态码: %d (尝试 %d/3)\n", resp.StatusCode, attempt)
			if attempt < 3 {
				delay := time.Duration(2+attempt) * time.Second
				fmt.Printf("等待 %v 后重试...\n", delay)
				time.Sleep(delay)
			}
			continue
		}

		var apiResp ApiResponse
		err = json.Unmarshal(body, &apiResp)
		if err != nil {
			log.Printf("JSON解析失败 (尝试 %d/3): %v\n", attempt, err)
			log.Println("响应内容:", string(body))
			if attempt < 3 {
				delay := time.Duration(2+attempt) * time.Second
				fmt.Printf("等待 %v 后重试...\n", delay)
				time.Sleep(delay)
			}
			continue
		}

		// 成功获取数据，处理时间偏移
		for i, item := range apiResp.Airdrops {
			if item.Phase == 2 && item.Date != "" && item.Time != "" {
				// 时间加18小时
				layout := "2006-01-02 15:04"
				parsed, err := time.Parse(layout, item.Date+" "+item.Time)
				if err == nil {
					parsed = parsed.Add(18 * time.Hour)
					item.Date = parsed.Format("2006-01-02")
					item.Time = parsed.Format("15:04")
				}
				apiResp.Airdrops[i] = item
			}
		}

		fmt.Println("API请求成功！")
		return &apiResp
	}

	// 所有重试都失败了
	fmt.Println("所有重试都失败，无法获取空投数据")
	return nil
}

// FetchTokenPrice 获取token单价
func (s *AirdropService) FetchTokenPrice(token string) (float64, error) {
	url := fmt.Sprintf("https://alpha123.uk/api/price/%s?t=%d&fresh=1", token, time.Now().UnixMilli())

	// 重试机制
	for attempt := 1; attempt <= 2; attempt++ {
		log.Printf("价格API请求地址: %s", url)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return 0, err
		}

		// 设置完整的浏览器请求头
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("DNT", "1")
		req.Header.Set("Origin", "https://alpha123.uk")
		req.Header.Set("Pragma", "no-cache")
		req.Header.Set("Referer", "https://alpha123.uk/")
		req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
		req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		req.Header.Set("Sec-Ch-Ua-Platform", `"Linux"`)
		req.Header.Set("Sec-Fetch-Dest", "empty")
		req.Header.Set("Sec-Fetch-Mode", "cors")
		req.Header.Set("Sec-Fetch-Site", "same-origin")
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")

		client := &http.Client{
			Timeout: 15 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			if attempt < 2 {
				time.Sleep(time.Duration(2+attempt) * time.Second)
				continue
			}
			return 0, err
		}
		defer resp.Body.Close()

		// 检查状态码
		if resp.StatusCode == 403 {
			if attempt < 2 {
				time.Sleep(time.Duration(3+attempt) * time.Second)
				continue
			}
			return 0, fmt.Errorf("price API blocked (403)")
		}

		if resp.StatusCode != 200 {
			if attempt < 2 {
				time.Sleep(time.Duration(2+attempt) * time.Second)
				continue
			}
			return 0, fmt.Errorf("price API failed with status %d", resp.StatusCode)
		}

		// 读取响应体（处理gzip压缩）
		body, err := readResponseBody(resp)
		if err != nil {
			if attempt < 2 {
				time.Sleep(time.Duration(2+attempt) * time.Second)
				continue
			}
			return 0, fmt.Errorf("failed to read response body: %v", err)
		}

		var result struct {
			Success bool    `json:"success"`
			Price   float64 `json:"price"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			if attempt < 2 {
				time.Sleep(time.Duration(2+attempt) * time.Second)
				continue
			}
			return 0, fmt.Errorf("failed to parse JSON: %v, body: %s", err, string(body))
		}

		if !result.Success {
			return 0, fmt.Errorf("price fetch failed")
		}
		return result.Price, nil
	}

	return 0, fmt.Errorf("all price fetch attempts failed")
}

// GenerateMessageAndSnapshot 生成消息和快照
func (s *AirdropService) GenerateMessageAndSnapshot() (string, string) {
	fmt.Printf("今日日期: %s\n", time.Now().Format("2006-01-02"))

	apiResp := s.GetAirdropData()
	if apiResp == nil {
		fmt.Println("获取空投数据失败")
		return "", ""
	}

	// 收集符合条件的快照项
	var snapshotItems []SnapshotItem
	var validAirdrops []Airdrop

	for _, item := range apiResp.Airdrops {
		// 检查日期是否在今天往后3天内
		today := time.Now()
		itemDate, err := time.Parse("2006-01-02", item.Date)
		if err != nil {
			fmt.Printf("解析日期失败: %v\n", err)
			continue
		}

		// 计算日期差（天数）
		daysDiff := int(itemDate.Sub(today.Truncate(24*time.Hour)).Hours() / 24)

		// 只包含今天往后3天内的项目（今天=0，明天=1，后天=2，大后天=3）
		if daysDiff < 0 || daysDiff > 3 {
			continue
		}

		if s.config.FiterTge && item.Type == "tge" {
			fmt.Printf("过滤TGE: %+v\n", item)
			continue
		}

		// 添加到快照项列表
		snapshotItems = append(snapshotItems, SnapshotItem{
			Token:  item.Token,
			Name:   item.Name,
			Date:   item.Date,
			Time:   item.Time,
			Amount: item.Amount,
			Phase:  item.Phase,
		})
		validAirdrops = append(validAirdrops, item)
	}

	if len(snapshotItems) == 0 {
		return "", ""
	}

	// 对快照项进行排序
	s.sortSnapshotItems(snapshotItems)

	// 生成消息和快照
	msg := "| 项目 | 时间 | 积分 | 数量 | 阶段 | 价格(USD) |\n|---|---|---|---|---|---|\n"

	for _, snapshotItem := range snapshotItems {
		// 找到对应的airdrop项目来获取价格信息
		var correspondingAirdrop *Airdrop
		for _, airdrop := range validAirdrops {
			if airdrop.Token == snapshotItem.Token && airdrop.Date == snapshotItem.Date &&
				airdrop.Time == snapshotItem.Time && airdrop.Phase == snapshotItem.Phase {
				correspondingAirdrop = &airdrop
				break
			}
		}

		if correspondingAirdrop == nil {
			continue
		}

		var amount int
		if snapshotItem.Amount == "" {
			amount = 0
		} else {
			var err error
			amount, err = strconv.Atoi(snapshotItem.Amount)
			if err != nil {
				fmt.Printf("转换数量失败: %v\n", err)
				amount = 0
			}
		}

		price, err := s.FetchTokenPrice(snapshotItem.Token)
		if err != nil {
			fmt.Printf("获取%s价格失败: %v\n", snapshotItem.Token, err)
			price = 0
		}

		// 如果type是tge，在名字后面加上(tge)
		projectName := snapshotItem.Name
		if correspondingAirdrop.Type == "tge" {
			projectName += "(tge)"
		}

		msg += fmt.Sprintf("| %s(%s) | %s %s | %s | %s | %d | %.2f |\n",
			snapshotItem.Token, projectName, snapshotItem.Date, snapshotItem.Time,
			correspondingAirdrop.Points, snapshotItem.Amount, snapshotItem.Phase, price*float64(amount))
	}

	// 生成排序后的快照字符串
	snapshot := s.itemsToSnapshot(snapshotItems)

	return msg, snapshot
}

// 解析快照字符串为结构体切片
func (s *AirdropService) parseSnapshot(snapshot string) []SnapshotItem {
	var items []SnapshotItem
	lines := strings.Split(strings.TrimSpace(snapshot), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 6 {
			phase, _ := strconv.Atoi(parts[5])
			items = append(items, SnapshotItem{
				Token:  parts[0],
				Name:   parts[1],
				Date:   parts[2],
				Time:   parts[3],
				Amount: parts[4],
				Phase:  phase,
			})
		}
	}
	return items
}

// 将结构体切片转换为快照字符串
func (s *AirdropService) itemsToSnapshot(items []SnapshotItem) string {
	var lines []string
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("%s|%s|%s|%s|%s|%d",
			item.Token, item.Name, item.Date, item.Time, item.Amount, item.Phase))
	}
	return strings.Join(lines, "\n")
}

// 对快照项进行排序：先按时间排序（时间靠近的在上面），再按字母排序
func (s *AirdropService) sortSnapshotItems(items []SnapshotItem) {
	sort.Slice(items, func(i, j int) bool {
		// 解析时间
		timeI := s.parseDateTime(items[i].Date, items[i].Time)
		timeJ := s.parseDateTime(items[j].Date, items[j].Time)

		// 如果时间不同，按时间排序（时间靠近的在上面）
		if !timeI.Equal(timeJ) {
			return timeI.Before(timeJ)
		}

		// 时间相同时，按Token字母排序
		return items[i].Token < items[j].Token
	})
}

// 解析日期时间字符串
func (s *AirdropService) parseDateTime(date, timeStr string) time.Time {
	if date == "" {
		return time.Time{}
	}
	dateTimeStr := date
	if timeStr != "" {
		dateTimeStr += " " + timeStr
		layout := "2006-01-02 15:04"
		if parsed, err := time.Parse(layout, dateTimeStr); err == nil {
			return parsed
		}
	}
	layout := "2006-01-02"
	if parsed, err := time.Parse(layout, date); err == nil {
		return parsed
	}
	return time.Time{}
}

// 比较两个快照是否相同（忽略顺序）
func (s *AirdropService) CompareSnapshots(snapshot1, snapshot2 string) bool {
	items1 := s.parseSnapshot(snapshot1)
	items2 := s.parseSnapshot(snapshot2)

	// 如果数量不同，直接返回false
	if len(items1) != len(items2) {
		return false
	}

	// 创建map来统计每个项目的出现次数
	count1 := make(map[string]int)
	count2 := make(map[string]int)

	for _, item := range items1 {
		key := fmt.Sprintf("%s|%s|%s|%s|%s|%d", item.Token, item.Name, item.Date, item.Time, item.Amount, item.Phase)
		count1[key]++
	}

	for _, item := range items2 {
		key := fmt.Sprintf("%s|%s|%s|%s|%s|%d", item.Token, item.Name, item.Date, item.Time, item.Amount, item.Phase)
		count2[key]++
	}

	// 比较两个map是否相同
	for key, count := range count1 {
		if count2[key] != count {
			return false
		}
	}

	return true
}

// 检测快照变化类型
func (s *AirdropService) DetectSnapshotChange(oldSnapshot, newSnapshot string) (bool, bool) {
	oldItems := s.parseSnapshot(oldSnapshot)
	newItems := s.parseSnapshot(newSnapshot)

	// 创建map来快速查找
	oldMap := make(map[string]bool)
	newMap := make(map[string]bool)

	for _, item := range oldItems {
		key := fmt.Sprintf("%s|%s|%s|%s|%s|%d", item.Token, item.Name, item.Date, item.Time, item.Amount, item.Phase)
		oldMap[key] = true
	}

	for _, item := range newItems {
		key := fmt.Sprintf("%s|%s|%s|%s|%s|%d", item.Token, item.Name, item.Date, item.Time, item.Amount, item.Phase)
		newMap[key] = true
	}

	// 检查是否有新增项目
	hasAddition := false
	for key := range newMap {
		if !oldMap[key] {
			hasAddition = true
			break
		}
	}

	// 检查是否只有删除（没有新增）
	isOnlyDeletion := !hasAddition && len(newItems) < len(oldItems)

	return hasAddition, isOnlyDeletion
}
