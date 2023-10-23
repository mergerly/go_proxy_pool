package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron"
)

func runProxyFetch() {
	proxyQueue := make(chan string)
	go func() {
		app.fetcher.run(app.Config.ProxyFetcher, proxyQueue)
		for proxy := range proxyQueue {
			proxyType := app.validator.VerifyProxy(proxy)
			fmt.Printf("%s proxy type:%0x\n", proxy, proxyType)
			if proxyType > 0 {
				if app.Database.Exists(proxy) {
					app.logger.Printf("RawProxyCheck - %s exist", proxy)
				} else {
					app.logger.Printf("RawProxyCheck - %s pass", proxy)
					region, _ := app.validator.regionGetter(proxy)
					item := NewProxyItem(proxy, region, proxyType)
					err := app.Database.Put(item)
					if err != nil {
						app.logger.Printf("RawProxyCheck - put %s fail", proxy)
						return
					}
				}
			} else {
				app.logger.Printf("RawProxyCheck - %s fail", proxy)
			}
		}
	}()
}

func runProxyCheck() {
	proxies, err := app.Database.GetAll()
	if err != nil {
		app.logger.Printf("UseProxyCheck - get all fail")
		return
	}
	if len(proxies) < app.Config.PoolSizeMin {
		runProxyFetch()
	}
	for _, proxy := range proxies {
		proxy.CheckCount += 1
		proxy.LastTime = time.Now().Format("2006-01-02 15:04:05")
		proxyType := app.validator.VerifyProxy(proxy.IP)
		if proxyType > 0 {
			proxy.LastStatus = true
			if proxy.FailCount > 0 {
				proxy.FailCount -= 1
			}
			app.logger.Printf("UseProxyCheck - %s pass", proxy.IP)
			err := app.Database.Put(proxy)
			if err != nil {
				app.logger.Printf("UseProxyCheck - put %s fail", proxy)
				return
			}
		} else {
			proxy.LastStatus = false
			proxy.FailCount += 1
			if proxy.FailCount > app.Config.MaxFailCount {
				app.logger.Printf("UseProxyCheck - %s fail, count %d delete", proxy.IP, proxy.FailCount)
				err := app.Database.Delete(proxy.IP)
				if err != nil {
					app.logger.Printf("UseProxyCheck - delete %s fail", proxy)
					return
				}
			} else {
				app.logger.Printf("UseProxyCheck - %s fail, count %d keep", proxy.IP, proxy.FailCount)
				err := app.Database.Put(proxy)
				if err != nil {
					app.logger.Printf("UseProxyCheck - put %s fail", proxy)
					return
				}
			}
		}
	}
}

func runScheduler() {
	runProxyFetch()

	// 创建调度器对象
	timezone, _ := time.LoadLocation(app.Config.Timezone)
	s := gocron.NewScheduler(timezone)

	// 定义获取代理的计划任务，每隔 4 分钟执行一次
	_, err := s.Every(4).Minutes().Do(runProxyFetch)
	if err != nil {
		log.Fatalf("Failed to define fetchProxies task: %s", err)
	}

	_, err = s.Every(2).Minutes().Do(runProxyCheck)
	if err != nil {
		log.Fatalf("Failed to schedule proxy check job: %s", err)
	}

	s.StartBlocking()
}

func testScheduler() {
	runScheduler()
}
