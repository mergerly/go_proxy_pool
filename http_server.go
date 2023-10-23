package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
	"text/template"
)

type ProxyItemString struct {
	IP         string `json:"ip"`
	TypeString string `json:"typeString"`
	Address    string `json:"address"`
	CheckCount int    `json:"checkCount"`
	FailCount  int    `json:"failCount"`
	LastTime   string `json:"lastTime"`
	LastStatus bool   `json:"lastStatus"`
}

func httpStart() {
	router := mux.NewRouter()
	router.HandleFunc("/api", apiIndex).Methods("GET")
	router.HandleFunc("/all", getAllProxies).Methods("GET")
	router.HandleFunc("/get", getProxy).Methods("GET")
	router.HandleFunc("/api/all", getAllProxies).Methods("GET")
	router.HandleFunc("/api/get", getProxy).Methods("GET")
	router.HandleFunc("/api/pop", popProxy).Methods("GET")
	router.HandleFunc("/api/delete", deleteProxy).Methods("GET")
	router.HandleFunc("/api/count", couuntProxy).Methods("GET")

	addr := fmt.Sprintf("%s:%d", app.Config.Host, app.Config.Port) // 指定监听的地址和端口号
	fmt.Printf("Server running on %s\n", addr)
	err := http.ListenAndServe(addr, router)
	if err != nil {
		app.logger.Println(err)
	}
}

func apiIndex(w http.ResponseWriter, r *http.Request) {
	apiList := `[{"url": "/api/get", "params": "type: ''https'|''", "desc": "get a proxy"},
{"url": "/api/pop", "params": "", "desc": "get and delete a proxy"},
{"url": "/api/delete", "params": "proxy: 'e.g. 127.0.0.1:8080'", "desc": "delete an unable proxy"},
{"url": "/api/all", "params": "type: ''https'|''", "desc": "get all proxy from proxy pool"},
{"url": "/api/count", "params": "", "desc": "return proxy count"}]`
	jsonDataHandler(w, r, []byte(apiList))
}

func jsonHandler(w http.ResponseWriter, r *http.Request, proxies []*ProxyItem) {
	w.Header().Set("Content-Type", "application/json")
	jsonData, err := json.Marshal(proxies)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	_, err = w.Write(jsonData)
	if err != nil {
		log.Println("Error writing JSON response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func jsonDataHandler(w http.ResponseWriter, r *http.Request, jsonData []byte) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(jsonData)
	if err != nil {
		log.Println("Error writing JSON response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func htmlHandler(w http.ResponseWriter, r *http.Request, proxies []*ProxyItem) {
	w.Header().Set("Content-Type", "text/html")

	tmpl := `
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<title>代理</title>
			<style>
				table {
					border-collapse: collapse;
					width: 100%;
				}
				th, td {
					border: 1px solid #ddd;
					padding: 8px;
				}
				th {
					background-color: #f2f2f2;
				}
			</style>
		</head>
		<body>
			<h1>代理</h1>
			<table>
				<tr>
					<th>IP</th>
					<th>类型</th>
					<th>地址</th>
					<th>检查次数</th>
					<th>失败次数</th>
					<th>最近时间</th>
					<th>最近状态</th>
				</tr>
				{{range .}}
				<tr>
					<td>{{.IP}}</td>
					<td>{{.TypeString}}</td>
					<td>{{.Address}}</td>
					<td>{{.CheckCount}}</td>
					<td>{{.FailCount}}</td>
					<td>{{.LastTime}}</td>
					<td>{{.LastStatus}}</td>
				</tr>
				{{end}}
			</table>
		</body>
		</html>
	`

	t, err := template.New("代理").Parse(tmpl)
	if err != nil {
		log.Println("Error parsing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var proxyDataList []ProxyItemString
	for _, proxy := range proxies {
		// 根据 .Type 类型转换为相应的字符串
		var typeString string
		if proxy.Type&0x01 == 0x01 {
			typeString = "HTTP"
		}
		if proxy.Type&0x10 == 0x10 {
			typeString += "|HTTPS"
		}
		if proxy.Type&0x100 == 0x100 {
			typeString += "|SOCK5"
		}

		proxyData := ProxyItemString{
			IP:         proxy.IP,
			TypeString: typeString,
			Address:    proxy.Address,
			CheckCount: proxy.CheckCount,
			FailCount:  proxy.FailCount,
			LastTime:   proxy.LastTime,
			LastStatus: proxy.LastStatus,
		}

		proxyDataList = append(proxyDataList, proxyData)
	}

	err = t.Execute(w, proxyDataList)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func getAllProxies(w http.ResponseWriter, r *http.Request) {
	queryType := r.URL.Query().Get("type")
	proxyType := 0
	if queryType == "https" {
		proxyType = 17
	}
	var proxies []*ProxyItem
	if proxyType > 0 {
		proxies, _ = app.Database.GetAllType(proxyType)
	} else {
		proxies, _ = app.Database.GetAll()
	}
	if strings.HasPrefix(r.RequestURI, "/api/all") {
		jsonHandler(w, r, proxies)
	} else {
		htmlHandler(w, r, proxies)
	}
}

func getProxy(w http.ResponseWriter, r *http.Request) {
	queryType := r.URL.Query().Get("type")
	proxyType := 0
	if queryType == "https" {
		proxyType = 0x10
	}
	var proxy *ProxyItem
	if proxyType > 0 {
		proxy, _ = app.Database.GetType(proxyType)
	} else {
		proxy, _ = app.Database.Get()
	}
	if strings.HasPrefix(r.RequestURI, "/api/get") {
		jsonHandler(w, r, []*ProxyItem{proxy})
	} else {
		htmlHandler(w, r, []*ProxyItem{proxy})
	}
}

func popProxy(w http.ResponseWriter, r *http.Request) {
	proxy, _ := app.Database.Pop()
	jsonHandler(w, r, []*ProxyItem{proxy})
}

func deleteProxy(w http.ResponseWriter, r *http.Request) {
	proxy := r.URL.Query().Get("proxy")
	var jsonData string
	err := app.Database.Delete(proxy)
	if err != nil {
		log.Println(err)
		jsonData = fmt.Sprintf("{\"code\":0, \"status\":\"fail %s\"}", err.Error())
	} else {
		jsonData = fmt.Sprintf("{\"code\":0, \"status\":\"success\"}")
	}

	jsonDataHandler(w, r, []byte(jsonData))
}

func couuntProxy(w http.ResponseWriter, r *http.Request) {
	proxies, _ := app.Database.GetAll()
	jsonData := fmt.Sprintf("{\"count\":%d}", len(proxies))
	jsonDataHandler(w, r, []byte(jsonData))
}
