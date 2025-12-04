package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {

	// 通过 shell 命令行指定端口
	port := flag.String("port", "5678", "http service port") // 默认为 5678 端口
	// 解析 shell 命令输入的参数
	flag.Parse()

	// 构造 Hub
	hub := NewHub()
	go hub.Run() // 启动 Hub

	http.HandleFunc("/", loadHomePage) // 首页路由
	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) { // 启动聊天
		StartServer(hub, w, r)
	})
	
	addr := "localhost:" + *port // 拼接出监听地址
	fmt.Printf("服务启动地址为 %s\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("服务启动失败 err : %s/n", err)
	}
}

// 加载 html 页面
func loadHomePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./im_channel/home.html")
}

// go run ./im_channel --port 5678
