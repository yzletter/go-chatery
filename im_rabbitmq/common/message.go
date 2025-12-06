package common

// Message 定义消息结构体
type Message struct {
	Id       int64  // 消息的 ID, 用于排重
	Time     int64  // 精确到微妙, 以进入 Server 的时间为准
	From, To string // 消息的发送方和接收方, 带前缀 u 或 g, 表示是单聊或群聊
	Content  string // 消息的内容
}
