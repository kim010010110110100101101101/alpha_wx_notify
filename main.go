package main

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	serverchan_sdk "github.com/easychen/serverchan-sdk-golang"
)

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

type ApiResponse struct {
	Airdrops []Airdrop `json:"airdrops"`
}

// 配置结构体
type Config struct {
	SendKeys []string `json:"sendkeys"`
	Interval int      `json:"interval"`
	FiterTge bool     `json:"fiterTge"`
}

// 读取配置文件
func loadConfig() (*Config, error) {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// 处理可能被gzip压缩的响应体
func readResponseBody(resp *http.Response) ([]byte, error) {
	var reader io.Reader = resp.Body

	// 检查是否是gzip压缩
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	return ioutil.ReadAll(reader)
}

// 发送到Server酱
func sendToServerChan(msg string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	title := "今日空投播报"
	for _, sendkey := range cfg.SendKeys {
		resp, err := serverchan_sdk.ScSend(sendkey, title, msg, nil)
		if err != nil {
			fmt.Printf("推送Server酱失败: %v\n", err)
		} else {
			fmt.Println("Server酱响应:", resp)
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func getAirdrop() *ApiResponse {
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

		body, err := ioutil.ReadAll(resp.Body)
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
		// if len(body) > 100 {
		// fmt.Printf("响应内容前100字符: %s\n", string(body[:100]))
		// } else {
		fmt.Printf("响应内容: %s\n", string(body))
		// }

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

// 获取token单价
func fetchTokenPrice(token string) (float64, error) {
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

func getSendMsgAndSnapshot() (string, string) {
	fmt.Printf("今日日期: %s\n", time.Now().Format("2006-01-02"))

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	apiResp := getAirdrop()
	if apiResp == nil {
		fmt.Println("获取空投数据失败")
		return "", ""
	}

	msg := "| 项目 | 时间 | 积分 | 数量 | 阶段 | 价格(USD) |\n|---|---|---|---|---|---|\n"
	snapshot := ""
	isEmpty := true
	for i, item := range apiResp.Airdrops {
		var amount int
		if item.Amount == "" {
			amount = 0
		} else {
			var err error
			amount, err = strconv.Atoi(item.Amount)
			if err != nil {
				fmt.Printf("转换数量失败: %v\n", err)
				amount = 0
			}
		}
		// 比较日期是否是今天
		if item.Date != time.Now().Format("2006-01-02") {
			continue
		}

		if cfg.FiterTge && item.Type == "tge" {
			fmt.Printf("过滤TGE: %+v\n", item)
			continue
		}

		price, err := fetchTokenPrice(item.Token)
		if err != nil {
			fmt.Printf("获取%s价格失败: %v\n", item.Token, err)
			price = 0
		}
		msg += fmt.Sprintf("| %s(%s) | %s %s | %s | %s | %d | %.2f |\n",
			item.Token, item.Name, item.Date, item.Time, item.Points, item.Amount, item.Phase, price*float64(amount))
		snapshot += fmt.Sprintf("%s|%s|%s|%s|%s|%d\n",
			item.Token, item.Name, item.Date, item.Time, item.Amount, item.Phase)
		apiResp.Airdrops[i] = item
		isEmpty = false
	}
	if isEmpty {
		return "", ""
	}
	return msg, snapshot
}
func hashMsg(msg string) string {
	h := md5.New()
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}

// 保存快照到文件
func saveSnapshot(snapshot string) error {
	return ioutil.WriteFile("last_snapshot.txt", []byte(snapshot), 0644)
}

// 读取上次的快照
func loadLastSnapshot() (string, error) {
	data, err := ioutil.ReadFile("last_snapshot.txt")
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // 文件不存在，返回空字符串
		}
		return "", err
	}
	return string(data), nil
}

func main() {
	fmt.Printf("[%s] 开始检查空投信息...\n", time.Now().Format("2006-01-02 15:04:05"))

	msg, snapshot := getSendMsgAndSnapshot()

	if msg != "" {
		// 读取上次的快照
		lastSnapshot, err := loadLastSnapshot()
		if err != nil {
			fmt.Printf("读取上次快照失败: %v\n", err)
		}

		// 比较当前快照和上次快照
		currentHash := hashMsg(snapshot)
		lastHash := hashMsg(lastSnapshot)

		if currentHash != lastHash {
			fmt.Println("检测到空投信息变化，推送通知...")
			fmt.Println(msg)

			// 推送通知
			if err := sendToServerChan(msg); err != nil {
				fmt.Println("推送Server酱失败:", err)
			} else {
				fmt.Println("推送成功！")
			}

			// 保存当前快照
			if err := saveSnapshot(snapshot); err != nil {
				fmt.Printf("保存快照失败: %v\n", err)
			}
		} else {
			fmt.Println("空投信息无变化，跳过推送。")
		}
	} else {
		fmt.Println("今日无空投信息。")
	}
}
