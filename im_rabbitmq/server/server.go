package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
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

	// 心跳保持
	go heartBeat(conn) //心跳保持

	uid, _ := strconv.Atoi(ctx.Query("id"))

	window := make(chan common.Message, 100)

	handler.RabbitMQService.Consume(uid, window)

	go handler.Receive(uid, conn, window)
	go handler.Send(conn)
}

func (handler *UserHandler) Receive(uid int, conn *websocket.Conn, window chan common.Message) {
	defer func() {
		err := conn.Close() // 错误忽略
		if err != nil {
			fmt.Println(err)
			return
		}
	}()

	// 拉取历史消息
	var msgs []common.Message
	mysql.DB.Model(&common.Message{}).Where("`from` = ?", uid).Or("`to` = ?", uid).Find(&msgs)

	// 按时间排序
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].Time > msgs[j].Time
	})

	if len(msgs) > 3 {
		msgs = msgs[:3]
	}

	Set := make(map[common.Message]struct{})

	for _, msg := range msgs {
		Set[msg] = struct{}{}
		printToFrontend(conn, msg)
	}

	// 从 window 中读取数据
	for {
		msg := <-window
		if _, exits := Set[msg]; !exits {
			Set[msg] = struct{}{}
			printToFrontend(conn, msg)
		}
	}
}

// Send 接受用户的消息, 打给 RabbitMQ
func (handler *UserHandler) Send(conn *websocket.Conn) {
	defer func() {
		err := conn.Close() // 错误忽略
		if err != nil {
			fmt.Println(err)
			return
		}
	}()
	for {
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

func printToFrontend(conn *websocket.Conn, msg common.Message) {
	_ = conn.WriteJSON(msg)
}
