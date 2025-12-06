package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Http 升级器
var upgrader = websocket.Upgrader{
	HandshakeTimeout: 1 * time.Second,
	ReadBufferSize:   100,
	WriteBufferSize:  100,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	engine := gin.Default()

	if err := engine.Run("localhost:8081"); err != nil {
		fmt.Println("服务启动失败")
		panic(err)
	}
}
