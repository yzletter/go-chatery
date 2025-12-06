package main

import (
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {

	// 通过 shell 命令行指定端口
	port := flag.String("port", "5678", "http service port") // 默认为 5678 端口
	// 解析 shell 命令输入的参数
	flag.Parse()

	// 构造 Hub
	hub := NewHub()
	go hub.Run() // 启动 Hub

	engine := gin.Default()

	engine.GET("/", loadHomePage)                // 首页路由
	engine.GET("/chat", func(ctx *gin.Context) { // 启动聊天
		StartServer(hub, ctx.Writer, ctx.Request)
	})

	addr := "localhost:" + *port // 拼接出监听地址
	fmt.Printf("服务启动地址为 %s\n", addr)

	if err := engine.Run(addr); err != nil {
		fmt.Printf("服务启动失败 err : %s/n", err)
	}
}

// 加载 html 页面
func loadHomePage(ctx *gin.Context) {
	ctx.File("./im_channel/home.html")
}

// go run ./im_channel --port 5678
