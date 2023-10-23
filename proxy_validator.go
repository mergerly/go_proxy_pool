package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	// IP_REGEX 用于匹配代理格式
	IP_REGEX = regexp.MustCompile(`(.*:.*@)?\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d{1,5}`)
)

// ProxyValidator 类用于验证代理
type ProxyValidator struct {
	httpUrl       string
	httpsUrl      string
	verifyTimeout int
}

// NewProxyValidator 返回 ProxyValidator 实例
func NewProxyValidator(httpURL, httpsURL string, timeout int) *ProxyValidator {
	return &ProxyValidator{
		httpUrl:       httpURL,
		httpsUrl:      httpsURL,
		verifyTimeout: timeout,
	}
}

// Initialize 初始化 httpUrl 和 httpsUrl
func (pv *ProxyValidator) Initialize(httpURL, httpsURL string) {
	pv.httpUrl = httpURL
	pv.httpsUrl = httpsURL
}

// FormatValidator 检查代理格式是否合法
func (pv *ProxyValidator) FormatValidator(proxy string) bool {
	return IP_REGEX.MatchString(proxy)
}

// TimeoutValidator 检测代理超时
func (pv *ProxyValidator) TimeoutValidator(proxy, protocol string) bool {
	proxies := map[string]string{
		"http":   fmt.Sprintf("http://%s", proxy),
		"https":  fmt.Sprintf("https://%s", proxy),
		"socks5": fmt.Sprintf("socks5://%s", proxy),
	}

	proxys := proxies[protocol]
	if protocol == "https" {
		proxys = proxies["http"]
	}
	// 创建代理 URL
	proxyURL, err := url.Parse(proxys)
	if err != nil {
		fmt.Println("无法解析代理 URL:", err)
		return false
	}

	// 创建自定义的 Transport，并设置代理
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyURL(proxyURL),
	}

	// 创建自定义的 Client
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(pv.verifyTimeout) * time.Second,
	}

	httpsUrl := pv.httpsUrl
	if protocol == "http" {
		httpsUrl = pv.httpUrl
	}

	req, err := http.NewRequest("GET", httpsUrl, nil)
	if err != nil {
		return false
	}

	// 设置请求头
	pv.setRequestHeaders(req)

	// 设置代理
	//req.Header.Set("ProxyItem", proxies[protocol])

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to get url:", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// CustomValidatorExample 自定义验证函数示例
func (pv *ProxyValidator) CustomValidatorExample(proxy string) bool {
	// 自定义验证逻辑
	return false
}

type IPData struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data Data   `json:"data"`
}

type Data struct {
	Address string `json:"address"`
	IP      string `json:"ip"`
}

// 返回的JSON结构 {"code":200,"msg":"success","data":{"address":"中国 上海 上海 电信","ip":"101.230.187.69"}}
func (pv *ProxyValidator) regionGetter(proxy string) (string, error) {
	// 带有用户名密码的格式
	parts := strings.Split(proxy, "@")
	if len(parts) == 2 {
		proxy = parts[1]
	}
	ip := strings.Split(proxy, ":")[0]
	httpsUrl := fmt.Sprintf("https://searchplugin.csdn.net/api/v1/ip/get?ip=%s", ip)

	resp, err := http.Get(httpsUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ipData IPData
	err = json.Unmarshal(body, &ipData)
	if err != nil {
		return "", err
	}

	if ipData.Code == 200 {
		return ipData.Data.Address, nil
	}

	return "", fmt.Errorf("API response code is not 200")
}

// VerifyProxy 验证代理
func (pv *ProxyValidator) VerifyProxy(proxy string) int {
	isValid := pv.FormatValidator(proxy)
	proxyType := 0

	if isValid {
		isValid = pv.TimeoutValidator(proxy, "http")
		if isValid {
			proxyType = 0x1
		}
	}

	if isValid {
		isValid = pv.TimeoutValidator(proxy, "https")
		if isValid {
			proxyType |= 0x10
		}
	}

	if isValid {
		isValid = pv.TimeoutValidator(proxy, "socks5")
		if isValid {
			proxyType |= 0x100
		}
	}

	if isValid {
		isValid = pv.CustomValidatorExample(proxy)
		if isValid {
			proxyType |= 0x1000
		}
	}

	return proxyType
}

// setRequestHeaders 设置请求头
func (pv *ProxyValidator) setRequestHeaders(req *http.Request) {
	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:34.0) Gecko/20100101 Firefox/34.0",
		"Accept":          "*/*",
		"Connection":      "keep-alive",
		"Accept-Language": "zh-CN,zh;q=0.8",
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

func testProxyValidator() {
	HTTP_URL := "http://httpbin.org"
	HTTPS_URL := "https://www.qq.com"

	proxyValidator := NewProxyValidator(HTTP_URL, HTTPS_URL, 5)
	//proxyValidator.Initialize("http://example.com", "https://example.com")

	//proxy := "192.168.92.152:1080"
	proxy := "proxy:ztgame123456@211.159.201.232:18187"
	isValid := proxyValidator.VerifyProxy(proxy)
	fmt.Printf("代理验证结果:%0x\n", isValid)
	region, _ := proxyValidator.regionGetter(proxy)
	fmt.Println("代理%s的地址:", proxy, region)
}
