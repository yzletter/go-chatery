package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/yzletter/go-chatery/im_rabbitmq/common"
	"github.com/yzletter/go-chatery/im_rabbitmq/mq"
	"github.com/yzletter/go-chatery/im_rabbitmq/mysql"
)

// HTTP 升级器
var upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	ReadBufferSize:   10000,
	WriteBufferSize:  10000,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	pongWait   = 5 * time.Second //等待pong的超时时间
	pingPeriod = 3 * time.Second //发送ping的周期，必须短于pongWait
)

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
	handler.RabbitMQService.Register(uid)
}

// Speak 建立 WebSocket 连接, 并不断读取消息
func (handler *UserHandler) Speak(ctx *gin.Context) {
	uid, _ := strconv.Atoi(ctx.Param("id"))

	// 升级 Http
	wsConn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 创建可取消的子 Context
	connCtx, cancel := context.WithCancel(ctx)

	// defer 关闭连接
	defer func() {
		cancel()
		fmt.Println("close websocket connection")
		wsConn.Close()
	}()

	// 串行写 WS
	writeCh := make(chan common.WsWriteRequest, 30)
	send := func(message common.WsWriteRequest) bool {
		select {
		case <-connCtx.Done():
			return false
		case writeCh <- message:
			return true
		}
	}

	// 单独 Writer 协程写 Websocket
	go func() {
		defer cancel()
		for {
			select {
			case <-connCtx.Done():
				return
			case req, ok := <-writeCh:
				if !ok {
					return
				}
				var err error
				// 判断传入的 req
				if req.IsJSON { // 是 JSON
					err = wsConn.WriteJSON(req.JSONPayload)
				} else { // 是其他
					err = wsConn.WriteMessage(req.MessageType, req.Data)
				}
				if err != nil {
					fmt.Println("Websocket write failed", "error", err)
					return
				}
			}
		}
	}()

	// 心跳保持
	go heartBeat(connCtx, wsConn, send)

	// 子程：写数据到 Websocket 中
	go func() {
		if err := handler.RabbitMQService.Consume(ctx, uid, send); err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("Consume MQ Failed", "error", err)
		}
	}()

	// 主程：从 Websocket 中读数据打给 MQ
	for {
		_, body, err := wsConn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}

		// todo 判断消息类型，如 read_ack

		// 把 body 解析到 message 中
		var message common.Message
		_ = json.Unmarshal(body, &message)

		// 处理消息内容, 正常应进行对非法内容进行拦截。比如机器人消息（发言频率过快）；包含欺诈、涉政等违规内容；涉嫌私下联系/交易等。
		ok := intercept(message.Content)
		if !ok {
			continue // 略过此条消息不发给 RabbitMQ
		}

		// 落库
		message.Id = int(time.Now().UnixMicro())
		message.CreatedAt = time.Now()
		mysql.DB.Create(&message)

		// todo 更新会话信息

		// 将消息分别发给 MessageTo 和 MessageFrom
		_ = handler.RabbitMQService.Produce(ctx, &message, message.MessageTo)
		_ = handler.RabbitMQService.Produce(ctx, &message, message.MessageFrom)
	}

	cancel()
}

func heartBeat(ctx context.Context, conn *websocket.Conn, send func(common.WsWriteRequest) bool) {
	conn.SetPongHandler(func(appData string) error {
		return nil
	})

	// 启动 Ticker
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	if !send(common.WsWriteRequest{MessageType: websocket.PingMessage}) {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !send(common.WsWriteRequest{MessageType: websocket.PingMessage}) {
				return
			}
			deadline := time.Now().Add(pongWait) // ping发出去以后，期望5秒之内从conn里能计到数据（至少能读到pong）
			conn.SetReadDeadline(deadline)
			//fmt.Printf("must read before %s\n", deadline.Format("2006-01-02 15:04:05"))
		}
	}
}

// intercept 进行消息过滤
func intercept(content string) bool {
	if len(content) == 0 {
		return false
	}
	return true
}
