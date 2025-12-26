package common

import "time"

// Message 定义消息结构体
type Message struct {
	Id          int        `json:"id" gorm:"primaryKey"`                // 消息的 ID, 用于排重
	MessageFrom int        `json:"message_from" gorm:"message_from"`    // 消息的发送方
	MessageTo   int        `json:"message_to" gorm:"message_to"`        // 消息接收方
	Content     string     `json:"content" gorm:"content"`              // 消息的内容
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at"` // 创建时间
	UpdatedAt   time.Time  `json:"updated_at" gorm:"column:updated_at"` // 更新时间
	DeletedAt   *time.Time `json:"deleted_at" gorm:"column:deleted_at"` // 逻辑删除时间
}

// 写入 Websocket 的 Request
type WsWriteRequest struct {
	MessageType int
	Data        []byte
	JSONPayload interface{}
	IsJSON      bool
}
