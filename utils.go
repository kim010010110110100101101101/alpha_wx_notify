package main

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	serverchan_sdk "github.com/easychen/serverchan-sdk-golang"
)

// Config 配置结构体
type Config struct {
	SendKeys []string `json:"sendkeys"`
	Interval int      `json:"interval"`
	FiterTge bool     `json:"fiterTge"`
}

// LoadConfig 读取配置文件
func LoadConfig() (*Config, error) {
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

// ReadResponseBody 处理可能被gzip压缩的响应体
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

// SendToServerChan 发送到Server酱
func SendToServerChan(msg string, title string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
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

// HashMsg 计算消息的MD5哈希
func HashMsg(msg string) string {
	h := md5.New()
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}

// SaveSnapshot 保存快照到文件
func SaveSnapshot(snapshot string, filename string) error {
	return ioutil.WriteFile(filename, []byte(snapshot), 0644)
}

// LoadLastSnapshot 读取上次的快照
func LoadLastSnapshot(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // 文件不存在，返回空字符串
		}
		return "", err
	}
	return string(data), nil
}