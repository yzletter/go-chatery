package common

// Message 定义消息结构体
type Message struct {
	Id      int    `json:"id"`   // 消息的 ID, 用于排重
	Time    int    `json:"time"` // 精确到微妙, 以进入 Server 的时间为准
	From    int    `json:"from"` // 消息的发送方和接收方, 带前缀 u 或 g, 表示是单聊或群聊
	To      int    `json:"to"`
	Content string `json:"content"` // 消息的内容
}
