package main

import (
	"io"
	"log"
	"os"
)

type App struct {
	Config    *Config
	validator *ProxyValidator
	fetcher   *ProxyFetcher
	Database  *ProxyDB
	logger    *log.Logger
	Version   string
}

var app *App

func main() {

	app = &App{}

	app.Version = "2.4.0"

	app.Config, _ = NewConfig("config.toml")
	app.Database, _ = NewProxyDB(app.Config.DBName, app.Config.TableName)

	app.fetcher, _ = NewProxyFetcher()
	app.validator = NewProxyValidator(app.Config.HttpURL, app.Config.HttpsURL, app.Config.VerifyTimeout)

	// 创建日志文件
	fileName := "go_proxy_pool.log"
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to create log file:%s", err)
	}
	defer file.Close()
	// 创建日志记录器
	app.logger = log.New(io.MultiWriter(file, os.Stdout), "", log.Ldate|log.Ltime)

	app.logger.Printf("Go Proxy Pool v%s Start\n", app.Version)

	go runScheduler()

	httpStart()

}

func test() {

	testProxyDB()

	testProxyValidator()

	testProxyFetcher()

	testConfig()
}
