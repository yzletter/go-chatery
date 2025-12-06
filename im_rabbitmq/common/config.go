package common

// 定义用户和群组前缀
const (
	TypeUser  = "u"
	TypeGroup = "g"
)

// RabbitMQ 账号密码
const (
	RabbitMQUser = "yzletter"
	RabbitMQPass = "123456"
)

// 本地文件存放位置
const (
	GroupMemberPath = "im_rabbitmq/data/server/group"
	ReceiveUserPath = "im_rabbitmq/data/client/user"
)
