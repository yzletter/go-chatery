package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-chatery/im_rabbitmq/mq"
	"github.com/yzletter/go-chatery/im_rabbitmq/server"
)

func main() {
	engine := gin.Default()
	RabbitMQSvc := mq.NewRabbitMQService()
	UserHdl := server.NewUserHandler(RabbitMQSvc)

	engine.GET("/register", UserHdl.Register)
	engine.GET("/speak", UserHdl.Speak)
	if err := engine.Run("localhost:8081"); err != nil {
		fmt.Println("服务启动失败")
		panic(err)
	}
}
