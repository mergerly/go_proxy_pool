package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ProxyFetcher struct {
	app *App
}

func NewProxyFetcher() (*ProxyFetcher, error) {
	return &ProxyFetcher{}, nil
}

func (pf *ProxyFetcher) Header() http.Header {
	headers := http.Header{}
	headers.Set("User-Agent", pf.getUserAgent())
	headers.Set("Accept", "*/*")
	headers.Set("Connection", "keep-alive")
	headers.Set("Accept-Language", "zh-CN,zh;q=0.8")
	return headers
}

func (pf *ProxyFetcher) getUserAgent() string {
	uaList := []string{
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/30.0.1599.101",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/38.0.2125.122",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.71",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.95",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.1 (KHTML, like Gecko) Chrome/21.0.1180.71",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1; QQDownload 732; .NET4.0C; .NET4.0E)",
		"Mozilla/5.0 (Windows NT 5.1; U; en; rv:1.8.1) Gecko/20061208 Firefox/2.0.0 Opera 9.50",
		"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:34.0) Gecko/20100101 Firefox/34.0",
	}
	rand.Seed(time.Now().UnixNano())
	return uaList[rand.Intn(len(uaList))]
}

func (pf *ProxyFetcher) Get(url string, verify bool) (*goquery.Document, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Get %s", err)
		return nil, err
	}

	// 设置请求头
	req.Header = pf.Header()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: !verify},
		},
	}
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Get %s", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Get status code error: %d %s", resp.StatusCode, resp.Status)
		return nil, fmt.Errorf("Get status code error: %d %s", resp.StatusCode, resp.Status)
	}

	// 使用 goquery 解析 HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("Get %s", err)
		return nil, err
	}

	return doc, nil
}

func (pf *ProxyFetcher) FreeProxy01(proxyChan chan<- string) {
	startURL := "https://www.zdaye.com/dayProxy.html"
	doc, err := pf.Get(startURL, false)
	if err != nil {
		log.Printf("Failed to parse %s: %s", startURL, err)
		return
	}

	latestPageTime := doc.Find("span.thread_time_info").First().Text()
	layout := "2006/01/02 15:04:05" // 定义日期时间的格式
	latestPageTimeParsed, err := time.Parse(layout, strings.TrimSpace(latestPageTime))
	if err != nil {
		log.Printf("Failed to parse latest_page_time: %s", err)
		return
	}

	interval := time.Since(latestPageTimeParsed)
	if interval.Seconds() < 300 {
		targetURL := "https://www.zdaye.com/" + doc.Find("h3.thread_title a").First().AttrOr("href", "")
		for targetURL != "" {
			doc, err := pf.Get(targetURL, false)
			if err != nil {
				log.Printf("Failed to parse %s: %s", targetURL, err)
				break
			}

			doc.Find("table tr").Each(func(i int, s *goquery.Selection) {
				if i > 0 {
					ip := strings.TrimSpace(s.Find("td:nth-child(1)").Text())
					port := strings.TrimSpace(s.Find("td:nth-child(2)").Text())
					proxyChan <- fmt.Sprintf("%s:%s", ip, port)
					log.Printf("FreeProxy01 get proxy %s:%s", ip, port)
				}
			})

			nextPage := doc.Find("div.page a[title='下一页']").AttrOr("href", "")
			if nextPage == "" {
				break
			}
			targetURL = "https://www.zdaye.com/" + nextPage
			// Sleep for 5 seconds before making the next request
			time.Sleep(5 * time.Second)
		}
	}
}

func (pf *ProxyFetcher) FreeProxy02(proxyChan chan<- string) {
	url := "http://www.66ip.cn/"
	doc, err := pf.Get(url, false)
	if err != nil {
		log.Printf("Failed to parse %s: %s", url, err)
		return
	}

	doc.Find("table:nth-child(3) tr").Each(func(i int, s *goquery.Selection) {
		if i > 0 {
			ip := strings.TrimSpace(s.Find("td:nth-child(1)").Text())
			port := strings.TrimSpace(s.Find("td:nth-child(2)").Text())
			proxyChan <- fmt.Sprintf("%s:%s", ip, port)
			log.Printf("FreeProxy02 get proxy %s:%s", ip, port)
		}
	})
}

