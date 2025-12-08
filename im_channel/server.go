package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	newLine = []byte{'\n'}
	space   = []byte{' '}
)

type Server struct {
	send     chan []byte // 消息管道
	userName []byte      // 用户名
	// 绑定 Hub 和 Conn
	hub  *Hub
	conn *websocket.Conn
}

// http 升级器
var upgrader = websocket.Upgrader{
	HandshakeTimeout: 1 * time.Second,
	ReadBufferSize:   100,
	WriteBufferSize:  100,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func StartServer(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil) // 将 Http 升级为 WebSocket协议
	if err != nil {
		fmt.Printf("升级 http 协议失败, err : %s\n", err)
		return
	}
	fmt.Printf("成功连接 %s\n", conn.RemoteAddr().String())

	// 新建一个 Server
	server := &Server{
		send: make(chan []byte, 256),
		hub:  hub,
		conn: conn,
	}

	// 设置读写超时
	err1 := server.conn.SetWriteDeadline(time.Now().Add(24 * time.Hour))
	err2 := server.conn.SetReadDeadline(time.Now().Add(24 * time.Hour))
	if err1 != nil || err2 != nil {
		fmt.Printf("设置 Server 读写超时失败 \n")
		return
	}

	// 向 Hub 中注册当前 Server
	server.hub.register <- server

	// 开启协程
	go server.Read()
	go server.Write()
}

// 从 WebSocket 中读取数据, 写到 Hub 中去, 由 Hub 广播给其他 Server
func (server *Server) Read() {
	// 收尾工作
	defer func() {
		server.hub.unRegister <- server // 注销当前 Server
		fmt.Printf("用户 %s 下线\n", server.userName)
		fmt.Printf("关闭连接 %s\n", server.conn.RemoteAddr().String())
		err := server.conn.Close() //关闭websocket管道
		if err != nil {
			return
		}
	}()

	for {
		// 如果前端主动断开连接，该行会报错，for循环会退出。注销server时，hub那儿会关闭server.send管道
		_, content, err := server.conn.ReadMessage()
		if err != nil {
			break
		}

		// 对消息进行简单的处理, 换行符用空格替代，bytes.TrimSpace把首尾连续的空格去掉
		content = bytes.TrimSpace(bytes.Replace(content, newLine, space, -1))
		if len(server.userName) <= 0 {
			// 约定第一条消息是当前用户名, 把上线消息发给其他用户
			server.userName = content
			server.hub.broadcast <- []byte(fmt.Sprintf("%s 已上线", string(server.userName)))
			fmt.Println("写入 hub 成功")
		} else {
			// 把消息前拼上用户名
			server.hub.broadcast <- bytes.Join([][]byte{server.userName, content}, []byte(":"))
			fmt.Println("写入 hub 成功")
		}
	}
}

// 读取 Hub 传来的其他 Server 发来的数据, 通过 WebSocket 返回给前端
func (server *Server) Write() {
	// 收尾工作
	defer func() {
		// 给前端写数据失败, 说明前端已经关了, 直接关闭连接
		fmt.Printf("关闭连接 %s\n", server.conn.RemoteAddr().String())
		err := server.conn.Close()
		if err != nil {
			return
		}
	}()

	for {
		msg, ok := <-server.send // ok 判断管道是否还能读
		if !ok {
			// 管道已经不能读
			err := server.conn.WriteMessage(websocket.CloseMessage, []byte("bye bye"))
			if err != nil {
				return
			}
			return
		} else {
			// 写给前端
			err := server.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				fmt.Printf("向浏览器发送数据失败:%v\n", err)
				return
			}
		}
	}
}
