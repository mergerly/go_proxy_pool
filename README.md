### go_proxy_pool爬虫代理IP池

### 项目简介

参考[proxy_pool](https://github.com/jhao104/proxy_pool)用go语言实现的代理池，功能和proxy_pool一样。实现了单文件运行，不需要任何配置就可以运行项目。

### 运行项目

##### 编译代码:

* 使用go1.21以上版本编译

```bash
git clone https://github.com/mergerly/go_proxy_pool.git
go build
```

#### 启动项目

```
./go_proxy_pool
```

##### 更新配置:

首次运行项目会自动生成配置config.toml，自动创建SQLite数据库proxies.db，可以根据需要修改配置。

```toml
Host = "0.0.0.0"
Port = 5010
HttpURL = "http://httpbin.org"
HttpsURL = "https://www.qq.com"
MaxFailCount = 0
PoolSizeMin = 20
ProxyFetcher = ["FreeProxy01", "FreeProxy02", "FreeProxy03", "FreeProxy04", "FreeProxy05", "FreeProxy06", "FreeProxy07", "FreeProxy08", "FreeProxy09", "FreeProxy10", "FreeProxy11"]
ProxyRegion = true
DBName = "proxies.db"
TableName = "use_proxy"
Timezone = "Asia/Shanghai"
VerifyTimeout = 10
```

### 使用

* html

启动web服务后, 默认配置下会开启 http://127.0.0.1:5010 的html网页:

| url  | method | Description      | params                                                       |
| ---- | ------ | ---------------- | ------------------------------------------------------------ |
| /get | GET    | 随机获取一个代理 | 可选参数: `?type=https` 过滤支持https的代理, `?type=sock5` 过滤支持sock5的代理 |
| /all | GET    | 获取所有代理     | 可选参数: `?type=https` 过滤支持https的代理, `?type=sock5` 过滤支持sock5的代理 |


* Api

默认配置下会开启 http://127.0.0.1:5010 的api接口服务:

| api         | method | Description        | params                                                       |
| ----------- | ------ | ------------------ | ------------------------------------------------------------ |
| /api        | GET    | api介绍            | None                                                         |
| /api/get    | GET    | 随机获取一个代理   | 可选参数: `?type=https` 过滤支持https的代理, `?type=sock5` 过滤支持sock5的代理 |
| /api/pop    | GET    | 获取并删除一个代理 | 可选参数: `?type=https` 过滤支持https的代理, `?type=sock5` 过滤支持sock5的代理 |
| /api/all    | GET    | 获取所有代理       | 可选参数: `?type=https` 过滤支持https的代理, `?type=sock5` 过滤支持sock5的代理 |
| /api/count  | GET    | 查看代理数量       | None                                                         |
| /api/delete | GET    | 删除代理           | `?proxy=host:ip`                                             |


### 免费代理源

   目前实现的采集免费代理网站有(排名不分先后, 下面仅是对其发布的免费代理情况, 付费代理测评可以参考[这里](https://zhuanlan.zhihu.com/p/33576641)): 

| 代理名称      | 状态 | 更新速度 | 可用率 | 地址                                       |
| ------------- | ---- | -------- | ------ | ------------------------------------------ |
| 站大爷        | ✔    | ★        | **     | [地址](https://www.zdaye.com/)             |
| 66代理        | ✔    | ★        | *      | [地址](http://www.66ip.cn/)                |
| 开心代理      | ✔    | ★        | *      | [地址](http://www.kxdaili.com/)            |
| FreeProxyList | ✔    | ★        | *      | [地址](https://www.freeproxylists.net/zh/) |
| 快代理        | ✔    | ★        | *      | [地址](https://www.kuaidaili.com/)         |
| FateZero      | ✔    | ★★       | *      | [地址](http://proxylist.fatezero.org)      |
| 云代理        | ✔    | ★        | *      | [地址](http://www.ip3366.net/)             |
| 小幻代理      | ✔    | ★★       | *      | [地址](https://ip.ihuan.me/)               |
| 免费代理库    | ✔    | ☆        | *      | [地址](http://ip.jiangxianli.com/)         |
| 89代理        | ✔    | ☆        | *      | [地址](https://www.89ip.cn/)               |
