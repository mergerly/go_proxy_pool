package main

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"os"
)

type Config struct {
	Host          string
	Port          int
	DBName        string
	TableName     string
	ProxyFetcher  []string
	HttpURL       string
	HttpsURL      string
	VerifyTimeout int
	MaxFailCount  int
	PoolSizeMin   int
	ProxyRegion   bool
	Timezone      string
}

func NewConfig(filePath string) (*Config, error) {
	config := &Config{}
	return config, config.LoadFromFile(filePath)
}

func (c *Config) LoadFromFile(filePath string) error {
	defaultConfig := &Config{
		Host:          "0.0.0.0",
		Port:          5010,
		DBName:        "proxies.db",
		TableName:     "use_proxy",
		ProxyFetcher:  []string{"FreeProxy01", "FreeProxy02", "FreeProxy03", "FreeProxy04", "FreeProxy05", "FreeProxy06", "FreeProxy07", "FreeProxy08", "FreeProxy09", "FreeProxy10", "FreeProxy11"},
		HttpURL:       "http://httpbin.org",
		HttpsURL:      "https://www.qq.com",
		VerifyTimeout: 10,
		MaxFailCount:  0,
		PoolSizeMin:   20,
		ProxyRegion:   true,
		Timezone:      "Asia/Shanghai",
	}

	config, err := toml.LoadFile(filePath)
	if err != nil {
		// 加载 TOML 文件失败，使用默认配置
		*c = *defaultConfig

		// 如果某些值在配置文件中缺失，则保存默认配置到文件中
		if err := c.SaveToFile(filePath); err != nil {
			return err
		}
		return nil
	}

	// 从 TOML 文件中加载值并覆盖默认配置
	if err := config.Unmarshal(c); err != nil {
		return err
	}

	return nil
}

func (c *Config) SaveToFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(c); err != nil {
		return err
	}

	return nil
}

func testConfig() {
	config := &Config{}
	err := config.LoadFromFile("config.toml")
	if err != nil {
		fmt.Println("加载配置失败:", err)
		return
	}

	// 在这里使用配置值...
	fmt.Println("主机:", config.Host)
	fmt.Println("端口:", config.Port)
	// ...
}