func (pf *ProxyFetcher) FreeProxy03(proxyChan chan<- string) {
	targetURLs := []string{"http://www.kxdaili.com/dailiip.html", "http://www.kxdaili.com/dailiip/2/1.html"}
	for i := 2; i <= 10; i++ {
		targetURLs = append(targetURLs, fmt.Sprintf("http://www.kxdaili.com/dailiip/%d/1.html", i))
		targetURLs = append(targetURLs, fmt.Sprintf("http://www.kxdaili.com/dailiip/2/%d.html", i))
	}
	for _, url := range targetURLs {
		doc, err := pf.Get(url, false)
		if err != nil {
			log.Printf("Failed to parse %s: %s", url, err)
			continue
		}

		doc.Find("table.active tr").Each(func(i int, s *goquery.Selection) {
			if i > 0 {
				ip := strings.TrimSpace(s.Find("td:nth-child(1)").Text())
				port := strings.TrimSpace(s.Find("td:nth-child(2)").Text())
				proxyChan <- fmt.Sprintf("%s:%s", ip, port)
				log.Printf("FreeProxy03 get proxy %s:%s", ip, port)
			}
		})
		time.Sleep(5 * time.Second)
	}
}

func (pf *ProxyFetcher) FreeProxy04(proxyChan chan<- string) {
	url := "https://www.freeproxylists.net/zh/?c=CN&pt=&pr=&a%5B%5D=0&a%5B%5D=1&a%5B%5D=2&u=50"
	doc, err := pf.Get(url, false)
	if err != nil {
		log.Printf("Failed to parse %s: %s", url, err)
		return
	}

	re := regexp.MustCompile(`(?i)\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	doc.Find("tr.Odd, tr.Even").Each(func(i int, s *goquery.Selection) {
		ipScript := strings.TrimSpace(s.Find("td:nth-child(1) script").Text())
		ip := re.FindStringSubmatch(ipScript)
		port := strings.TrimSpace(s.Find("td:nth-child(2)").Text())
		if len(ip) > 0 {
			proxyChan <- fmt.Sprintf("%s:%s", ip[0], port)
			log.Printf("FreeProxy04 get proxy %s:%s", ip[0], port)
		}
	})
}

func (pf *ProxyFetcher) FreeProxy05(proxyChan chan<- string) {
	pageCount := 10
	urlPattern := []string{
		"https://www.kuaidaili.com/free/inha/%d/",
		"https://www.kuaidaili.com/free/intr/%d/",
	}
	urlList := []string{}
	for pageIndex := 1; pageIndex <= pageCount; pageIndex++ {
		for _, pattern := range urlPattern {
			urlList = append(urlList, fmt.Sprintf(pattern, pageIndex))
		}
	}

	for _, url := range urlList {
		time.Sleep(1 * time.Second) // Sleep for 1 second
		doc, err := pf.Get(url, false)
		if err != nil {
			log.Printf("Failed to create document from response: %v", err)
			continue
		}

		doc.Find("table tr").Each(func(i int, s *goquery.Selection) {
			if i > 0 {
				ip := s.Find("td").Eq(0).Text()
				port := s.Find("td").Eq(1).Text()
				proxy := ip + ":" + port
				proxyChan <- proxy
			}
		})

		time.Sleep(5 * time.Second)
	}
}

func (pf *ProxyFetcher) FreeProxy06(proxyChan chan<- string) error {
	url := "http://proxylist.fatezero.org/proxy.list"
	doc, err := pf.Get(url, false)
	if err != nil {
		log.Printf("Failed to parse %s: %s", url, err)
		return err
	}

	body := doc.Find("body").Text()
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		var jsonInfo map[string]interface{}
		err := json.Unmarshal([]byte(line), &jsonInfo)
		if err != nil {
			return err
		}

		host := jsonInfo["host"].(string)
		port := int(jsonInfo["port"].(float64))
		proxy := fmt.Sprintf("%s:%d", host, port)
		proxyChan <- proxy
	}

	return nil
}

func (pf *ProxyFetcher) FreeProxy07(proxyChan chan<- string) {
	urls := []string{"http://www.ip3366.net/free/?stype=1", "http://www.ip3366.net/free/?stype=2"}
	for _, url := range urls {
		doc, err := pf.Get(url, false)
		if err != nil {
			fmt.Println("Error requesting URL:", err)
			continue
		}

		body, _ := doc.Html()
		proxyRegex := regexp.MustCompile(`<td>(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})</td>[\s\S]*?<td>(\d+)</td>`)
		matches := proxyRegex.FindAllStringSubmatch(body, -1)
		for _, match := range matches {
			ip := match[1]
			port := match[2]
			proxy := ip + ":" + port
			proxyChan <- proxy
		}
	}
}

func (pf *ProxyFetcher) FreeProxy08(proxyChan chan<- string) {
	urls := []string{"https://ip.ihuan.me/address/5Lit5Zu9.html"}
	for _, url := range urls {
		doc, err := pf.Get(url, false)
		if err != nil {
			fmt.Println("Error requesting URL:", err)
			continue
		}

		html, err := doc.Html()
		if err != nil {
			fmt.Println("Error getting HTML content:", err)
			continue
		}

		proxyRegex := regexp.MustCompile(`>\s*?(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s*?</a></td><td>(\d+)</td>`)
		matches := proxyRegex.FindAllStringSubmatch(html, -1)
		for _, match := range matches {
			ip := match[1]
			port := match[2]
			proxy := ip + ":" + port
			proxyChan <- proxy
		}
	}
}

func (pf *ProxyFetcher) FreeProxy09(proxyChan chan<- string) {
	pageCount := 1
	urlList := []string{}
	for pageIndex := 1; pageIndex <= pageCount; pageIndex++ {
		urlList = append(urlList, fmt.Sprintf("http://ip.jiangxianli.com/?country=中国&page=%d", pageIndex))
	}

	for _, url := range urlList {
		doc, err := pf.Get(url, false)
		if err != nil {
			fmt.Println("Error requesting URL:", err)
			continue
		}

		body, _ := doc.Html()
		proxyRegex := regexp.MustCompile(`<td>(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})</td>[\s\S]*?<td>(\d+)</td>`)
		matches := proxyRegex.FindAllStringSubmatch(body, -1)
		for _, match := range matches {
			ip := match[1]
			port := match[2]
			proxy := ip + ":" + port
			proxyChan <- proxy
		}
	}
}

func (pf *ProxyFetcher) FreeProxy10(proxyChan chan<- string) {
	pageCount := 100
	var urls []string
	for pageIndex := 1; pageIndex <= pageCount; pageIndex++ {
		urls = append(urls, fmt.Sprintf("https://www.89ip.cn/index_%d.html", pageIndex))
	}
	for _, url := range urls {
		doc, err := pf.Get(url, false)
		if err != nil {
			fmt.Println("Error requesting URL:", err)
			continue
		}

		html, err := doc.Html()
		if err != nil {
			fmt.Println("Error getting HTML content:", err)
			continue
		}

		proxyRegex := regexp.MustCompile(`<td.*?>[\s\S]*?(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})[\s\S]*?</td>[\s\S]*?<td.*?>[\s\S]*?(\d+)[\s\S]*?</td>`)
		matches := proxyRegex.FindAllStringSubmatch(html, -1)
		for _, match := range matches {
			ip := match[1]
			port := match[2]
			proxy := ip + ":" + port
			proxyChan <- proxy
		}
		time.Sleep(5 * time.Second)
	}
}

func (pf *ProxyFetcher) FreeProxy11(proxyChan chan<- string) {
	pageCount := 10
	var urlList []string
	for pageIndex := 1; pageIndex <= pageCount; pageIndex++ {
		urlList = append(urlList, fmt.Sprintf("https://list.proxylistplus.com/Fresh-HTTP-Proxy-List-%d", pageIndex))
	}

	for _, url := range urlList {
		doc, err := pf.Get(url, false)
		if err != nil {
			fmt.Println("Error requesting URL:", err)
			continue
		}

		body, _ := doc.Html()
		proxyRegex := regexp.MustCompile(`<td>(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})</td>[\s\S]*?<td>(\d+)</td>`)
		matches := proxyRegex.FindAllStringSubmatch(body, -1)
		for _, match := range matches {
			ip := match[1]
			port := match[2]
			proxy := ip + ":" + port
			proxyChan <- proxy
		}

		time.Sleep(5 * time.Second)
	}
}

func (pf *ProxyFetcher) run(fetchers []string, output chan string) {
	wg := sync.WaitGroup{}
	for _, fetcherName := range fetchers {
		// 检查ProxyFetcher结构体是否存在与fetcherName相同的方法
		methodValue := reflect.ValueOf(pf).MethodByName(fetcherName)
		if methodValue.IsValid() {
			wg.Add(1)
			// 调用存在的方法
			go func() {
				//methodValue.Call(nil)
				methodValue.Call([]reflect.Value{reflect.ValueOf(output)})
				wg.Done()
			}()
		}
	}

	// 启动一个协程等待所有协程完成
	go func() {
		wg.Wait()
		close(output) // 所有协程完成后关闭输出管道
		log.Printf("All fetchers completed")
	}()
}

func testProxyFetcher() {
	fetchers := []string{"FreeProxy10"} //"FreeProxy01","FreeProxy02",
	proxyQueue := make(chan string)
	pf := &ProxyFetcher{}
	pf.run(fetchers, proxyQueue)

	for proxy := range proxyQueue {
		fmt.Println(proxy)
	}
	log.Printf("testProxyFetcher completed")
}
