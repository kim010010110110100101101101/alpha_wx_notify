// Package internal 包含项目的核心功能实现
// 该包提供空投数据获取、处理和通知推送等功能
package internal

import (
	"encoding/json" // 用于JSON编解码
	"fmt"           // 用于格式化输出
	"log"           // 用于日志记录
	"net/http"      // 用于HTTP请求
	"sort"          // 用于排序
	"strconv"       // 用于字符串转换
	"strings"       // 用于字符串处理
	"time"          // 用于时间处理
)

// Airdrop 空投结构体，用于存储从API获取的空投信息
// 包含项目名称、代币、日期、时间、数量等关键信息
type Airdrop struct {
	Token           string `json:"token"`           // 代币符号
	Name            string `json:"name"`            // 项目名称
	Date            string `json:"date"`            // 空投日期，格式为YYYY-MM-DD
	Time            string `json:"time"`            // 空投时间，格式为HH:MM
	Points          string `json:"points"`          // 所需积分
	Amount          string `json:"amount"`          // 空投数量
	Type            string `json:"type"`            // 空投类型，如"airdrop"或"tge"
	Phase           int    `json:"phase"`           // 空投阶段
	Status          string `json:"status"`          // 空投状态
	SystemTimestamp int64  `json:"system_timestamp"` // 系统时间戳
	Completed       bool   `json:"completed"`       // 是否已完成
	ContractAddress string `json:"contract_address"` // 合约地址
	ChainID         string `json:"chain_id"`         // 链ID
}

// ApiResponse API响应结构体，包含空投列表
// 用于解析从API获取的JSON响应
type ApiResponse struct {
	Airdrops []Airdrop `json:"airdrops"` // 空投列表
}

// SnapshotItem 快照项结构体，用于存储空投信息的简化版本
// 用于生成和比较快照，只保留关键信息
type SnapshotItem struct {
	Token  string // 代币符号
	Name   string // 项目名称
	Date   string // 空投日期
	Time   string // 空投时间
	Amount string // 空投数量
	Phase  int    // 空投阶段
}

// AirdropService 空投服务，提供空投数据处理的核心功能
// 包括获取数据、生成消息、比较快照等
type AirdropService struct {
	config *Config // 配置信息，包含SendKey、检查间隔等
}

// NewAirdropService 创建空投服务实例
// 参数:
//   - config: 配置信息，包含SendKey、检查间隔等
// 返回:
//   - *AirdropService: 空投服务实例
func NewAirdropService(config *Config) *AirdropService {
	return &AirdropService{
		config: config,
	}
}

