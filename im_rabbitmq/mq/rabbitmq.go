package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yzletter/go-chatery/im_rabbitmq/common"
	"github.com/yzletter/go-chatery/im_rabbitmq/mysql"
)

var mqOnce sync.Once

type RabbitMQService struct {
	mqConn    *amqp.Connection
	mqChannel *amqp.Channel
}

// NewRabbitMQService 构造函数
func NewRabbitMQService() *RabbitMQService {
	var mq *RabbitMQService

	// 确保只连接一次 RabbitMQ
	mqOnce.Do(
		func() {
			// 连接到 RabbitMQ
			conn, _ := amqp.Dial(fmt.Sprintf("amqp://%s:%s@localhost:5672", common.RabbitMQUser, common.RabbitMQPass))

			// 创建 Channel
			ch, _ := conn.Channel()

			mq = &RabbitMQService{
				mqConn:    conn,
				mqChannel: ch,
			}
		})

	return mq
}

// RegisterUser 注册用户的 exchange 和 queue
func (mq *RabbitMQService) RegisterUser(uid int) {
	// 声明 Exchange
	exchangeName := fmt.Sprintf("%d_exchange", uid)
	_ = mq.mqChannel.ExchangeDeclare(
		exchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)

	// 声明 Queue
	queueName := fmt.Sprintf("%d_queue", uid)
	_, _ = mq.mqChannel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)

	fmt.Println(exchangeName, queueName)

	// 将 Queue 绑定到 Exchange
	_ = mq.mqChannel.QueueBind(
		queueName,
		"",
		exchangeName,
		false,
		nil,
	)
}

// Produce 向 RabbitMQ 发送消息
func (mq *RabbitMQService) Produce(message *common.Message, uid int) error {
	// JSON 序列化
	msg, _ := json.Marshal(message)
	exchangeName := fmt.Sprintf("%d_exchange", uid)

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fmt.Printf("向 %s 写入 %s\n", exchangeName, msg)
	_ = mq.mqChannel.PublishWithContext(
		ctx,
		exchangeName,
		"",
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json", // MIME content type
			Body:         msg,
		},
	)

	return nil
}

// Consume 从 RabbitMQ 中获得读消息
func (mq *RabbitMQService) Consume(uid int, window chan common.Message) {
	queueName := fmt.Sprintf("%d_queue", uid)

	deliverCh, _ := mq.mqChannel.Consume(queueName, "", false, false, false, false, nil)
	go func() {
		for deliver := range deliverCh {
			var message common.Message
			_ = json.Unmarshal(deliver.Body, &message)

			mysql.DB.Create(&message)
			fmt.Printf("%d %s\n", uid, message)
			window <- message
			deliver.Ack(false)
		}
	}()
}

// 释放MQ连接
func (mq *RabbitMQService) Release() {
	if mq.mqChannel != nil {
		mq.mqChannel.Close()
	}
	if mq.mqConn != nil {
		mq.mqConn.Close()
	}
}
