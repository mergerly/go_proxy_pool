package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ProxyItem struct {
	IP         string `json:"ip"`
	Type       int    `json:"type"`
	Address    string `json:"address"`
	CheckCount int    `json:"checkCount"`
	FailCount  int    `json:"failCount"`
	LastTime   string `json:"lastTime"`
	LastStatus bool   `json:"lastStatus"`
}

func NewProxyItem(ip, address string, proxyType int) *ProxyItem {
	return &ProxyItem{
		IP:         ip,
		Type:       proxyType,
		Address:    address,
		CheckCount: 1,
		FailCount:  0,
		LastTime:   time.Now().Format("2006-01-02 15:04:05"),
		LastStatus: true,
	}
}

type ProxyDB struct {
	db    *sql.DB
	table string
}

func NewProxyDB(dbPath, tableName string) (*ProxyDB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		ip TEXT PRIMARY KEY,
		address TEXT,
		type INTEGER,
		check_count INTEGER,
		fail_count INTEGER,
		last_time TEXT,
		last_status INTEGER
	)`, tableName))
	if err != nil {
		return nil, err
	}

	return &ProxyDB{
		db:    db,
		table: tableName,
	}, nil
}

func (pdb *ProxyDB) Close() {
	pdb.db.Close()
}

func (pdb *ProxyDB) Get() (*ProxyItem, error) {
	row := pdb.db.QueryRow(fmt.Sprintf("SELECT ip, address, type, check_count, fail_count, last_time, last_status FROM %s LIMIT 1", pdb.table))

	proxy := &ProxyItem{}
	err := row.Scan(&proxy.IP, &proxy.Address, &proxy.Type, &proxy.CheckCount, &proxy.FailCount, &proxy.LastTime, &proxy.LastStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return proxy, nil
}

func (pdb *ProxyDB) GetType(proxyType int) (*ProxyItem, error) {
	row := pdb.db.QueryRow(fmt.Sprintf("SELECT ip, address, type, check_count, fail_count, last_time, last_status FROM %s WHERE (type & ?) = ?", pdb.table), proxyType, proxyType)

	proxy := &ProxyItem{}
	err := row.Scan(&proxy.IP, &proxy.Address, &proxy.Type, &proxy.CheckCount, &proxy.FailCount, &proxy.LastTime, &proxy.LastStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return proxy, nil
}

func (pdb *ProxyDB) Put(proxy *ProxyItem) error {
	_, err := pdb.db.Exec(fmt.Sprintf("INSERT OR REPLACE INTO %s (ip, address, type, check_count, fail_count, last_time, last_status) VALUES (?, ?, ?, ?, ?, ?, ?)", pdb.table), proxy.IP, proxy.Address, proxy.Type, proxy.CheckCount, proxy.FailCount, proxy.LastTime, proxy.LastStatus)
	if err != nil {
		return err
	}

	return nil
}

func (pdb *ProxyDB) Delete(ip string) error {
	_, err := pdb.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE ip = ?", pdb.table), ip)
	if err != nil {
		return err
	}

	return nil
}

func (pdb *ProxyDB) Exists(ip string) bool {
	row := pdb.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE ip = ?", pdb.table), ip)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return false
	}

	return count > 0
}

func (pdb *ProxyDB) GetAll() ([]*ProxyItem, error) {
	rows, err := pdb.db.Query(fmt.Sprintf("SELECT ip, address, type, check_count, fail_count, last_time, last_status FROM %s ORDER BY type,check_count,last_time DESC", pdb.table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var proxies []*ProxyItem
	for rows.Next() {
		proxy := &ProxyItem{}
		err := rows.Scan(&proxy.IP, &proxy.Address, &proxy.Type, &proxy.CheckCount, &proxy.FailCount, &proxy.LastTime, &proxy.LastStatus)
		if err != nil {
			log.Printf("Error scanning proxy data: %s\n", err)
			continue
		}

		proxies = append(proxies, proxy)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return proxies, nil
}

func (pdb *ProxyDB) GetAllType(proxyType int) ([]*ProxyItem, error) {
	rows, err := pdb.db.Query(fmt.Sprintf("SELECT ip, address, type, check_count, fail_count, last_time, last_status FROM %s WHERE (type & ?) = ? ORDER BY type,check_count,last_time DESC", pdb.table), proxyType, proxyType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var proxies []*ProxyItem
	for rows.Next() {
		proxy := &ProxyItem{}
		err := rows.Scan(&proxy.IP, &proxy.Address, &proxy.Type, &proxy.CheckCount, &proxy.FailCount, &proxy.LastTime, &proxy.LastStatus)
		if err != nil {
			log.Printf("Error scanning proxy data: %s\n", err)
			continue
		}

		proxies = append(proxies, proxy)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return proxies, nil
}

func (pdb *ProxyDB) Pop() (*ProxyItem, error) {
	row := pdb.db.QueryRow(fmt.Sprintf("SELECT ip, address, type, check_count, fail_count, last_time, last_status FROM %s LIMIT 1", pdb.table))

	proxy := &ProxyItem{}
	err := row.Scan(&proxy.IP, &proxy.Address, &proxy.Type, &proxy.CheckCount, &proxy.FailCount, &proxy.LastTime, &proxy.LastStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	_, err = pdb.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE ip = ?", pdb.table), proxy.IP)
	if err != nil {
		return nil, err
	}

	return proxy, nil
}

func testProxyDB() {
	// 创建代理数据库
	db, err := NewProxyDB("proxies.db", "proxies")
	if err != nil {
		log.Fatalf("Failed to create proxy database: %v", err)
	}
	defer db.Close()

	// 添加代理项
	proxy1 := NewProxyItem("192.168.0.1", "http://proxy1.com", 1)
	err = db.Put(proxy1)
	if err != nil {
		log.Fatalf("Failed to add proxy1: %v", err)
	}

	proxy2 := NewProxyItem("192.168.0.2", "http://proxy2.com", 2)
	err = db.Put(proxy2)
	if err != nil {
		log.Fatalf("Failed to add proxy2: %v", err)
	}

	// 获取代理项
	proxy, err := db.Get()
	if err != nil {
		log.Fatalf("Failed to get proxy: %v", err)
	}
	if proxy != nil {
		fmt.Printf("Proxy: %+v\n", proxy)
	} else {
		fmt.Println("Proxy not found")
	}

	// 获取所有代理项
	proxies, err := db.GetAll()
	if err != nil {
		log.Fatalf("Failed to get all proxies: %v", err)
	}
	for _, p := range proxies {
		fmt.Printf("Proxy: %+v\n", p)
	}

	// 删除代理项
	err = db.Delete("192.168.0.1")
	if err != nil {
		log.Fatalf("Failed to delete proxy: %v", err)
	}
}