// GetAirdropData 获取空投数据
// 该方法从API获取最新的空投信息，包含重试机制和错误处理
// 返回:
//   - *ApiResponse: 包含空投列表的API响应，如果获取失败则返回nil
func (s *AirdropService) GetAirdropData() *ApiResponse {
	// 使用当前时间戳作为URL参数避免缓存
	url := fmt.Sprintf("https://alpha123.uk/api/data?t=%d&fresh=1", time.Now().UnixMilli())

	// 重试机制，最多尝试3次
	for attempt := 1; attempt <= 3; attempt++ {
		fmt.Printf("尝试第 %d 次请求...\n", attempt)
		log.Printf("请求地址: %s", url)

		// 创建HTTP请求
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Println(err) // 记录错误但继续尝试
			continue
		}

		// 设置更完整的浏览器请求头，模拟真实浏览器请求
		// 这些请求头有助于绕过一些反爬虫措施
		req.Header.Set("Accept", "application/json, text/plain, */*") // 接受JSON和文本响应
		req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7") // 语言偏好
		req.Header.Set("Accept-Encoding", "gzip, deflate, br") // 支持的压缩方式
		req.Header.Set("Cache-Control", "no-cache") // 禁用缓存
		req.Header.Set("Connection", "keep-alive") // 保持连接
		req.Header.Set("DNT", "1") // 请求不跟踪
		req.Header.Set("Origin", "https://alpha123.uk") // 请求来源
		req.Header.Set("Pragma", "no-cache") // 禁用缓存
		req.Header.Set("Referer", "https://alpha123.uk/") // 引用页
		req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`) // 浏览器信息
		req.Header.Set("Sec-Ch-Ua-Mobile", "?0") // 非移动设备
		req.Header.Set("Sec-Ch-Ua-Platform", `"Linux"`) // 操作系统平台
		req.Header.Set("Sec-Fetch-Dest", "empty") // 请求目标
		req.Header.Set("Sec-Fetch-Mode", "cors") // 请求模式
		req.Header.Set("Sec-Fetch-Site", "same-origin") // 请求站点
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36") // 用户代理
		req.Header.Set("X-Requested-With", "XMLHttpRequest") // XHR请求标识

		// 设置HTTP客户端，包括30秒超时时间
		// 超时设置可以防止请求长时间挂起
		client := &http.Client{
			Timeout: 30 * time.Second, // 30秒超时，避免请求长时间挂起
		}

		// 执行HTTP请求
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("请求失败 (尝试 %d/3): %v\n", attempt, err)
			if attempt < 3 { // 如果不是最后一次尝试，则延迟后重试
				// 随机延迟 2-5 秒后重试，延迟时间随尝试次数增加
				delay := time.Duration(2+attempt) * time.Second
				fmt.Printf("等待 %v 后重试...\n", delay)
				time.Sleep(delay)
			}
			continue // 继续下一次尝试
		}
		defer resp.Body.Close() // 确保响应体被关闭

		// 读取响应体，处理可能的gzip压缩
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

		// 打印响应状态码和响应内容用于调试
		fmt.Printf("HTTP状态码: %d\n", resp.StatusCode)
		fmt.Printf("响应内容: %s\n", string(body))

		// 检查HTTP状态码
		if resp.StatusCode == 403 { // 403表示禁止访问，可能是被反爬虫机制拦截
			fmt.Printf("遇到403错误，可能被反爬虫拦截 (尝试 %d/3)\n", attempt)
			if attempt < 3 {
				// 403错误时延迟更长时间，给服务器更多冷却时间
				delay := time.Duration(5+attempt*2) * time.Second
				fmt.Printf("等待 %v 后重试...\n", delay)
				time.Sleep(delay)
			}
			continue
		}

		// 处理其他非200状态码
		if resp.StatusCode != 200 { // 200表示请求成功
			fmt.Printf("API请求失败，状态码: %d (尝试 %d/3)\n", resp.StatusCode, attempt)
			if attempt < 3 {
				delay := time.Duration(2+attempt) * time.Second
				fmt.Printf("等待 %v 后重试...\n", delay)
				time.Sleep(delay)
			}
			continue
		}

		// 解析JSON响应
		var apiResp ApiResponse
		err = json.Unmarshal(body, &apiResp)
		if err != nil { // JSON解析失败
			log.Printf("JSON解析失败 (尝试 %d/3): %v\n", attempt, err)
			log.Println("响应内容:", string(body)) // 打印完整响应内容以便调试
			if attempt < 3 {
				delay := time.Duration(2+attempt) * time.Second
				fmt.Printf("等待 %v 后重试...\n", delay)
				time.Sleep(delay)
			}
			continue
		}

		// 成功获取数据，处理时间偏移
		// 对于Phase 2的项目，需要将时间加上18小时进行调整
		for i, item := range apiResp.Airdrops {
			if item.Phase == 2 && item.Date != "" && item.Time != "" {
				// 时间加18小时，这是针对特定阶段的时间调整
				layout := "2006-01-02 15:04" // Go的时间格式化模板
				parsed, err := time.Parse(layout, item.Date+" "+item.Time)
				if err == nil {
					// 时间调整，加18小时
					parsed = parsed.Add(18 * time.Hour)
					// 更新日期和时间
					item.Date = parsed.Format("2006-01-02")
					item.Time = parsed.Format("15:04")
				}
				// 更新切片中的项目
				apiResp.Airdrops[i] = item
			}
		}

		fmt.Println("API请求成功！")
		return &apiResp // 返回成功获取的数据
	}

	// 所有重试都失败了
	fmt.Println("所有重试都失败，无法获取空投数据")
	return nil // 返回nil表示获取失败
}

// FetchTokenPrice 获取token单价
// 该方法从价格API获取指定代币的当前价格
// 参数:
//   - token: 代币符号，如"ETH"、"BTC"等
// 返回:
//   - float64: 代币价格，单位为USD
//   - error: 错误信息，如果获取成功则为nil
func (s *AirdropService) FetchTokenPrice(token string) (float64, error) {
	// 构建价格API的URL，添加时间戳参数避免缓存
	url := fmt.Sprintf("https://alpha123.uk/api/price/%s?t=%d&fresh=1", token, time.Now().UnixMilli())

	// 重试机制，最多尝试2次
	for attempt := 1; attempt <= 2; attempt++ {
		log.Printf("价格API请求地址: %s", url)
		// 创建HTTP GET请求
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return 0, err // 创建请求失败，直接返回错误
		}

		// 设置完整的浏览器请求头，模拟真实浏览器请求
		// 这些请求头有助于绕过一些反爬虫措施
		req.Header.Set("Accept", "application/json, text/plain, */*") // 接受JSON和文本响应
		req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7") // 语言偏好
		req.Header.Set("Accept-Encoding", "gzip, deflate, br") // 支持的压缩方式
		req.Header.Set("Cache-Control", "no-cache") // 禁用缓存
		req.Header.Set("Connection", "keep-alive") // 保持连接
		req.Header.Set("DNT", "1") // 请求不跟踪
		req.Header.Set("Origin", "https://alpha123.uk") // 请求来源
		req.Header.Set("Pragma", "no-cache") // 禁用缓存
		req.Header.Set("Referer", "https://alpha123.uk/") // 引用页
		req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`) // 浏览器信息
		req.Header.Set("Sec-Ch-Ua-Mobile", "?0") // 非移动设备
		req.Header.Set("Sec-Ch-Ua-Platform", `"Linux"`) // 操作系统平台
		req.Header.Set("Sec-Fetch-Dest", "empty") // 请求目标
		req.Header.Set("Sec-Fetch-Mode", "cors") // 请求模式
		req.Header.Set("Sec-Fetch-Site", "same-origin") // 请求站点
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36") // 用户代理
		req.Header.Set("X-Requested-With", "XMLHttpRequest") // XHR请求标识

		// 设置HTTP客户端，包括15秒超时时间
		// 价格API请求超时时间比空投数据API短，因为价格查询应该更快返回
		client := &http.Client{
			Timeout: 15 * time.Second, // 15秒超时
		}

		// 执行HTTP请求
		resp, err := client.Do(req)
		if err != nil { // 请求失败
			if attempt < 2 { // 如果不是最后一次尝试，则延迟后重试
				time.Sleep(time.Duration(2+attempt) * time.Second)
				continue
			}
			return 0, err // 所有尝试都失败，返回错误
		}
		defer resp.Body.Close() // 确保响应体被关闭

		// 检查HTTP状态码
		if resp.StatusCode == 403 { // 403表示禁止访问，可能是被反爬虫机制拦截
			if attempt < 2 {
				time.Sleep(time.Duration(3+attempt) * time.Second) // 403错误时延迟更长时间
				continue
			}
			return 0, fmt.Errorf("price API blocked (403)") // 返回被拦截错误
		}

		// 处理其他非200状态码
		if resp.StatusCode != 200 { // 200表示请求成功
			if attempt < 2 {
				time.Sleep(time.Duration(2+attempt) * time.Second)
				continue
			}
			return 0, fmt.Errorf("price API failed with status %d", resp.StatusCode) // 返回API失败错误
		}

		// 读取响应体，处理可能的gzip压缩
		body, err := readResponseBody(resp)
		if err != nil { // 读取响应体失败
			if attempt < 2 {
				time.Sleep(time.Duration(2+attempt) * time.Second)
				continue
			}
			return 0, fmt.Errorf("failed to read response body: %v", err) // 返回读取失败错误
		}

		// 定义匿名结构体用于解析价格API的JSON响应
		var result struct {
			Success bool    `json:"success"` // 是否成功
			Price   float64 `json:"price"`   // 价格值
		}
		
		// 解析JSON响应
		if err := json.Unmarshal(body, &result); err != nil { // JSON解析失败
			if attempt < 2 {
				time.Sleep(time.Duration(2+attempt) * time.Second)
				continue
			}
			return 0, fmt.Errorf("failed to parse JSON: %v, body: %s", err, string(body)) // 返回解析失败错误
		}

		// 检查API返回的成功标志
		if !result.Success { // API返回失败
			return 0, fmt.Errorf("price fetch failed") // 返回获取失败错误
		}
		return result.Price, nil // 返回成功获取的价格
	}

	// 所有重试都失败
	return 0, fmt.Errorf("all price fetch attempts failed")
}

