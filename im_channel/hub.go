package main

import (
	"fmt"
	"sync/atomic"
)

type Hub struct {
	broadcast  chan []byte          // 需要广播的消息
	servers    map[*Server]struct{} // 所有 Server 构成的集合
	register   chan *Server         // 需要注册的 Server
	unRegister chan *Server         // 需要注销的 Server
	state      int32                // 确保唯一
}

// NewHub 构造函数
func NewHub() *Hub {
	return &Hub{
		servers: make(map[*Server]struct{}),
		// 全部为无缓冲管道, 确保 Server 和 Hub 的交互是同步的, 即 Server 的请求 Hub 处理完之后, Server 才执行下一步工作
		broadcast:  make(chan []byte),
		register:   make(chan *Server),
		unRegister: make(chan *Server),
		state:      0,
	}
}

func (hub *Hub) Run() {
	// 当且仅当 hub.state 当前的值为 0 时，原子地将它设置为 1，并返回 true；否则不修改并返回 false
	if atomic.CompareAndSwapInt32(&hub.state, 0, 1) { // 确保只 Run 一次
		fmt.Printf("Hub 启动成功\n")
		for {
			select {
			case server := <-hub.register: // 有 Server 需要注册
				hub.servers[server] = struct{}{}
				fmt.Printf("IP : %s 注册成功\n", server.conn.RemoteAddr())
			case server := <-hub.unRegister: // 有 Server 需要注销
				if _, ok := hub.servers[server]; ok {
					delete(hub.servers, server) // 从集合中删除该 Server
					close(server.send)          // 关闭 Server 的管道
				}
			case msg := <-hub.broadcast: // 有消息需要广播
				for server := range hub.servers {
					server.send <- msg
				}
			}
		}
	}

}
