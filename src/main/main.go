package main

import (
	"arpscan/src/conf"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/timest/gomanuf"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var log = logrus.New()

// ipNet 存放 IP地址和子网掩码
var ipNet *net.IPNet

// 本机的mac地址，发以太网包需要用到
var localHaddr net.HardwareAddr
var iface string

// 存放最终的数据，key[string] 存放的是IP地址
var data map[string]Info

// 计时器，在一段时间没有新的数据写入data中，退出程序，反之重置计时器
var t *time.Ticker
var do chan string

const (
	// 3秒的计时器
	START = "start"
	END   = "end"
)

type Info struct {
	// IP地址
	Mac net.HardwareAddr
	// 主机名
	Hostname string
	// 厂商信息
	Manuf string
}

// 格式化输出结果
// xxx.xxx.xxx.xxx  xx:xx:xx:xx:xx:xx  hostname  manuf
// xxx.xxx.xxx.xxx  xx:xx:xx:xx:xx:xx  hostname  manuf
func PrintData() {
	var keys IPSlice
	for k := range data {
		keys = append(keys, ParseIPString(k))
	}
	sort.Sort(keys)
	for _, k := range keys {
		d := data[k.String()]
		mac := ""
		if d.Mac != nil {
			mac = d.Mac.String()
		}
		fmt.Printf("%-15s %-17s %-30s %-10s\n", k.String(), mac, d.Hostname, d.Manuf)
	}
}

// 将抓到的数据集加入到data中，同时重置计时器
func pushData(ip string, mac net.HardwareAddr, hostname, manuf string) {
	// 停止计时器
	do <- START
	var mu sync.RWMutex
	mu.RLock()
	defer func() {
		// 重置计时器
		do <- END
		mu.RUnlock()
	}()
	if _, ok := data[ip]; !ok {
		data[ip] = Info{Mac: mac, Hostname: hostname, Manuf: manuf}
		return
	}
	info := data[ip]
	if len(hostname) > 0 && len(info.Hostname) == 0 {
		info.Hostname = hostname
	}
	if len(manuf) > 0 && len(info.Manuf) == 0 {
		info.Manuf = manuf
	}
	if mac != nil {
		info.Mac = mac
	}
	data[ip] = info
}

func setupNetInfo(f string) {
	var ifs []net.Interface
	var err error
	if f == "" {
		ifs, err = net.Interfaces()
	} else {
		// 已经选择iface
		var it *net.Interface
		it, err = net.InterfaceByName(f)
		if err == nil {
			ifs = append(ifs, *it)
		}
	}
	if err != nil {
		log.Fatal("无法获取本地网络信息:", err)
	}
	for _, it := range ifs {
		addr, _ := it.Addrs()
		for _, a := range addr {
			if ip, ok := a.(*net.IPNet); ok && !ip.IP.IsLoopback() {
				if ip.IP.To4() != nil {
					ipNet = ip
					localHaddr = it.HardwareAddr
					iface = it.Name
					goto END
				}
			}
		}
	}
END:
	if ipNet == nil || len(localHaddr) == 0 {
		log.Fatal("无法获取本地网络信息")
	}
}

func localHost() {
	host, _ := os.Hostname()
	data[ipNet.IP.String()] = Info{Mac: localHaddr, Hostname: strings.TrimSuffix(host, ".local"), Manuf: manuf.Search(localHaddr.String())}
}

func sendARP() {
	// ips 是内网IP地址集合
	ips := Table(ipNet)
	for _, ip := range ips {
		go sendArpPackage(ip)
	}
}

//查询mac_lists表中数据
func GetIp(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println("解析表单数据失败!")
	}
	ip := r.Form.Get("ip")
	js, err := json.Marshal(QueryIp(ip))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

//查询podname
func GetPodName(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageSize := 10
	err := r.ParseForm()
	if err != nil {
		fmt.Println("解析表单数据失败!")
	}
	podname := r.Form.Get("podname")
	namespace := r.Form.Get("namespace")
	ip := r.Form.Get("ip")
	status := r.Form.Get("status")
	pageStr := r.Form.Get("page")
	pageSizeStr := r.Form.Get("pageSize")
	fmt.Println("GetPodName 变量", podname, namespace, ip, status, pageStr, pageSizeStr)
	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}
	if pageSizeStr != "" {
		pageSize, _ = strconv.Atoi(pageSizeStr)
	}
	podinfoes := QuertPodName(namespace, podname, ip, status, page, pageSize)
	count := CountPodName(namespace, podname, ip, status)
	podPage := IpallcatorPage{podinfoes, count}
	js, err := json.Marshal(podPage)
	log.Info(podPage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func main() {
	conf.InitConfig(os.Args[1])
	//连接mysql 并且创建表
	ConnMsql()
	CreateTab()
	//allow non root user to execute by compare with euid
	if os.Geteuid() != 0 {
		log.Fatal("goscan must run as root.")
	}

	//设置读取网卡的接口
	flag.StringVar(&iface, "I", "", "Network interface name")
	flag.Parse()
	// 初始化 data
	data = make(map[string]Info)
	do = make(chan string)
	// 初始化 网络信息
	setupNetInfo(iface)

	go listenARP()
	//启动http服务
	http.HandleFunc("/getip", GetIp)
	http.HandleFunc("/getpodname", GetPodName)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../views/static"))))
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("../views/pages"))))
	http.ListenAndServe(fmt.Sprintf(":%d", conf.GetInst().ApiPort), nil)
}