// GenerateMessageAndSnapshot 生成消息和快照
// 该方法获取空投数据，过滤符合条件的项目，并生成消息和快照
// 返回:
//   - string: 格式化的消息内容，用于推送通知
//   - string: 当前空投信息的快照，用于与上次快照比较检测变化
func (s *AirdropService) GenerateMessageAndSnapshot() (string, string) {
	// 打印当前日期，便于日志跟踪
	fmt.Printf("今日日期: %s\n", time.Now().Format("2006-01-02"))

	// 获取空投数据
	apiResp := s.GetAirdropData()
	if apiResp == nil { // 如果获取失败，返回空字符串
		fmt.Println("获取空投数据失败")
		return "", ""
	}

	// 收集符合条件的快照项
	var snapshotItems []SnapshotItem // 用于生成快照的项目列表
	var validAirdrops []Airdrop     // 有效的空投项目列表

	// 遍历所有空投项目，筛选符合条件的项目
	for _, item := range apiResp.Airdrops {
		// 检查日期是否在今天往后3天内
		today := time.Now()
		// 解析项目日期
		itemDate, err := time.Parse("2006-01-02", item.Date)
		if err != nil {
			fmt.Printf("解析日期失败: %v\n", err) // 日期格式错误，跳过该项目
			continue
		}

		// 计算日期差（天数）
		// 将日期截断到天级别，然后计算相差的小时数，再除以24得到天数
		daysDiff := int(itemDate.Sub(today.Truncate(24*time.Hour)).Hours() / 24)

		// 只包含今天往后3天内的项目（今天=0，明天=1，后天=2，大后天=3）
		// 过滤掉过去的和3天后的项目
		if daysDiff < 0 || daysDiff > 3 {
			continue
		}

		// 如果配置了过滤TGE类型的项目，且当前项目是TGE类型，则跳过
		if s.config.FiterTge && item.Type == "tge" {
			fmt.Printf("过滤TGE: %+v\n", item) // 记录被过滤的TGE项目
			continue
		}

		// 将符合条件的项目添加到快照项列表
		snapshotItems = append(snapshotItems, SnapshotItem{
			Token:  item.Token,  // 代币符号
			Name:   item.Name,   // 项目名称
			Date:   item.Date,   // 空投日期
			Time:   item.Time,   // 空投时间
			Amount: item.Amount, // 空投数量
			Phase:  item.Phase,  // 空投阶段
		})
		// 同时保存完整的空投信息，用于后续获取价格等详细信息
		validAirdrops = append(validAirdrops, item)
	}

	// 如果没有符合条件的项目，返回空字符串
	if len(snapshotItems) == 0 {
		return "", ""
	}

	// 对快照项进行排序（按时间和代币名称）
	s.sortSnapshotItems(snapshotItems)

	// 生成消息和快照
	// 使用Markdown表格格式生成消息内容
	msg := "| 项目 | 时间 | 积分 | 数量 | 阶段 | 价格(USD) |\n|---|---|---|---|---|---|\n"

	// 遍历排序后的快照项，生成消息内容
	for _, snapshotItem := range snapshotItems {
		// 找到对应的airdrop项目来获取价格信息和积分等详细信息
		var correspondingAirdrop *Airdrop
		for _, airdrop := range validAirdrops {
			// 通过Token、日期、时间和阶段匹配确保找到正确的项目
			if airdrop.Token == snapshotItem.Token && airdrop.Date == snapshotItem.Date &&
				airdrop.Time == snapshotItem.Time && airdrop.Phase == snapshotItem.Phase {
				correspondingAirdrop = &airdrop
				break
			}
		}

		// 如果找不到对应的完整空投信息，跳过该项目
		if correspondingAirdrop == nil {
			continue
		}

		// 解析空投数量为整数，用于计算价值
		var amount int
		if snapshotItem.Amount == "" { // 如果数量为空，设为0
			amount = 0
		} else {
			var err error
			// 将字符串数量转换为整数
			amount, err = strconv.Atoi(snapshotItem.Amount)
			if err != nil {
				fmt.Printf("转换数量失败: %v\n", err)
				amount = 0 // 转换失败时设为0
			}
		}

		// 获取代币价格
		price, err := s.FetchTokenPrice(snapshotItem.Token)
		if err != nil {
			fmt.Printf("获取%s价格失败: %v\n", snapshotItem.Token, err)
			price = 0 // 获取价格失败时设为0
		}

		// 如果type是tge，在名字后面加上(tge)标识
		projectName := snapshotItem.Name
		if correspondingAirdrop.Type == "tge" {
			projectName += "(tge)"
		}

		// 格式化消息行，添加到消息内容中
		// 包含：代币符号、项目名称、日期、时间、积分、数量、阶段和价值(USD)
		msg += fmt.Sprintf("| %s(%s) | %s %s | %s | %s | %d | %.2f |\n",
			snapshotItem.Token, projectName, snapshotItem.Date, snapshotItem.Time,
			correspondingAirdrop.Points, snapshotItem.Amount, snapshotItem.Phase, price*float64(amount))
	}

	// 生成排序后的快照字符串，用于保存和比较
	snapshot := s.itemsToSnapshot(snapshotItems)

	return msg, snapshot // 返回消息内容和快照字符串
}

