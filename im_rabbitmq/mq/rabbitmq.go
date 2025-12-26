package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yzletter/go-chatery/im_rabbitmq/common"
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

// Register 注册用户的 exchange 和 queue
func (mq *RabbitMQService) Register(uid int) {
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
func (mq *RabbitMQService) Produce(ctx context.Context, message *common.Message, uid int) error {
	// JSON 序列化
	msg, _ := json.Marshal(message)
	exchangeName := fmt.Sprintf("%d_exchange", uid)

	err := mq.mqChannel.PublishWithContext(
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

	return err
}

// Consume 从 RabbitMQ 中获得读消息
func (mq *RabbitMQService) Consume(ctx context.Context, uid int, send func(req common.WsWriteRequest) bool) error {
	// 队列名
	queueName := fmt.Sprintf("%d_queue", uid)

	// 消费队列
	deliverCh, err := mq.mqChannel.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case deliver, ok := <-deliverCh:
			if !ok {
				return nil
			}

			var message common.Message
			// 序列化失败
			if err := json.Unmarshal(deliver.Body, &message); err != nil {
				slog.Error("Unmarshal MQ message failed", "error", err)
				_ = deliver.Nack(false, false)
				continue
			}

			// 打入 wsChannel
			if !send(common.WsWriteRequest{IsJSON: true, JSONPayload: message}) {
				return ctx.Err()
			}
			if err := deliver.Ack(false); err != nil { // 打入 wsChannel 后才 ACK
				slog.Error("ACK MQ message failed", "error", err)
			}
		}
	}

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
