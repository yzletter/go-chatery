package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/yzletter/go-chatery/im_rabbitmq/common"
	"github.com/yzletter/go-chatery/im_rabbitmq/mq"
)

// HTTP 升级器
var upgrader = websocket.Upgrader{
	HandshakeTimeout: 1 * time.Second,
	ReadBufferSize:   100,
	WriteBufferSize:  100,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type UserHandler struct {
	RabbitMQService *mq.RabbitMQService
}

func NewUserHandler(rabbitMQService *mq.RabbitMQService) *UserHandler {
	return &UserHandler{
		RabbitMQService: rabbitMQService,
	}
}

func (handler *UserHandler) Register(ctx *gin.Context) {
	uid, _ := strconv.Atoi(ctx.Query("id"))
	handler.RabbitMQService.RegisterUser(uid)
}

// Speak 建立 WebSocket 连接, 并不断读取消息
func (handler *UserHandler) Speak(ctx *gin.Context) {
	// 升级 Http
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		_ = conn.Close() // 错误忽略
	}()

	// 心跳保持
	go heartBeat(conn) //心跳保持

	uid, _ := strconv.Atoi(ctx.Query("id"))

	window := make(chan []byte, 100)

	handler.RabbitMQService.Consume(uid, window)

	go func(conn *websocket.Conn) {
		for {
			msg := <-window
			fmt.Printf("msg 为%s\n", msg)

			var message common.Message
			_ = json.Unmarshal(msg, &message)
			_ = conn.WriteJSON(message)
		}
	}(conn)

	for {
		// 接受用户的消息, 打给 RabbitMQ
		_, body, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}

		var msg common.Message
		_ = json.Unmarshal(body, &msg)                            // 把 body 解析到 msg 中
		msg.Content = strings.ReplaceAll(msg.Content, "\n", "  ") // 换行符用空格代替。将来要用换行符来分隔每条Message，所以一条Message内部不能出现换行符
		msg.Id = int(time.Now().UnixMicro())
		msg.Time = int(time.Now().UnixMicro())

		ok := intercept(msg.Content) // 处理消息内容, 正常应进行对非法内容进行拦截。比如机器人消息（发言频率过快）；包含欺诈、涉政等违规内容；涉嫌私下联系/交易等。
		if !ok {
			continue // 略过此条消息不发给 RabbitMQ
		}
		fmt.Println(msg)
		// 将消息分别发给 msg.To 和 msg.From
		_ = handler.RabbitMQService.Produce(&msg, msg.To)
		_ = handler.RabbitMQService.Produce(&msg, msg.From)
	}

}

// intercept 进行消息过滤
func intercept(content string) bool {
	if len(content) == 0 {
		return false
	}
	return true
}

var (
	pongWait   = 5 * time.Second //等待pong的超时时间
	pingPeriod = 3 * time.Second //发送ping的周期，必须短于pongWait
)

func heartBeat(conn *websocket.Conn) {
	conn.SetPongHandler(func(appData string) error {
		return nil
	})

	err := conn.WriteMessage(websocket.PingMessage, nil)
	if err != nil {
		conn.WriteMessage(websocket.CloseMessage, nil)
	}

	ticker := time.NewTicker(pingPeriod)
LOOP:
	for {
		<-ticker.C
		err := conn.WriteMessage(websocket.PingMessage, nil)
		if err != nil {
			conn.WriteMessage(websocket.CloseMessage, nil)
			break LOOP
		}
		deadline := time.Now().Add(pongWait) // ping发出去以后，期望5秒之内从conn里能计到数据（至少能读到pong）
		conn.SetReadDeadline(deadline)
	}
}