// parseSnapshot 解析快照字符串为结构体切片
// 该方法将保存的快照字符串转换回SnapshotItem结构体切片，用于比较和处理
// 参数:
//   - snapshot: 快照字符串，每行表示一个项目，字段用|分隔
// 返回:
//   - []SnapshotItem: 解析后的快照项目切片
func (s *AirdropService) parseSnapshot(snapshot string) []SnapshotItem {
	var items []SnapshotItem
	// 按行分割快照字符串，并去除首尾空白
	lines := strings.Split(strings.TrimSpace(snapshot), "\n")
	
	// 遍历每一行
	for _, line := range lines {
		if line == "" { // 跳过空行
			continue
		}
		// 按|分割字段
		parts := strings.Split(line, "|")
		// 确保有足够的字段
		if len(parts) >= 6 {
			// 将阶段字符串转换为整数
			phase, _ := strconv.Atoi(parts[5])
			// 创建SnapshotItem并添加到切片
			items = append(items, SnapshotItem{
				Token:  parts[0], // 代币符号
				Name:   parts[1], // 项目名称
				Date:   parts[2], // 空投日期
				Time:   parts[3], // 空投时间
				Amount: parts[4], // 空投数量
				Phase:  phase,    // 空投阶段
			})
		}
	}
	return items // 返回解析后的快照项目切片
}

