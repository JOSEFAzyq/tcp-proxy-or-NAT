package internal

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Server ServerConfig `json:"server"`
}

var MyConfig Config

type ServerConfig struct {
	Host  string
	Port  string
	Proxy string
}

func init() {
	LoadConfig()
}

func LoadConfig() {
	dataBytes, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Println("读取文件失败：", err)
		return
	}
	var config Config
	err = yaml.Unmarshal(dataBytes, &config)
	if err != nil {
		fmt.Println("解析 yaml 文件失败：", err)
		return
	}
	MyConfig = config
	fmt.Println("配置文件读取完成", MyConfig)
}
