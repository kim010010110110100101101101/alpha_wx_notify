// Package internal 包含项目的核心功能实现
// 该包提供空投数据获取、处理、快照比较和消息推送等功能
package internal

import (
	"compress/gzip"    // 用于处理gzip压缩的响应
	"crypto/md5"      // 用于计算消息的MD5哈希
	"encoding/hex"    // 用于将MD5哈希转换为十六进制字符串
	"encoding/json"   // 用于JSON编码和解码
	"fmt"            // 用于格式化输出
	"io"             // 用于I/O操作
	"net/http"       // 用于HTTP请求
	"os"             // 用于文件操作
	"strings"        // 用于字符串处理
	"time"           // 用于时间相关操作

	serverchan_sdk "github.com/easychen/serverchan-sdk-golang" // Server酱SDK，用于消息推送
)

// Config 配置结构体
// 用于存储应用程序的配置信息，从config.json文件中加载
type Config struct {
	SendKeys []string `json:"sendkeys"` // Server酱的推送密钥列表
	Interval int      `json:"interval"` // 检查间隔时间（分钟）
	FiterTge bool     `json:"fiterTge"` // 是否过滤TGE类型的空投项目
}

// LoadConfig 读取配置文件
// 该函数从指定路径加载JSON格式的配置文件，并解析为Config结构体
// 参数:
//   - configPath: 配置文件的路径
// 返回:
//   - *Config: 解析后的配置对象指针
//   - error: 如果读取或解析失败，返回相应的错误
func LoadConfig(configPath string) (*Config, error) {
	// 读取配置文件内容
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err // 文件读取失败
	}
	
	// 解析JSON到Config结构体
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err // JSON解析失败
	}
	
	return &cfg, nil // 返回配置对象指针
}

// readResponseBody 处理可能被gzip压缩的响应体
// 该函数检测HTTP响应是否使用gzip压缩，并相应地解压缩和读取响应体
// 参数:
//   - resp: HTTP响应对象
// 返回:
//   - []byte: 响应体的字节数组
//   - error: 如果读取或解压缩失败，返回相应的错误
func readResponseBody(resp *http.Response) ([]byte, error) {
	// 初始化reader为响应体
	var reader io.Reader = resp.Body

	// 检查响应头中是否包含gzip压缩标识
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		// 创建gzip解压缩reader
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err // gzip解压缩初始化失败
		}
		defer gzipReader.Close() // 确保关闭gzip reader
		reader = gzipReader // 使用gzip reader替代原始reader
	}

	// 读取全部内容并返回
	return io.ReadAll(reader)
}

// SendToServerChan 发送消息到Server酱
// 该函数使用Server酱SDK将消息推送到指定的接收端（如微信）
// 参数:
//   - msg: 要发送的消息内容
//   - title: 消息标题
//   - cfg: 包含SendKeys的配置对象
// 返回:
//   - error: 目前总是返回nil，实际错误会打印到控制台
func SendToServerChan(msg string, title string, cfg *Config) error {
	// 遍历所有SendKey，分别发送消息
	for _, sendkey := range cfg.SendKeys {
		// 调用Server酱SDK发送消息
		resp, err := serverchan_sdk.ScSend(sendkey, title, msg, nil)
		if err != nil {
			// 发送失败，打印错误信息
			fmt.Printf("推送Server酱失败: %v\n", err)
		} else {
			// 发送成功，打印响应
			fmt.Println("Server酱响应:", resp)
		}
		// 每次发送后等待1秒，避免频率限制
		time.Sleep(1 * time.Second)
	}
	return nil // 始终返回nil，实际错误已在上面处理
}

// HashMsg 计算消息的MD5哈希值
// 该函数用于生成消息内容的唯一标识，用于比较消息是否发生变化
// 参数:
//   - msg: 要计算哈希的消息字符串
// 返回:
//   - string: 消息的MD5哈希值的十六进制字符串表示
func HashMsg(msg string) string {
	// 创建MD5哈希对象
	h := md5.New()
	// 写入消息内容
	h.Write([]byte(msg))
	// 计算哈希值并转换为十六进制字符串
	return hex.EncodeToString(h.Sum(nil))
}

// SaveSnapshot 保存快照到文件
// 该函数将当前的空投数据快照保存到指定文件，用于后续比较
// 参数:
//   - snapshot: 要保存的快照字符串
//   - filename: 保存的文件路径
// 返回:
//   - error: 如果写入文件失败，返回相应的错误
func SaveSnapshot(snapshot string, filename string) error {
	// 将快照字符串写入文件，权限设置为0644（用户可读写，组和其他用户可读）
	return os.WriteFile(filename, []byte(snapshot), 0644)
}

// LoadLastSnapshot 读取上次保存的快照
// 该函数从指定文件加载上次保存的空投数据快照，用于与当前快照比较
// 参数:
//   - filename: 快照文件的路径
// 返回:
//   - string: 读取的快照内容，如果文件不存在则返回空字符串
//   - error: 如果读取文件失败（且不是因为文件不存在），返回相应的错误
func LoadLastSnapshot(filename string) (string, error) {
	// 读取文件内容
	data, err := os.ReadFile(filename)
	if err != nil {
		// 检查错误类型
		if os.IsNotExist(err) {
			return "", nil // 文件不存在，返回空字符串，不视为错误
		}
		return "", err // 其他错误，返回错误信息
	}
	// 将文件内容转换为字符串并返回
	return string(data), nil
}