// itemsToSnapshot 将结构体切片转换为快照字符串
// 该方法将SnapshotItem结构体切片转换为可保存的字符串格式
// 参数:
//   - items: 快照项目切片
// 返回:
//   - string: 格式化的快照字符串，每行表示一个项目，字段用|分隔
func (s *AirdropService) itemsToSnapshot(items []SnapshotItem) string {
	var lines []string
	// 遍历每个快照项，转换为字符串格式
	for _, item := range items {
		// 格式化为字符串，字段用|分隔
		lines = append(lines, fmt.Sprintf("%s|%s|%s|%s|%s|%d",
			item.Token, item.Name, item.Date, item.Time, item.Amount, item.Phase))
	}
	// 用换行符连接所有行
	return strings.Join(lines, "\n")
}

// sortSnapshotItems 对快照项进行排序
// 排序规则：先按时间排序（时间靠近的在上面），再按字母排序
// 参数:
//   - items: 需要排序的快照项目切片（原地排序）
func (s *AirdropService) sortSnapshotItems(items []SnapshotItem) {
	// 使用sort.Slice进行自定义排序
	sort.Slice(items, func(i, j int) bool {
		// 解析两个项目的日期时间
		timeI := s.parseDateTime(items[i].Date, items[i].Time)
		timeJ := s.parseDateTime(items[j].Date, items[j].Time)

		// 如果时间不同，按时间排序（时间靠近的在上面）
		// 即时间早的排在前面
		if !timeI.Equal(timeJ) {
			return timeI.Before(timeJ)
		}

		// 时间相同时，按Token字母排序（字母顺序）
		return items[i].Token < items[j].Token
	})
}

// parseDateTime 解析日期时间字符串为time.Time对象
// 该方法尝试解析日期和时间字符串，支持多种格式
// 参数:
//   - date: 日期字符串，格式为"2006-01-02"
//   - timeStr: 时间字符串，格式为"15:04"，可以为空
// 返回:
//   - time.Time: 解析后的时间对象，如果解析失败则返回零值
func (s *AirdropService) parseDateTime(date, timeStr string) time.Time {
	// 如果日期为空，直接返回零值时间
	if date == "" {
		return time.Time{}
	}
	
	// 构建日期时间字符串
	dateTimeStr := date
	// 如果有时间部分，添加到日期后面
	if timeStr != "" {
		dateTimeStr += " " + timeStr
		// 尝试解析完整的日期时间格式
		layout := "2006-01-02 15:04" // Go的时间格式化模板
		if parsed, err := time.Parse(layout, dateTimeStr); err == nil {
			return parsed // 解析成功，返回时间对象
		}
	}
	
	// 如果没有时间部分或解析失败，尝试只解析日期部分
	layout := "2006-01-02"
	if parsed, err := time.Parse(layout, date); err == nil {
		return parsed // 解析成功，返回时间对象
	}
	
	// 所有解析尝试都失败，返回零值时间
	return time.Time{}
}

// CompareSnapshots 比较两个快照是否相同（忽略顺序）
// 该方法通过统计每个项目的出现次数来比较两个快照是否包含相同的项目
// 参数:
//   - snapshot1: 第一个快照字符串
//   - snapshot2: 第二个快照字符串
// 返回:
//   - bool: 如果两个快照包含相同的项目（忽略顺序），则返回true，否则返回false
func (s *AirdropService) CompareSnapshots(snapshot1, snapshot2 string) bool {
	// 解析两个快照字符串为结构体切片
	items1 := s.parseSnapshot(snapshot1)
	items2 := s.parseSnapshot(snapshot2)

	// 如果项目数量不同，直接返回false
	if len(items1) != len(items2) {
		return false
	}

	// 创建map来统计每个项目的出现次数
	count1 := make(map[string]int)
	count2 := make(map[string]int)

	// 统计第一个快照中每个项目的出现次数
	for _, item := range items1 {
		// 将项目转换为字符串键
		key := fmt.Sprintf("%s|%s|%s|%s|%s|%d", item.Token, item.Name, item.Date, item.Time, item.Amount, item.Phase)
		count1[key]++ // 增加该项目的计数
	}

	// 统计第二个快照中每个项目的出现次数
	for _, item := range items2 {
		// 将项目转换为字符串键
		key := fmt.Sprintf("%s|%s|%s|%s|%s|%d", item.Token, item.Name, item.Date, item.Time, item.Amount, item.Phase)
		count2[key]++ // 增加该项目的计数
	}

	// 比较两个map是否相同
	// 检查第一个快照中的每个项目在第二个快照中是否有相同数量
	for key, count := range count1 {
		if count2[key] != count {
			return false // 如果数量不同，返回false
		}
	}

	// 所有项目数量都匹配，两个快照相同
	return true
}

// DetectSnapshotChange 检测快照变化类型
// 该方法分析两个快照之间的差异，判断是否有新增项目或只有删除项目
// 参数:
//   - oldSnapshot: 旧的快照字符串
//   - newSnapshot: 新的快照字符串
// 返回:
//   - bool: 是否有新增项目
//   - bool: 是否只有删除项目（没有新增）
func (s *AirdropService) DetectSnapshotChange(oldSnapshot, newSnapshot string) (bool, bool) {
	// 解析两个快照字符串为结构体切片
	oldItems := s.parseSnapshot(oldSnapshot)
	newItems := s.parseSnapshot(newSnapshot)

	// 创建map来快速查找项目是否存在
	oldMap := make(map[string]bool) // 旧快照中的项目集合
	newMap := make(map[string]bool) // 新快照中的项目集合

	// 将旧快照中的项目添加到map
	for _, item := range oldItems {
		// 将项目转换为字符串键
		key := fmt.Sprintf("%s|%s|%s|%s|%s|%d", item.Token, item.Name, item.Date, item.Time, item.Amount, item.Phase)
		oldMap[key] = true // 标记该项目存在于旧快照
	}

	// 将新快照中的项目添加到map
	for _, item := range newItems {
		// 将项目转换为字符串键
		key := fmt.Sprintf("%s|%s|%s|%s|%s|%d", item.Token, item.Name, item.Date, item.Time, item.Amount, item.Phase)
		newMap[key] = true // 标记该项目存在于新快照
	}

	// 检查是否有新增项目
	// 遍历新快照中的每个项目，检查是否在旧快照中不存在
	hasAddition := false
	for key := range newMap {
		if !oldMap[key] { // 如果旧快照中不存在该项目
			hasAddition = true // 标记有新增项目
			break
		}
	}

	// 检查是否只有删除（没有新增）
	// 条件：没有新增项目，且新快照的项目数量少于旧快照
	isOnlyDeletion := !hasAddition && len(newItems) < len(oldItems)

	return hasAddition, isOnlyDeletion
}
